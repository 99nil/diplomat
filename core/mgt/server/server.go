// Copyright © 2022 99nil.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"net/http"
	"time"

	"github.com/99nil/diplomat/pkg/sse"

	v1 "github.com/99nil/diplomat/pkg/api/v1"

	"github.com/99nil/diplomat/pkg/k8s/watchsched"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/diplomat/pkg/nodeset"
	"github.com/99nil/diplomat/pkg/types"
	"github.com/99nil/diplomat/pkg/util"
	"github.com/99nil/dsync"
	badgerstorage "github.com/99nil/dsync/storage/badger"
	"github.com/99nil/dsync/suid"
	"github.com/99nil/gopkg/server"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func Run(cfg *Config, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	storageClient, err := badgerstorage.New(cfg.Storage.Badger)
	if err != nil {
		return err
	}

	ins, err := dsync.New(
		dsync.WithStorageOption(storageClient))
	if err != nil {
		return err
	}
	set := nodeset.New()

	s := server.New(&cfg.Server)
	s.Handler = NewRouter(cfg, kubeClient, ins, set)
	s.WriteTimeout = 0
	s.ReadTimeout = 0

	wg, ctx := errgroup.WithContext(context.Background())
	wg.Go(func() error {
		return s.ListenAndServe()
	})
	wg.Go(func() error {
		return s.ShutdownGraceful(ctx)
	})
	wg.Go(func() error {
		eventHandlerFuncs := NewEventHandlerFuncs(ctx, ins, set)
		sched := watchsched.New(kubeClient, dynamicClient, eventHandlerFuncs)
		return sched.Run(ctx)
	})
	wg.Go(func() error {
		return DatasetGC(ctx, ins)
	})
	return wg.Wait()
}

func NewRouter(cfg *Config, kubeClient kubernetes.Interface, ins dsync.Interface, set nodeset.Interface) http.Handler {
	mux := chi.NewMux()
	mux.Use(
		middleware.Recoverer,
		middleware.Logger,
	)

	mux.Route("/api/v1", func(r chi.Router) {
		r.Get("/manifest", manifest(cfg, kubeClient, ins, set))
		r.Get("/data", sse.Wrap(data(ins)))
	})
	return mux
}

// NewEventHandlerFuncs returns a resource event handler funcs
func NewEventHandlerFuncs(ctx context.Context, ins dsync.Interface, set nodeset.Interface) cache.ResourceEventHandlerFuncs {
	eventHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { eventOperate(ctx, ins, set, watch.Added, obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { eventOperate(ctx, ins, set, watch.Modified, newObj) },
		DeleteFunc: func(obj interface{}) { eventOperate(ctx, ins, set, watch.Deleted, obj) },
	}
	return eventHandlerFuncs
}

// eventOperate defines the resource event handler
func eventOperate(
	ctx context.Context,
	ins dsync.Interface,
	set nodeset.Interface,
	eventType watch.EventType,
	obj interface{},
) {
	object, ok := obj.(*unstructured.Unstructured)
	if !ok {
		logr.Warnf("Parse as *unstructured.Unstructured failed, ignore. Type: %T", obj)
		return
	}
	objectBytes, err := object.MarshalJSON()
	if err != nil {
		logr.WithError(err).Warnf("Marshal *unstructured.Unstructured failed, ignore")
		return
	}

	gvk := object.GroupVersionKind()
	namespace := object.GetNamespace()
	metaKey := types.NewMeta(gvk.Group, gvk.Version, gvk.Kind, namespace, object.GetName(), object.GetResourceVersion())

	uid := suid.NewByCustom(metaKey.String())
	event := v1.Event{
		Type: eventType,
		Data: objectBytes,
	}
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		logr.WithError(err).Error("Marshal JSON from event failed")
		return
	}
	if err := ins.DataSet().Add(ctx, dsync.Item{
		UID:   uid,
		Value: jsonBytes,
	}); err != nil {
		logr.WithError(err).WithField("uid", uid.CustomUID()).Errorf("Add uid to dateset failed")
		return
	}

	plural, _ := meta.UnsafeGuessKindToResource(gvk)
	nodes := set.Get(nodeset.Key{
		Group:     metaKey.Group,
		Resource:  plural.Resource,
		Namespace: namespace,
	})
	rootNodes := set.Get(nodeset.Any)
	allNodes := nodes.Union(rootNodes)
	for k := range allNodes {
		if err := ins.Syncer(k).Add(ctx, uid); err != nil {
			logr.WithError(err).WithFields(map[string]interface{}{
				"uid":  uid.CustomUID(),
				"node": k,
			}).Error("Add uid to node manifest failed")
		}
	}
}

// DatasetGC runs the dataset garbage collection
func DatasetGC(ctx context.Context, ins dsync.Interface) error {
	records := make(map[string]string)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Minute * 30):
		}

		err := ins.DataSet().RangeCustom(ctx, func(uid suid.UID) error {
			key := uid.CustomUID()
			metaKey, err := types.ParseMetaStr(key)
			if err != nil {
				logr.WithError(err).WithField("key", key).Error("DataSet GC, parse meta key failed")
				// Need to continue garbage collection, so don't exit.
				return nil
			}
			rv, ok := records[metaKey.NameString()]
			if !ok {
				records[metaKey.NameString()] = metaKey.ResourceVersion
				return nil
			}
			if util.CompareResourceVersion(rv, metaKey.ResourceVersion) < 0 {
				records[metaKey.NameString()] = metaKey.ResourceVersion
				return nil
			}

			if err := ins.DataSet().Del(ctx, uid); err != nil {
				logr.WithError(err).WithField("key", key).Error("DataSet GC, delete data failed")
				// Need to continue garbage collection, so don't exit.
			}
			return nil
		})
		if err != nil {
			logr.WithError(err).Error("DataSet GC, Range failed")
		}
	}
}

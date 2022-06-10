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
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/99nil/diplomat/pkg/nodeset"

	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/99nil/diplomat/pkg/k8s/watchsched"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/diplomat/pkg/types"
	"github.com/99nil/dsync"
	"github.com/99nil/dsync/suid"
	"github.com/99nil/gopkg/server"
	"github.com/go-chi/chi"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func Run(cfg *Config, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	ins := dsync.New()
	set := nodeset.New()

	s := server.New(&cfg.Server)
	s.Handler = NewRouter(kubeClient, ins, set)
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
	return wg.Wait()
}

func NewRouter(kubeClient kubernetes.Interface, ins dsync.Interface, set nodeset.Interface) http.Handler {
	mux := chi.NewMux()
	mux.Route("/api/v1", func(r chi.Router) {
		r.Get("/manifest", manifest(kubeClient, ins, set))
		r.Post("/data", data(ins))
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

	gvk := object.GroupVersionKind()
	namespace := object.GetNamespace()
	metaKey := types.NewMeta(gvk.Group, gvk.Version, gvk.Kind, namespace, object.GetName(), object.GetResourceVersion())

	uid := suid.NewByCustom(metaKey.String())
	event := watch.Event{
		Type:   eventType,
		Object: object,
	}
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		logr.Errorf("Marshal JSON from event failed: %v", err)
		return
	}
	if err := ins.DataSet().Add(ctx, dsync.Item{
		UID:   uid,
		Value: jsonBytes,
	}); err != nil {
		logr.Errorf("Add uid(%s) to dateset failed: %v", uid.CustomUID(), err)
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
			logr.Errorf("Add uid(%s) to node(%s) manifest failed: %v", uid.CustomUID(), k, err)
		}
	}
}

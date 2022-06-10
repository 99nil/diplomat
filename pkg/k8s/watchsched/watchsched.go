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

package watchsched

import (
	"context"
	"time"

	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/gopkg/sets"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Interface interface {
	Run(ctx context.Context) error
}

type Engine struct {
	kubeClient        kubernetes.Interface
	dynamicClient     dynamic.Interface
	informerFactory   dynamicinformer.DynamicSharedInformerFactory
	eventHandlerFuncs cache.ResourceEventHandlerFuncs
	set               sets.String
}

func New(
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
	eventHandlerFuncs cache.ResourceEventHandlerFuncs,
) Interface {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0)
	return &Engine{
		kubeClient:        kubeClient,
		informerFactory:   informerFactory,
		set:               sets.NewString(),
		eventHandlerFuncs: eventHandlerFuncs,
		dynamicClient:     dynamicClient,
	}
}

func (e *Engine) Run(ctx context.Context) error {
	e.run(ctx)

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			e.run(ctx)
		}
	}
}

func (e *Engine) run(ctx context.Context) {
	gvrSet, err := e.resourceSchedule()
	if err != nil {
		logr.Errorf("resource schedule failed: %v", err)
	}
	for gvr := range gvrSet {
		e.informerFactory.ForResource(gvr).Informer().AddEventHandler(e.eventHandlerFuncs)
	}
	e.informerFactory.Start(ctx.Done())
}

func (e *Engine) resourceSchedule() (map[schema.GroupVersionResource]struct{}, error) {
	resources, err := e.kubeClient.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	noSet := make(map[schema.GroupVersionResource]struct{})
	for _, v := range resources {
		gv, err := schema.ParseGroupVersion(v.GroupVersion)
		if err != nil {
			logr.Debugf("Skip resource schedule, GroupVersion(%s) parse failed: %v", v.GroupVersion, err)
			continue
		}
		for _, vv := range v.APIResources {
			if UnWatchResourceSet.Has(vv.Kind) {
				continue
			}
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: vv.Name,
			}
			key := gvr.String()
			if !e.set.Has(key) {
				noSet[gvr] = struct{}{}
				e.set.Add(key)
			}
		}
	}
	return noSet, nil
}

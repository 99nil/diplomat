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

	"github.com/99nil/diplomat/pkg/k8s/watchsched"
	"github.com/99nil/dsync"
	"github.com/99nil/gopkg/server"
	"github.com/go-chi/chi"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func Run(cfg *Config, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) error {
	ins := dsync.New()

	s := server.New(&cfg.Server)
	s.Handler = NewRouter(ins)
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
		// TODO add to dsync
		eventHandlerFuncs := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {

			},
			UpdateFunc: func(oldObj, newObj interface{}) {

			},
			DeleteFunc: func(obj interface{}) {

			},
		}
		sched := watchsched.New(kubeClient, dynamicClient, eventHandlerFuncs)
		return sched.Run(ctx)
	})
	return wg.Wait()
}

func NewRouter(ins dsync.Interface) http.Handler {
	mux := chi.NewMux()
	mux.Route("/api/v1", func(r chi.Router) {
		r.Get("/manifest", manifest(ins))
		r.Post("/data", data(ins))
	})
	return mux
}

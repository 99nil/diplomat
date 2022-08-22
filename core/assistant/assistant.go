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

package assistant

import (
	"context"
	"time"

	"github.com/99nil/diplomat/pkg/logr"

	"golang.org/x/sync/errgroup"

	"github.com/99nil/diplomat/core/plugin"
	"github.com/docker/docker/client"
)

func Run(cfg *Config, dockerClient *client.Client) error {
	ctx := context.Background()
	set := []plugin.Interface{
		plugin.NewDragonfly2(cfg.Part),
	}

	engineSet := make(map[string]plugin.Engine)
	for _, v := range set {
		var engine plugin.Engine
		switch v.Kind() {
		case "docker":
			engine = plugin.NewDockerEngine(dockerClient, v.Options())
		case "process":
			engine = plugin.NewProcessEngine(v.Options())
		default:
			continue
		}
		if err := engine.Start(ctx); err != nil {
			return err
		}
		engineSet[v.Name()] = engine
	}

	// TODO 支持插件的升级及重启
	// TODO 支持自升级
	// TODO 监听插件的健康状态
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second * 60):
		}

		eg, ctx := errgroup.WithContext(ctx)
		for name, e := range engineSet {
			eg.Go(func() error {
				stats, err := e.Stats(ctx)
				if err != nil {
					logr.WithField("name", e.Name()).
						WithField("error", err).
						Errorf("stats failed")
					return err
				}

				for _, stat := range stats {
					logr.WithField("name", name).
						WithField("module", stat.Name).
						Debugf("stats completed")
				}
				return nil
			})
		}
		_ = eg.Wait()
	}
}

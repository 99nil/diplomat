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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/99nil/diplomat/pkg/sse"

	v1 "github.com/99nil/diplomat/pkg/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/99nil/diplomat/pkg/health"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/dsync"
	badgerstorage "github.com/99nil/dsync/storage/badger"

	"golang.org/x/sync/errgroup"
)

func Run(cfg *Config) error {
	storageClient, err := badgerstorage.New(cfg.Storage.Badger)
	if err != nil {
		return err
	}

	ins, err := dsync.New(
		dsync.WithStorageOption(storageClient))
	if err != nil {
		return err
	}

	healthIns := health.New()

	wg, ctx := errgroup.WithContext(context.Background())
	wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			err := SyncData(ctx, cfg, ins, healthIns)
			if err == dsync.ErrDataNotMatch {
				continue
			}
			if err == nil {
				// When the synchronization and the cloud are consistent,
				// wait for a period of time to initiate the request again.
				logr.Debugf("Sync data finished, wait 10s to continue")
				// TODO wait time to be configurable
				time.Sleep(time.Second * 10)
				continue
			}
			logr.WithError(err).Error("Sync data failed")
		}
	})
	wg.Go(func() error {
		return StartAPIServer(ctx)
	})
	return wg.Wait()
}

func SyncData(
	ctx context.Context,
	cfg *Config,
	ins dsync.Interface,
	healthIns health.Interface,
) error {
	state := ins.DataSet().State(ctx)

	client := NewClient(cfg.Server.Host, cfg.Agent.Name)
	manifest, err := client.Manifest(ctx, state)
	if err != nil {
		return fmt.Errorf("request manifest failed: %v", err)
	}
	if manifest == nil {
		return nil
	}

	ins.DataSet().SyncManifest(ctx, manifest)
	return client.Data(ctx, manifest, func(msg *sse.Message) error {
		var items []dsync.Item
		if err := json.Unmarshal([]byte(msg.Data), &items); err != nil {
			return fmt.Errorf("unmarshal items failed: %v", err)
		}

		err := ins.DataSet().SyncAndDelete(ctx, items, func(ctx context.Context, item dsync.Item) error {
			var event v1.Event
			if err := json.Unmarshal(item.Value, &event); err != nil {
				return fmt.Errorf("unmarshal event failed: %v", err)
			}

			var object unstructured.Unstructured
			if err := object.UnmarshalJSON(event.Data); err != nil {
				return fmt.Errorf("unmarshal runtime.Object failed: %v", err)
			}
			logr.WithFields(map[string]interface{}{
				"gvk":             object.GroupVersionKind().String(),
				"namespace":       object.GetNamespace(),
				"name":            object.GetName(),
				"resourceVersion": object.GetResourceVersion(),
			}).Debugf("SyncAndDelete item")
			// TODO use lite-apiServer storage api to send item
			// TODO e.g. save and watch event
			return nil
		})

		if err == dsync.ErrDataNotMatch {
			logr.WithError(err).Debug("Sync and delete data item stopped")
		}
		return err
	})
}

// StartAPIServer
// TODO APIServer（轻量化 k8s APIServer），供节点直接调用，获取本地存储中的资源，增/改/删 操作需要透传至云端
func StartAPIServer(ctx context.Context) error {
	return nil
}

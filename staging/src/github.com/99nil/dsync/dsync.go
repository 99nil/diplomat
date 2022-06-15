// Copyright © 2021 zc2638 <zc2638@qq.com>.
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

package dsync

import (
	"context"
	"strings"

	"github.com/99nil/dsync/suid"
)

const prefix = "dsync"

const keyState = "dsync_state"

var (
	spaceStatePrefix   = buildName(prefix, "state")
	spaceDatasetPrefix = buildName(prefix, "dataset")
	spaceSyncerPrefix  = buildName(prefix, "syncer")
	spaceRelatePrefix  = buildName(prefix, "relate")
	spaceTmpPrefix     = buildName(prefix, "tmp")
)

// Item defines the data item
type Item struct {
	UID   suid.UID
	Value []byte
}

// Interface defines dsync core
type Interface interface {
	// DataSet returns a data set
	DataSet() DataSet

	// Syncer returns a synchronizer with a specified name
	Syncer(name string) Synchronizer

	// Clear clears all data
	Clear(ctx context.Context) error
}

// Synchronizer defines the synchronizer operations
type Synchronizer interface {
	// Add adds UIDs to sync set
	Add(ctx context.Context, uids ...suid.UID) error

	// Del deletes UIDs from sync set
	Del(ctx context.Context, uids ...suid.UID) error

	// Manifest gets a manifest that needs to be synchronized according to the UID
	Manifest(ctx context.Context, uid suid.UID, limit int) (*suid.AssembleManifest, error)

	// Data gets the data items to be synchronized according to the manifest
	Data(ctx context.Context, manifest *suid.AssembleManifest) ([]Item, error)
}

// DataSet defines the data set operations
type DataSet interface {
	// SetState sets the latest state of the dataset
	SetState(ctx context.Context, uid suid.UID) error

	// State gets the latest state of the dataset
	State(ctx context.Context) suid.UID

	// Get gets data according to UID
	Get(ctx context.Context, uid suid.UID) (*Item, error)

	// Add adds data items
	Add(ctx context.Context, items ...Item) error

	// Del deletes data according to UIDs
	Del(ctx context.Context, uids ...suid.UID) error

	// Range calls fn sequentially for item present in the dataset.
	// If fn returns error, range stops the iteration.
	Range(ctx context.Context, fn func(item *Item) error) error

	// RangeCustom calls fn sequentially for UID present in the dataset.
	// If fn returns error, range stops the iteration.
	RangeCustom(ctx context.Context, fn func(uid suid.UID) error) error

	// SyncManifest syncs the manifest that needs to be executed
	SyncManifest(ctx context.Context, manifest *suid.AssembleManifest)

	// Sync syncs the data according to manifest and items
	Sync(ctx context.Context, items []Item, callback ItemCallbackFunc) error

	// SyncAndDelete syncs and deletes the data according to manifest and items
	SyncAndDelete(ctx context.Context, items []Item, callback ItemCallbackFunc) error
}

type ItemCallbackFunc func(context.Context, Item) error

func buildName(ss ...string) string {
	nameSet := make([]string, 0, len(ss))
	for _, s := range ss {
		current := strings.TrimSpace(s)
		if current == "" {
			continue
		}
		nameSet = append(nameSet, current)
	}
	return strings.Join(nameSet, "_")
}

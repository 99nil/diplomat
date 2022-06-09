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
	"errors"
	"sync"

	"github.com/99nil/dsync/storage"
	"github.com/99nil/dsync/suid"
)

var (
	ErrDataNotMatch  = errors.New("data not match")
	ErrUnexpectState = errors.New("unexpect state")
)

type dataSet struct {
	mux      sync.Mutex
	manifest *suid.AssembleManifest
	state    suid.UID

	defaultOperation OperateInterface
	dataSetOperation OperateInterface
	tmpOperation     OperateInterface
	customOperation  OperateInterface
}

func newDataSet(insName string, storage storage.Interface) *dataSet {
	ds := new(dataSet)
	ds.defaultOperation = newSpaceOperation(buildName(spaceStatePrefix, insName), storage)
	ds.dataSetOperation = newSpaceOperation(buildName(spaceDatasetPrefix, insName), storage)
	ds.tmpOperation = newSpaceOperation(buildName(spaceTmpPrefix, insName), storage)
	ds.customOperation = newSpaceOperation(buildName(spaceRelatePrefix, insName), storage)
	return ds
}

func (ds *dataSet) completeUID(ctx context.Context, uids ...suid.UID) ([]suid.UID, error) {
	set := make([]suid.UID, 0, len(uids))
	for _, uid := range uids {
		if uid.IsCustom() {
			custom := uid.CustomUID()
			value, err := ds.customOperation.Get(ctx, custom)
			if err != nil {
				return nil, err
			}
			id, err := suid.ParseKSUID(string(value))
			if err != nil {
				return nil, err
			}
			uid = suid.NewWithCustom(id, custom)
		}
		set = append(set, uid)
	}
	return set, nil
}

func (ds *dataSet) SetState(ctx context.Context, uid suid.UID) error {
	if err := ds.defaultOperation.Add(ctx, keyState, []byte(uid.String())); err != nil {
		return err
	}
	ds.state = uid
	return nil
}

func (ds *dataSet) State(ctx context.Context) suid.UID {
	if !ds.state.IsNil() {
		return ds.state
	}

	value, err := ds.defaultOperation.Get(ctx, keyState)
	if err != nil {
		return ds.state
	}
	ds.state = value
	return ds.state
}

func (ds *dataSet) Get(ctx context.Context, uid suid.UID) (*Item, error) {
	uids, err := ds.completeUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	uid = uids[0]

	value, err := ds.dataSetOperation.Get(ctx, uid.KSUID().String())
	if err != nil {
		return nil, err
	}
	return &Item{
		UID:   uid,
		Value: value,
	}, nil
}

func (ds *dataSet) Add(ctx context.Context, items ...Item) error {
	if len(items) == 0 {
		return nil
	}

	state := ds.State(ctx)
	current := state.KSUID()
	for _, item := range items {
		itemCurrent := item.UID.KSUID()
		isCustom := item.UID.IsCustom()
		if isCustom && itemCurrent.IsNil() {
			itemCurrent = suid.NewKSUID()
		}
		// When adding data in batches, the order may not be guaranteed,
		// so perform the addition first, and then determine the latest state.
		if err := ds.dataSetOperation.Add(ctx, itemCurrent.String(), item.Value); err != nil {
			return err
		}
		if item.UID.IsCustom() {
			// Add data first, if the association relationship fails,
			// the orphaned data will be recovered by the GC soon
			if err := ds.customOperation.Add(ctx, item.UID.CustomUID(), []byte(itemCurrent.String())); err != nil {
				return err
			}
		}

		if suid.CompareKSUID(current, itemCurrent) > -1 {
			continue
		}
		if err := ds.SetState(ctx, item.UID); err != nil {
			return err
		}
	}
	return nil
}

func (ds *dataSet) Del(ctx context.Context, uids ...suid.UID) error {
	if len(uids) == 0 {
		return nil
	}
	uids, err := ds.completeUID(ctx, uids...)
	if err != nil {
		return err
	}

	for _, uid := range uids {
		if err := ds.dataSetOperation.Del(ctx, uid.KSUID().String()); err != nil {
			return err
		}
	}
	return nil
}

func (ds *dataSet) SyncManifest(ctx context.Context, manifest *suid.AssembleManifest) {
	ds.mux.Lock()
	defer ds.mux.Unlock()

	state := ds.State(ctx)
	if !state.IsNil() {
		var (
			set    []suid.UID
			exists bool
		)
		current := state.KSUID()
		for iter := manifest.Iter(); iter.Next(); {
			if iter.KSUID == current {
				exists = true
			}
			if exists {
				set = append(set, manifest.GetUID(iter.KSUID))
			}
		}
		manifest = suid.NewManifest()
		manifest.AppendUID(set...)
	}
	ds.manifest = manifest
}

func (ds *dataSet) Sync(ctx context.Context, items []Item, callback ItemCallbackFunc) error {
	return ds.sync(ctx, items, callback, false)
}

func (ds *dataSet) SyncAndDelete(ctx context.Context, items []Item, callback ItemCallbackFunc) error {
	return ds.sync(ctx, items, callback, true)
}

func (ds *dataSet) sync(ctx context.Context, items []Item, callback ItemCallbackFunc, needDelete bool) error {
	if len(items) == 0 {
		return nil
	}
	state := ds.State(ctx)
	current := state.KSUID()

	var count int
	// Store items in tmp space first,
	// and use items in subsequent synchronization to prevent the sequence of data from affecting synchronization.
	for _, item := range items {
		isExpire := suid.CompareKSUID(current, item.UID.KSUID())
		if isExpire > -1 {
			continue
		}
		if err := ds.tmpOperation.Add(ctx, item.UID.KSUID().String(), item.Value); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return nil
	}

	// In order to ensure the consistency of the manifest,
	// it needs to be locked before this.
	ds.mux.Lock()
	defer ds.mux.Unlock()

	var match bool
	for iter := ds.manifest.Iter(); iter.Next(); {
		uid := iter.KSUID
		uidStr := uid.String()
		state := ds.State(ctx)

		// When state is Nil, directly synchronize data
		if state.IsNil() {
			match = true
		}
		if !match {
			// When the corresponding state is found in the manifest,
			// start synchronization from the next.
			if state.KSUID() == uid {
				match = true
			}
			continue
		}

		value, err := ds.tmpOperation.Get(ctx, uidStr)
		if err != nil {
			return err
		}
		if value == nil {
			return ErrDataNotMatch
		}

		item := Item{
			UID:   ds.manifest.GetUID(uid),
			Value: value,
		}
		if err := ds.Add(ctx, item); err != nil {
			return err
		}
		if callback != nil {
			if err := callback(ctx, item); err != nil {
				return err
			}
		}

		// Try to delete the synchronized data in the tmp space.
		// If the deletion fails, no verification is required,
		// and it can be reclaimed by the GC later.
		_ = ds.tmpOperation.Del(ctx, uidStr)
		if needDelete {
			if err := ds.dataSetOperation.Del(ctx, uidStr); err != nil {
				return err
			}
		}
	}
	return nil
}

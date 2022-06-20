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

	"github.com/99nil/dsync/storage"
	"github.com/99nil/dsync/suid"
)

var (
	ErrEmptyManifest = errors.New("empty manifest")
)

type syncer struct {
	name             string
	syncerOperation  OperateInterface
	dataSetOperation OperateInterface
	customOperation  OperateInterface
}

func newSyncer(insName string, name string, storage storage.Interface) *syncer {
	s := &syncer{name: name}
	s.syncerOperation = newSpaceOperation(buildName(spaceSyncerPrefix, insName), storage)
	s.dataSetOperation = newSpaceOperation(buildName(spaceDatasetPrefix, insName), storage)
	s.customOperation = newSpaceOperation(buildName(spaceRelatePrefix, insName), storage)
	return s
}

func (s *syncer) filterUIDs(ctx context.Context, uids []suid.UID) ([]suid.UID, error) {
	set := make([]suid.UID, 0, len(uids))
	for _, uid := range uids {
		if uid.IsCustom() {
			custom := uid.CustomUID()
			value, err := s.customOperation.Get(ctx, custom)
			if err != nil {
				return nil, err
			}
			id, err := suid.ParseKSUID(string(value))
			if err != nil {
				return nil, err
			}
			if id.IsNil() {
				continue
			}
			uid = suid.NewWithCustom(id, custom)
		}
		set = append(set, uid)
	}
	return set, nil
}

func (s *syncer) setManifest(ctx context.Context, manifest *suid.AssembleManifest) error {
	var (
		b   []byte
		err error
	)
	if manifest != nil {
		b, err = manifest.Bytes()
		if err != nil {
			return err
		}
	}
	return s.syncerOperation.Add(ctx, s.name, b)
}

func (s *syncer) getManifest(ctx context.Context) (*suid.AssembleManifest, error) {
	value, err := s.syncerOperation.Get(ctx, s.name)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return suid.NewManifest(), ErrEmptyManifest
	}
	return suid.NewManifestFromBytes(value)
}

func (s *syncer) Add(ctx context.Context, uids ...suid.UID) error {
	if len(uids) == 0 {
		return nil
	}
	uids, err := s.filterUIDs(ctx, uids)
	if err != nil {
		return err
	}

	manifest, err := s.getManifest(ctx)
	if err != nil && err != ErrEmptyManifest {
		return err
	}
	manifest.AppendUID(uids...)
	return s.setManifest(ctx, manifest)
}

func (s *syncer) Del(ctx context.Context, uids ...suid.UID) error {
	if len(uids) == 0 {
		return nil
	}
	manifest, err := s.getManifest(ctx)
	if err != nil && err != ErrEmptyManifest {
		return err
	}

	var set []suid.UID
	for iter := manifest.Iter(); iter.Next(); {
		var current suid.UID
		for _, uid := range uids {
			if uid.KSUID() == iter.KSUID {
				current = uid[:]
				break
			}
		}
		if len(current) > 0 {
			set = append(set, current)
		}
	}

	manifest = suid.NewManifest()
	manifest.AppendUID(set...)
	return s.setManifest(ctx, manifest)
}

func (s *syncer) Manifest(ctx context.Context, uid suid.UID, limit int) (*suid.AssembleManifest, error) {
	manifest, err := s.getManifest(ctx)
	if err != nil {
		return manifest, err
	}

	result := suid.NewManifest()
	number := 0

	current := uid.KSUID()
	if current == suid.Nil {
		for iter := manifest.Iter(); iter.Next(); {
			keyUID := manifest.GetUID(iter.KSUID)
			if limit > 0 && number >= limit {
				break
			}
			result.AppendUID(keyUID)
			number++
		}
		return result, nil
	}

	var (
		set    []suid.UID
		exists bool
	)
	manifest.AppendUID(uid)
	for iter := manifest.Iter(); iter.Next(); {
		if iter.KSUID == current {
			exists = true
		}
		if !exists {
			continue
		}

		keyUID := manifest.GetUID(iter.KSUID)
		if limit <= 0 || number < limit {
			result.AppendUID(keyUID)
			number++
		}
		set = append(set, keyUID)
	}

	// If there is none or only uid itself, there is nothing to synchronize.
	if len(set) < 2 {
		_ = s.setManifest(ctx, nil)
		return nil, ErrEmptyManifest
	}
	manifest = suid.NewManifest()
	manifest.AppendUID(set...)
	manifest.Sort()
	if err := s.setManifest(ctx, manifest); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *syncer) Data(ctx context.Context, manifest *suid.AssembleManifest) ([]Item, error) {
	var items []Item
	for iter := manifest.Iter(); iter.Next(); {
		value, err := s.dataSetOperation.Get(ctx, iter.KSUID.String())
		if err != nil {
			return nil, err
		}
		items = append(items, Item{
			UID:   manifest.GetUID(iter.KSUID),
			Value: value,
		})
	}
	return items, nil
}

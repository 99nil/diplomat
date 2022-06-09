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

package badger

import (
	"errors"

	"github.com/dgraph-io/badger/v3"

	"github.com/99nil/dsync/storage"
)

type ClientIterator struct {
	iter  *badger.Iterator
	err   error
	valid bool
}

func NewClientIterator(iter *badger.Iterator) *ClientIterator {
	ci := &ClientIterator{iter: iter}
	if iter == nil {
		ci.err = errors.New("iterator unavailable")
	}
	return ci
}

func (i *ClientIterator) Next() bool {
	if !i.valid {
		i.valid = true
	} else {
		i.iter.Next()
	}
	return i.iter.Valid()
}

func (i *ClientIterator) Error() error {
	return nil
}

func (i *ClientIterator) Value() *storage.KV {
	item := i.iter.Item()
	kv := new(storage.KV)

	var err error
	if kv.Value, err = item.ValueCopy(kv.Value); err != nil {
		i.err = err
		return nil
	}
	item.KeyCopy(kv.Key)
	return kv
}

func (i *ClientIterator) Close() error {
	i.iter.Close()
	return nil
}

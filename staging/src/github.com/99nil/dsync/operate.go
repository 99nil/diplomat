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
	"encoding/json"

	"github.com/99nil/dsync/storage"
)

type OperateInterface interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Add(ctx context.Context, key string, value []byte) error
	Del(ctx context.Context, key string) error
	Range(ctx context.Context, fn func(key, value []byte) error) error

	AddData(ctx context.Context, key string, data interface{}) error
}

type spaceOperation struct {
	name    string
	storage storage.Interface
}

func newSpaceOperation(name string, storage storage.Interface) OperateInterface {
	return &spaceOperation{name: name, storage: storage}
}

func (o *spaceOperation) Get(ctx context.Context, key string) ([]byte, error) {
	return o.storage.Get(ctx, o.name, key)
}

func (o *spaceOperation) Add(ctx context.Context, key string, value []byte) error {
	return o.storage.Add(ctx, o.name, key, value)
}

func (o *spaceOperation) Del(ctx context.Context, key string) error {
	return o.storage.Del(ctx, o.name, key)
}

func (o *spaceOperation) Range(ctx context.Context, fn func(key, value []byte) error) error {
	return o.storage.Range(ctx, o.name, fn)
}

func (o *spaceOperation) AddData(ctx context.Context, key string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return o.storage.Add(ctx, o.name, key, b)
}

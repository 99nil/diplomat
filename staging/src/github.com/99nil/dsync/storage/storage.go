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

package storage

import "context"

// Interface defines storage related interface
type Interface interface {
	// Get gets data according to the specified key in current space
	Get(ctx context.Context, space, key string) ([]byte, error)

	// Add adds a set of key/value pairs in current space
	Add(ctx context.Context, space, key string, value []byte) error

	// Del deletes key/value pairs according to the specified key in current space
	Del(ctx context.Context, space, key string) error

	// Clear clears all data in the specified space
	Clear(ctx context.Context, space string) error

	// Range calls fn sequentially for each key and value present in the storage.
	// If fn returns error, range stops the iteration.
	Range(ctx context.Context, space string, fn func(key, value []byte) error) error
}

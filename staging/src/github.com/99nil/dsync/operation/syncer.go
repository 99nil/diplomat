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

package operation

import (
	"context"

	"github.com/99nil/dsync/suid"
)

// dataset init时，在一个 init_dataset 空间中初始化数据，中间所有的add/del操作都保留
// init 完成后，覆盖 dataset 空间

// dataset add/del时，调用 syncer 的 add/del 操作
// 新的syncer创建时，调用 syncer 的 init 操作

type SyncerOperation interface {
	Init() error
	Add(ctx context.Context, uids ...suid.UID) error
	Del(ctx context.Context, uids ...suid.UID) error
}

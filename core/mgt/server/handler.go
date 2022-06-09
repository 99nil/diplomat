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
	"encoding/json"
	"net/http"

	"github.com/99nil/dsync/suid"

	"github.com/99nil/gopkg/ctr"

	"github.com/99nil/dsync"
)

func manifest(ins dsync.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeName := r.Header.Get("node")
		state := r.Header.Get("state")
		m, err := ins.Syncer(nodeName).Manifest(r.Context(), []byte(state))
		if err != nil {
			ctr.InternalError(w, err)
			return
		}
		ctr.OK(w, m)
	}
}

func data(ins dsync.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeName := r.Header.Get("node")
		var m suid.AssembleManifest
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			ctr.BadRequest(w, err)
			return
		}

		// TODO 通过 SSE 获取资源接口
		// TODO 需要限制返回的 items 数量，循环判断是否获取完毕，一次性返回过多容易导致崩溃
		items, err := ins.Syncer(nodeName).Data(r.Context(), &m)
		if err != nil {
			ctr.InternalError(w, err)
			return
		}
		ctr.OK(w, items)
	}
}

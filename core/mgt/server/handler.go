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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/diplomat/pkg/nodeset"
	"github.com/99nil/dsync"
	"github.com/99nil/dsync/suid"
	"github.com/99nil/gopkg/ctr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func manifest(
	cfg *Config,
	kubeClient kubernetes.Interface,
	ins dsync.Interface,
	set nodeset.Interface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeName := r.Header.Get("node")
		state := r.Header.Get("state")
		ctx := r.Context()

		syncer := ins.Syncer(nodeName)
		if !set.Has(nodeName) {
			node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			if err != nil {
				ctr.InternalError(w, fmt.Errorf("get node(%s) failed: %v", nodeName, err))
				return
			}

			var keys []nodeset.Key
			// Resolve the ClusterRole associated with the node
			clusterRoleStr, clusterRoleOK := node.Annotations[constants.AnnotationRelateClusterRole]
			clusterRoles := strings.Split(clusterRoleStr, ",")
			for _, roleName := range clusterRoles {
				roleName = strings.TrimSpace(roleName)
				if roleName == "" {
					continue
				}
				clusterRole, err := kubeClient.RbacV1().ClusterRoles().Get(ctx, roleName, metav1.GetOptions{})
				if err != nil {
					ctr.InternalError(w, err)
					return
				}
				for _, rule := range clusterRole.Rules {
					for _, group := range rule.APIGroups {
						for _, resource := range rule.Resources {
							keys = append(keys, nodeset.Key{
								Group:    group,
								Resource: resource,
							})
						}
					}
				}
			}

			// Resolve the Role associated with the node
			// namespace1: role1,rol2,rol3; namespace2: role1,role2
			roleStr, roleOK := node.Annotations[constants.AnnotationRelateRole]
			parts := strings.Split(roleStr, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				arr := strings.Split(part, ":")
				if len(arr) != 2 {
					logr.Warnf("node(%s) relate role parse failed: format error, ignore.")
					continue
				}

				namespace := strings.TrimSpace(arr[0])
				roleName := strings.TrimSpace(arr[1])
				if roleName == "" {
					continue
				}

				role, err := kubeClient.RbacV1().Roles(namespace).Get(ctx, roleName, metav1.GetOptions{})
				if err != nil {
					ctr.InternalError(w, err)
					return
				}
				for _, rule := range role.Rules {
					for _, group := range rule.APIGroups {
						for _, resource := range rule.Resources {
							keys = append(keys, nodeset.Key{
								Group:    group,
								Resource: resource,
							})
						}
					}
				}
			}

			if !clusterRoleOK && !roleOK {
				keys = []nodeset.Key{nodeset.Any}
			}
			set.Set(nodeName, keys)

			// Get all matching UIDs and add them to the node manifest
			var uids []suid.UID
			err = ins.DataSet().RangeCustom(ctx, func(uid suid.UID) error {
				uids = append(uids, uid)
				return nil
			})
			if err != nil {
				ctr.InternalError(w, err)
				return
			}
			if err := syncer.Add(ctx, uids...); err != nil {
				ctr.InternalError(w, err)
				return
			}
		}

		m, err := syncer.Manifest(ctx, []byte(state), 100)
		if err != nil {
			ctr.InternalError(w, err)
			return
		}

		// Proxy Server forwards requests according to the specified instance.
		// So the agent must carry back this request header.
		w.Header().Set(fmt.Sprintf("%s-mgt-server-instance", constants.ProjectName), cfg.Instance.Name)
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
		logr.Debugf("events: stream started, node: %s", nodeName)

		// 需要限制返回的 items 数量，循环判断是否获取完毕，一次性返回过多容易导致崩溃
		num := 0
		current := suid.NewManifest()
		var ms []suid.AssembleManifest
		for iter := m.Iter(); iter.Next(); num++ {
			current.Append(iter.KSUID)
			// TODO number limit can be config
			if num > 10 {
				num = 0
				ms = append(ms, *current)
				current = suid.NewManifest()
			}
		}

		// Send data over SSE
		h := w.Header()
		h.Set("Content-Type", "text/event-stream")
		h.Set("Cache-Control", "no-cache")
		h.Set("Connection", "keep-alive")
		h.Set("X-Accel-Buffering", "no")

		f, ok := w.(http.Flusher)
		if !ok {
			return
		}
		_, _ = w.Write([]byte(": ping\n\n"))
		f.Flush()

		for _, v := range ms {
			items, err := ins.Syncer(nodeName).Data(r.Context(), &v)
			if err != nil {
				errStr := fmt.Sprintf("events: get sync data failed, node: %s, error: %s", nodeName, err)
				logr.Error(errStr)
				_, _ = io.WriteString(w, fmt.Sprintf("event: error\ndata: %s\n\n", errStr))
				return
			}
			b, err := json.Marshal(items)
			if err != nil {
				errStr := fmt.Sprintf("events: marshal sync data failed, node: %s, error: %s", nodeName, err)
				logr.Error(errStr)
				_, _ = io.WriteString(w, fmt.Sprintf("event: error\ndata: %s\n\n", errStr))
				return
			}
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(b)
			_, _ = w.Write([]byte("\n\n"))
			f.Flush()
		}
		_, _ = w.Write([]byte("event: error\ndata: eof\n\n"))
		f.Flush()
		logr.Debugf("events: stream closed, node: %s", nodeName)
	}
}

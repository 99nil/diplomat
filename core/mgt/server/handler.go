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

func manifest(kubeClient kubernetes.Interface, ins dsync.Interface, set nodeset.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeName := r.Header.Get("node")
		state := r.Header.Get("state")
		ctx := r.Context()

		if !set.Has(nodeName) {
			node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			if err != nil {
				ctr.InternalError(w, fmt.Errorf("get node(%s) failed: %v", nodeName, err))
				return
			}

			var keys []nodeset.Key
			// 解析 node 关联的 clusterrole
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

			// 解析 node 关联的 role
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
			// TODO 获取所有符合的 UID 添加到该节点 manifest
			// TODO dsync dateset 添加迭代方法，把所有关联关系获取到
		}

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

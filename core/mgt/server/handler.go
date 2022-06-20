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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/99nil/diplomat/pkg/sse"

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
		ctx := r.Context()
		state := r.Header.Get("state")
		nodeName := r.Header.Get("node")
		if nodeName == "" {
			ctr.BadRequest(w, errors.New("node name not found"))
			return
		}

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
		if err != nil && err != dsync.ErrEmptyManifest {
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
		if nodeName == "" {
			sse.NewErrMessage("", "node name not found").Send(w)
			return
		}
		logr.Debugf("events: stream started, node: %s", nodeName)

		manifestStr := r.Header.Get("manifest")
		if strings.TrimSpace(manifestStr) == "" {
			sse.NewErrMessage("", dsync.ErrEmptyManifest.Error()).Send(w)
			return
		}

		var m suid.AssembleManifest
		if err := json.Unmarshal([]byte(manifestStr), &m); err != nil {
			sse.NewErrMessage("", fmt.Sprintf("unmarshal manifest failed: %v", err)).Send(w)
			return
		}

		// It is necessary to limit the number of items returned,
		// and loop to determine whether the acquisition is complete.
		// Too many returns at one time can easily lead to crashes.
		num := 0
		current := suid.NewManifest()
		var ms []suid.AssembleManifest
		for iter := m.Iter(); iter.Next(); num++ {
			current.Append(iter.KSUID)
			// TODO number limit to be configurable
			if num > 10 {
				num = 0
				ms = append(ms, *current)
				current = suid.NewManifest()
			}
		}
		if num > 0 {
			ms = append(ms, *current)
		}

		for _, v := range ms {
			items, err := ins.Syncer(nodeName).Data(r.Context(), &v)
			if err != nil {
				errStr := fmt.Sprintf("get sync data failed, node: %s, error: %s", nodeName, err)
				logr.Error(errStr)
				sse.NewErrMessage("", errStr).Send(w)
				return
			}

			b, err := json.Marshal(items)
			if err != nil {
				return
			}
			sse.NewMessage("", "", string(b)).Send(w)
		}
		logr.Debugf("events: stream closed, node: %s", nodeName)
	}
}

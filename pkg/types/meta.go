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

package types

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	utilstrings "k8s.io/utils/strings"
)

type MetaKey struct {
	schema.GroupVersionKind

	Namespace       string
	Name            string
	ResourceVersion string
}

// e.g. (apps, v1, Deployment, default, test, 0) => apps/v1,Deployment,default/test,0
func (m MetaKey) String() string {
	apiVersion, kind := m.ToAPIVersionAndKind()
	qualifiedName := utilstrings.JoinQualifiedName(m.Namespace, m.Name)
	return strings.Join([]string{apiVersion, kind, qualifiedName, m.ResourceVersion}, ",")
}

func (m *MetaKey) QualifiedName() string {
	return utilstrings.JoinQualifiedName(m.Namespace, m.Name)
}

func NewMeta(group, version, kind, namespace, name, resourceVersion string) *MetaKey {
	return &MetaKey{
		GroupVersionKind: schema.GroupVersionKind{
			Group:   group,
			Version: version,
			Kind:    kind,
		},
		Namespace:       namespace,
		Name:            name,
		ResourceVersion: resourceVersion,
	}
}

func ParseMetaStr(s string) (*MetaKey, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, errors.New("meta string is not valid")
	}
	gvk := schema.FromAPIVersionAndKind(parts[0], parts[1])
	namespace, name := utilstrings.SplitQualifiedName(parts[2])
	resourceVersion := parts[3]
	return &MetaKey{
		GroupVersionKind: gvk,
		Namespace:        namespace,
		Name:             name,
		ResourceVersion:  resourceVersion,
	}, nil
}

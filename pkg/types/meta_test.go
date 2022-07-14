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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		group           string
		version         string
		kind            string
		namespace       string
		name            string
		resourceVersion string
	}

	tests := []struct {
		name string
		args args
		want *MetaKey
	}{
		{
			name: "default",
			args: args{},
			want: &MetaKey{},
		},
		{
			name: "set value",
			args: args{
				group:           "apps",
				version:         "v1",
				kind:            "deployments",
				namespace:       "default",
				name:            "test",
				resourceVersion: "0",
			},
			want: &MetaKey{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "deployments",
				},
				Namespace:       "default",
				Name:            "test",
				ResourceVersion: "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.args
			if got := NewMeta(m.group, m.version, m.kind, m.namespace, m.name, m.resourceVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMeta(...) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMetaStr(t *testing.T) {
	type fields struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		want   *MetaKey
	}{
		{
			name: "",
			fields: fields{
				s: "apps/v1,deployment,default/test,0",
			},
			want: &MetaKey{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "deployment",
				},
				Namespace:       "default",
				Name:            "test",
				ResourceVersion: "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseMetaStr(tt.fields.s)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMetaStr(tt.fields.s) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMetaStrError(t *testing.T) {
	type fields struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		want   *MetaKey
	}{
		{
			name: "",
			fields: fields{
				s: "apps/v1",
			},
			want: &MetaKey{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "deployment",
				},
				Namespace:       "default",
				Name:            "test",
				ResourceVersion: "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMetaStr(tt.fields.s)
			message := "meta string is not valid"
			if err.Error() != message {
				t.Errorf("wrong error message. want %q. got %q.", message, err.Error())
			}
		})
	}
}

func Test_MetaKey_String(t *testing.T) {
	type fields struct {
		group           string
		version         string
		kind            string
		namespace       string
		name            string
		resourceVersion string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pods",
			fields: fields{
				group:           "apps",
				version:         "v1",
				kind:            "deployment",
				namespace:       "default",
				name:            "test",
				resourceVersion: "0",
			},
			want: "apps/v1,deployment,default/test,0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MetaKey{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   tt.fields.group,
					Version: tt.fields.version,
					Kind:    tt.fields.kind,
				},
				Namespace:       tt.fields.namespace,
				Name:            tt.fields.name,
				ResourceVersion: tt.fields.resourceVersion,
			}
			if got := m.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_MetaKey_NameString(t *testing.T) {
	type fields struct {
		group           string
		version         string
		kind            string
		namespace       string
		name            string
		resourceVersion string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pods",
			fields: fields{
				group:     "apps",
				version:   "v1",
				kind:      "deployment",
				namespace: "default",
				name:      "test",
			},
			want: "apps/v1,deployment,default/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MetaKey{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   tt.fields.group,
					Version: tt.fields.version,
					Kind:    tt.fields.kind,
				},
				Namespace:       tt.fields.namespace,
				Name:            tt.fields.name,
				ResourceVersion: tt.fields.resourceVersion,
			}
			if got := m.NameString(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_MetaKey_QualifiedName(t *testing.T) {
	type fields struct {
		group           string
		version         string
		kind            string
		namespace       string
		name            string
		resourceVersion string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pods",
			fields: fields{
				namespace: "default",
				name:      "test",
			},
			want: "default/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MetaKey{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   tt.fields.group,
					Version: tt.fields.version,
					Kind:    tt.fields.kind,
				},
				Namespace:       tt.fields.namespace,
				Name:            tt.fields.name,
				ResourceVersion: tt.fields.resourceVersion,
			}
			if got := m.QualifiedName(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

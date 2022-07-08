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

package nodeset

import (
	"reflect"
	"testing"

	"github.com/99nil/gopkg/sets"
)

func TestKey_String(t *testing.T) {
	type fields struct {
		Group     string
		Resource  string
		Namespace string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pods",
			fields: fields{
				Group:     "v1",
				Resource:  "pods",
				Namespace: "default",
			},
			want: "v1/pods/default",
		},
		{
			name: "deployments",
			fields: fields{
				Group:     "apps/v1",
				Resource:  "deployments",
				Namespace: "default",
			},
			want: "apps/v1/deployments/default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := Key{
				Group:     tt.fields.Group,
				Resource:  tt.fields.Resource,
				Namespace: tt.fields.Namespace,
			}
			if got := k.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want Interface
	}{
		{
			name: "default",
			want: &set{
				data: make(map[string]sets.String),
				all:  sets.NewString(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_set_Get(t *testing.T) {
	defaultKey := Key{
		Group:     "v1",
		Resource:  "pods",
		Namespace: "default",
	}

	type fields struct {
		data map[string]sets.String
		all  sets.String
	}
	type args struct {
		key Key
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   sets.String
	}{
		{
			name: "default",
			fields: fields{
				data: map[string]sets.String{
					defaultKey.String(): {
						"test": struct{}{},
					},
				},
				all: sets.String{
					"test": struct{}{},
				},
			},
			args: args{key: defaultKey},
			want: sets.String{
				"test": struct{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &set{
				data: tt.fields.data,
				all:  tt.fields.all,
			}
			if got := s.Get(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_set_Has(t *testing.T) {
	defaultKey := Key{
		Group:     "v1",
		Resource:  "pods",
		Namespace: "default",
	}

	type fields struct {
		data map[string]sets.String
		all  sets.String
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "exists",
			fields: fields{
				data: map[string]sets.String{
					defaultKey.String(): {
						"test": struct{}{},
					},
				},
				all: sets.String{
					"test": struct{}{},
				},
			},
			args: args{name: "test"},
			want: true,
		},
		{
			name: "not exist",
			fields: fields{
				data: map[string]sets.String{
					defaultKey.String(): {
						"test": struct{}{},
					},
				},
				all: sets.String{
					"test": struct{}{},
				},
			},
			args: args{name: "mock"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &set{
				data: tt.fields.data,
				all:  tt.fields.all,
			}
			if got := s.Has(tt.args.name); got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_set_Set(t *testing.T) {
	defaultKey := Key{
		Group:     "v1",
		Resource:  "pods",
		Namespace: "default",
	}

	type fields struct {
		data map[string]sets.String
		all  sets.String
	}
	type args struct {
		name string
		keys []Key
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "default",
			fields: fields{
				data: map[string]sets.String{
					defaultKey.String(): {
						"test": struct{}{},
					},
				},
				all: sets.String{
					"test": struct{}{},
				},
			},
			args: args{
				name: "test",
				keys: []Key{defaultKey},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &set{
				data: tt.fields.data,
				all:  tt.fields.all,
			}
			s.Set(tt.args.name, tt.args.keys)
		})
	}
}

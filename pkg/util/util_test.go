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

package util

import "testing"

func TestCompareResourceVersion(t *testing.T) {
	type fields struct {
		after  string
		before string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "a equal b",
			fields: struct {
				after  string
				before string
			}{
				after:  "1",
				before: "1",
			},
			want: 0,
		},
		{
			name: "a less than b",
			fields: struct {
				after  string
				before string
			}{
				after:  "1",
				before: "2",
			},
			want: -1,
		},
		{
			name: "a greater than b",
			fields: struct {
				after  string
				before string
			}{
				after:  "2",
				before: "1",
			},
			want: +1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.fields
			if got := CompareResourceVersion(v.after, v.before); got != tt.want {
				t.Errorf("CompareResourceVersion(v.after, v.before) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseResourceVersion(t *testing.T) {
	type fields struct {
		ResourceVersion string
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{
			name: "parse",
			fields: fields{
				ResourceVersion: "1",
			},
			want: 1,
		},
		{
			name: "default 0",
			fields: fields{
				ResourceVersion: "0",
			},
			want: 0,
		},
		{
			name: `default ""`,
			fields: fields{
				ResourceVersion: "",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseResourceVersion(tt.fields.ResourceVersion); got != tt.want {
				t.Errorf("ParseResourceVersion(tt.args) = %v, want %v", got, tt.want)
			}
		})
	}
}

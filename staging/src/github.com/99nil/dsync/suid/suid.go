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

package suid

import (
	"bytes"
)

var Separator = []byte(".")

type RelateMap map[string]KSUID

func (rm RelateMap) Reverse() map[KSUID]string {
	result := make(map[KSUID]string, len(rm))
	for k, v := range rm {
		result[v] = k
	}
	return result
}

type UID []byte

func (u UID) String() string {
	return string(u)
}

func (u UID) Len() int {
	return len(u)
}

func (u UID) IsNil() bool {
	return u.KSUID().IsNil()
}

func (u UID) IsCustom() bool {
	return len(u) > 27
}

func (u UID) KSUID() KSUID {
	arr := bytes.SplitN(u[:], Separator, 2)
	id, _ := ParseKSUID(string(arr[0]))
	return id
}

func (u UID) CustomUID() string {
	arr := bytes.SplitN(u[:], Separator, 2)
	if len(arr) != 2 {
		return ""
	}
	return string(arr[1])
}

func New() UID {
	return []byte(NewKSUID().String())
}

func NewByCustom(id string) UID {
	return NewWithCustom(Nil, id)
}

func NewWithCustom(ksuid KSUID, id string) UID {
	uid := []byte(ksuid.String())
	if id == "" {
		return uid
	}
	uid = append(uid, Separator...)
	uid = append(uid[:], []byte(id)...)
	return uid
}

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
	"encoding/json"

	"github.com/segmentio/ksuid"
)

func NewManifest() *AssembleManifest {
	return &AssembleManifest{}
}

func NewManifestFromBytes(b []byte) (*AssembleManifest, error) {
	manifest := NewManifest()
	if err := manifest.FromBytes(b); err != nil {
		return nil, err
	}
	return manifest, nil
}

var sep = []byte("...")

// AssembleManifest defines a list of UIDs to be synchronized
type AssembleManifest struct {
	cs  ksuid.CompressedSet
	set map[KSUID]string
}

func (am *AssembleManifest) getIds() []KSUID {
	var set []KSUID
	for iter := am.cs.Iter(); iter.Next(); {
		set = append(set, iter.KSUID)
	}
	return set
}

func (am *AssembleManifest) Clone() *AssembleManifest {
	manifest := &AssembleManifest{}
	manifest.cs = am.cs[:]
	if am.set != nil {
		manifest.set = make(map[KSUID]string)
		for k, v := range am.set {
			manifest.set[k] = v
		}
	}
	return manifest
}

func (am *AssembleManifest) Iter() CompressedSetIter {
	set := am.getIds()
	for id := range am.set {
		set = append(set, id)
	}
	manifest := ksuid.Compress(set...)
	return manifest.Iter()
}

func (am *AssembleManifest) Sort() {
	set := am.getIds()
	am.cs = ksuid.Compress(set...)
}

func (am *AssembleManifest) Append(ids ...KSUID) {
	am.cs = ksuid.AppendCompressed(am.cs, ids...)
}

func (am *AssembleManifest) AppendCustom(id KSUID, custom string) {
	if am.set == nil {
		am.set = make(map[KSUID]string)
	}
	am.set[id] = custom
}

func (am *AssembleManifest) AppendUID(uids ...UID) {
	for _, uid := range uids {
		if uid.IsCustom() {
			am.AppendCustom(uid.KSUID(), uid.CustomUID())
		} else {
			am.Append(uid.KSUID())
		}
	}
}

func (am *AssembleManifest) GetUID(id KSUID) UID {
	return NewWithCustom(id, am.set[id])
}

func (am *AssembleManifest) Bytes() ([]byte, error) {
	if len(am.set) == 0 {
		return am.cs, nil
	}

	b, err := json.Marshal(am.set)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(am.cs)+len(b)+1)
	out = append(out, am.cs...)
	out = append(out, sep...)
	out = append(out, b...)
	return out, nil
}

func (am *AssembleManifest) FromBytes(b []byte) error {
	bytesSet := bytes.SplitN(b, sep, 2)
	if len(bytesSet[0]) > 0 {
		am.cs = bytesSet[0]
	}
	if len(bytesSet) == 2 {
		return json.Unmarshal(bytesSet[1], &am.set)
	}
	return nil
}

func (am *AssembleManifest) MarshalText() ([]byte, error) {
	b, err := am.Bytes()
	if err != nil {
		return nil, err
	}
	return json.Marshal(b)
}

func (am *AssembleManifest) UnmarshalText(data []byte) error {
	var b []byte
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	return am.FromBytes(b)
}

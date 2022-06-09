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
	"github.com/segmentio/ksuid"
)

// Nil represents a completely empty (invalid) UID
var Nil KSUID

// KSUID references KSUID implementation.
// KSUID is for K-Sortable Unique Identifier.
// It is a kind of globally unique identifier similar to a RFC 4122 UUID,
// built from the ground-up to be "naturally" sorted by generation timestamp without any special type-aware logic.
// In short, running a set of KSUIDs through the UNIX sort command will result in a list ordered by generation time.
type KSUID = ksuid.KSUID

// CompressedSetIter references CompressedSetIter implementation.
// CompressedSetIter is an iterator type returned by Set.Iter to produce the
// list of KSUIDs stored in a set.
type CompressedSetIter = ksuid.CompressedSetIter

// NewKSUID generates a new KSUID. In the strange case that random bytes
// can't be read, it will panic.
func NewKSUID() KSUID {
	return ksuid.New()
}

// CompareKSUID references Compare implementation.
func CompareKSUID(a KSUID, b KSUID) int {
	return ksuid.Compare(a, b)
}

// ParseKSUID references Parse implementation.
func ParseKSUID(s string) (KSUID, error) {
	return ksuid.Parse(s)
}

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
	"path"
	"sync"

	"github.com/99nil/gopkg/sets"
)

type Interface interface {
	Has(name string) bool
	Get(key Key) sets.String
	Set(name string, keys []Key)
}

func New() Interface {
	return &set{
		// group/resource/namespace => nodes
		data: make(map[string]sets.String),
		all:  sets.NewString(),
	}
}

var Any = Key{Group: "*", Resource: "*", Namespace: "*"}

type Key struct {
	Group     string
	Resource  string
	Namespace string
}

func (k Key) String() string {
	return path.Join(k.Group, k.Resource, k.Namespace)
}

type set struct {
	sync.Mutex
	data map[string]sets.String
	all  sets.String
}

func (s *set) Has(name string) bool {
	return s.all.Has(name)
}

func (s *set) Get(key Key) sets.String {
	s.Lock()
	defer s.Unlock()
	return s.data[key.String()]
}

func (s *set) Set(name string, keys []Key) {
	s.Lock()
	defer s.Unlock()

	s.all.Add(name)
	for _, key := range keys {
		keyPath := key.String()
		if _, ok := s.data[keyPath]; !ok {
			s.data[keyPath] = sets.NewString()
		}
		s.data[keyPath].Add(name)
	}
}

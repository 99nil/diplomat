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

package plugin

func NewDragonfly2(part string) Interface {
	return &dragonfly2{part: part}
}

type dragonfly2 struct {
	part string
}

func (p *dragonfly2) Name() string {
	return "dragonfly2"
}

func (p *dragonfly2) Kind() string {
	return "docker"
}

func (p *dragonfly2) Options() []Option {
	opts := []Option{
		{
			Name:  "dragonfly2-dfdaemon",
			Image: "dragonflyoss/dfdaemon:v2.0.5",
		},
	}
	if p.part == "cloud" {
		opts = append(opts, Option{
			Name:  "dragonfly2-scheduler",
			Image: "dragonflyoss/scheduler:v2.0.5",
		}, Option{
			Name:  "dragonfly2-manager",
			Image: "dragonflyoss/manager:v2.0.5",
		})
	}
	return opts
}

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

package assistant

import (
	"github.com/99nil/diplomat/pkg/k8s"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	PartCloud = "cloud"
	PartEdge  = "edge"
)

func Environ(envPrefix string) *Config {
	cfg := &Config{}
	cfg.Part = PartCloud

	viper.MustBindEnv("PART")
	return cfg
}

type Config struct {
	Kubernetes *k8s.Config `json:"kubernetes,omitempty"`
	Logger     logr.Config `json:"logger,omitempty"`
	Part       string      `json:"part"`
	// TODO 优化，避免多个组件yaml导致添加更多的字段
	KubeEdgeResources [][]unstructured.Unstructured
	RavenResources    [][]unstructured.Unstructured
}

func (c *Config) Complete() {
}

func (c *Config) Validate() error {
	return nil
}

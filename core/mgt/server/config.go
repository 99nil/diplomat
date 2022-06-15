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

package server

import (
	"os"

	"github.com/99nil/diplomat/pkg/logr"

	badgerstorage "github.com/99nil/dsync/storage/badger"

	"github.com/99nil/diplomat/pkg/k8s"
	"github.com/99nil/gopkg/server"
)

func Environ(envPrefix string) *Config {
	cfg := &Config{}
	cfg.Server.Port = 3000
	cfg.Instance.Name = os.Getenv("HOSTNAME")
	instanceName := os.Getenv(envPrefix + "_INSTANCE_NAME")
	if instanceName != "" {
		cfg.Instance.Name = instanceName
	}
	return cfg
}

type Config struct {
	Logger     logr.Config   `json:"logger,omitempty"`
	Instance   Instance      `json:"instance,omitempty"`
	Server     server.Config `json:"server,omitempty"`
	Kubernetes *k8s.Config   `json:"kubernetes,omitempty"`
	Storage    Storage       `json:"storage,omitempty"`
}

func (c *Config) Complete() {
	// TODO When supporting multiple storage backends, we can remove this default setting.
	if c.Storage.Badger == nil {
		c.Storage.Badger = &badgerstorage.Config{
			Path: "/tmp/diplomat/storage",
		}
	}
}

func (c *Config) Validate() error {
	return nil
}

type Instance struct {
	Name string `json:"name"`
}

type Storage struct {
	Badger *badgerstorage.Config `json:"badger"`
}

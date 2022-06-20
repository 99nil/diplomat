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

package agent

import (
	"errors"
	"os"

	"github.com/99nil/diplomat/pkg/logr"
	badgerstorage "github.com/99nil/dsync/storage/badger"
)

func Environ(envPrefix string) *Config {
	cfg := &Config{}
	cfg.Agent.Name = os.Getenv("HOSTNAME")
	name := os.Getenv(envPrefix + "_AGENT_NAME")
	if name != "" {
		cfg.Agent.Name = name
	}
	return cfg
}

type Config struct {
	Logger  logr.Config  `json:"logger,omitempty"`
	Agent   ConfigAgent  `json:"agent,omitempty"`
	Server  ConfigServer `json:"server"`
	Storage Storage      `json:"storage,omitempty"`
}

func (c *Config) Complete() {
	// TODO When supporting multiple storage backends, we can remove this default setting.
	if c.Storage.Badger == nil {
		c.Storage.Badger = &badgerstorage.Config{
			Path: "/tmp/diplomat/agent/storage",
		}
	}
}

func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return errors.New("server.host must exist")
	}
	return nil
}

type ConfigAgent struct {
	Name string `json:"name"`
}

type ConfigServer struct {
	Host string `json:"host" yaml:"host"`
}

type Storage struct {
	Badger *badgerstorage.Config `json:"badger"`
}

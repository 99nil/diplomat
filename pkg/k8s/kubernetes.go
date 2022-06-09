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

package k8s

import (
	"encoding/base64"
	"strings"

	"k8s.io/client-go/rest"
)

type Config struct {
	Host    string `json:"host"`
	Token   string `json:"token"`
	CaCrt   string `json:"ca_crt"`
	SkipTLS bool   `json:"skip_tls"`
}

func NewRestConfig(cfg *Config) (*rest.Config, error) {
	if cfg == nil {
		return rest.InClusterConfig()
	}

	token := strings.ReplaceAll(cfg.Token, " ", "")
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	restConfig := &rest.Config{
		BearerToken: string(tokenBytes),
		Host:        cfg.Host,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	if !cfg.SkipTLS {
		restConfig.Insecure = false
		restConfig.CAData = []byte(cfg.CaCrt)
	}
	return restConfig, nil
}

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
	dockerclient "github.com/docker/docker/client"
)

func Run(cfg *Config, dockerClient *dockerclient.Client) error {
	// TODO 支持插件的升级及重启
	// TODO 支持自升级
	// TODO 监听插件的健康状态
	return nil
}

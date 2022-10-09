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
	"context"
	"fmt"

	"github.com/99nil/diplomat/core/component"
	"github.com/99nil/diplomat/pkg/k8s"
	"github.com/99nil/diplomat/static"

	dockerclient "github.com/docker/docker/client"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func Run(cfg *Config,
	dockerClient *dockerclient.Client,
	kubeClient *kubernetes.Clientset,
	dynamicClient dynamic.Interface) error {
	resources, err := k8s.ParseAllYamlToObject(static.RavenYaml)
	if err != nil {
		return err
	}
	ri := component.RavenInstallTool{
		Ctx:           context.Background(),
		Resources:     resources,
		KubeClient:    *kubeClient,
		DynamicClient: dynamicClient,
	}
	ok, err := ri.PreInstall()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("failed to install assistant, please check your network plugins whether it has been installed")
	}

	return ri.Install()
	// TODO 支持插件的升级及重启
	// TODO 支持自升级
	// TODO 监听插件的健康状态
}

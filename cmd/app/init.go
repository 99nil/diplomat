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

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/99nil/diplomat/core/assistant"
	"github.com/99nil/diplomat/pkg/common"
	"github.com/99nil/diplomat/pkg/k8s"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/diplomat/static"
	"github.com/99nil/gopkg/server"

	"github.com/spf13/cobra"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
)

func NewAssistantInitCommand(globalOpt *common.GlobalOption) *cobra.Command {
	opt := &common.InitOption{}
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Initial installation",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// parse config
			cfg := assistant.Environ(globalOpt.EnvPrefix)
			if err := server.ParseConfigWithEnv(globalOpt.Config, cfg, globalOpt.EnvPrefix); err != nil {
				return fmt.Errorf("parse config failed: %v", err)
			}
			cfg.Complete()
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}
			loggerIns := logr.NewLogrusInstance(&cfg.Logger)
			logr.SetDefault(loggerIns)

			// init client
			restConfig, err := clientcmd.BuildConfigFromFlags("", globalOpt.KubeConfig)
			if err != nil {
				return err
			}
			restConfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)
			kubeClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return fmt.Errorf("init kubernetes client failed: %v", err)
			}
			dynamicClient, err := dynamic.NewForConfig(restConfig)
			if err != nil {
				return fmt.Errorf("init dynamic client failed: %v", err)
			}
			//dockerClient, err := client.NewClientWithOpts(client.FromEnv)
			//if err != nil {
			//	return err
			//}

			kubeEdgeResources, err := k8s.ParseAllYamlToObject(static.KubeEdgeYaml)
			if err != nil {
				return err
			}
			ravenResources, err := k8s.ParseAllYamlToObject(static.RavenYaml)
			if err != nil {
				return err
			}

			cfg.KubeEdgeResources = kubeEdgeResources
			cfg.RavenResources = ravenResources

			advAddrs := strings.Split(opt.AdvertiseAddress, ",")
			advAddr := advAddrs[0]
			if len(advAddr) > 0 {
				opt.CurrentIP = advAddr
			} else {
				// 获取IP地址
				hostInter, err := utilnet.ChooseHostInterface()
				if err != nil {
					return fmt.Errorf("get host interface failed: %v", err)
				}
				opt.CurrentIP = hostInter.String()
				if opt.AdvertiseAddress == "" {
					opt.AdvertiseAddress = opt.CurrentIP
				}
			}

			logr.Infof("logger level: %v", logr.Level())
			logr.Debugf("%#v", cfg)

			inst := assistant.NewInitInstance(globalOpt, opt, cfg, kubeClient, dynamicClient)
			return inst.Run(context.Background())
		},
	}
	initAssistantFlags(cmd, globalOpt, opt)
	return cmd
}

func initAssistantFlags(cmd *cobra.Command, globalOpt *common.GlobalOption, opt *common.InitOption) {
	cmd.Flags().StringVar(&opt.Type, "type", common.TypeContainer, "Install KubeEdge Type: binary、container")
	cmd.Flags().StringVar(&opt.Env, "env", common.EnvDev, "Install Env: dev、prod")
	cmd.Flags().StringVar(&opt.Domain, "domain", "", "set domain names,eg: www.99nil.com,www.baidu.com")

	cmd.Flags().StringVar(&opt.AdvertiseAddress, "advertise-address", "",
		"set IPs in cloudcore's certificate SubAltNames field,eg: 10.10.102.78,10.10.102.79")

	cmd.Flags().StringArrayVar(&opt.Disable, "disable", nil,
		"Components that need to be disabled, if there are multiple components separated by commas. eg: CloudStream")

	cmd.Flags().StringVar(&opt.K8sCertPath, "k8s-cert-path", "/etc/kubernetes/pki",
		"Use this key to set the Kubernetes certificate path, eg: /etc/kubernetes/pki; If not provide, will not generate certificate")

	cmd.Flags().StringVar(&opt.ImageRepository, "image-repository", "",
		"Choose a container registry to pull control plane images from (default \"docker.io/99nil\")")
	cmd.Flags().StringVar(&opt.ImageRepositoryUsername, "image-repository-username", "", "Choose a container registry username")
	cmd.Flags().StringVar(&opt.ImageRepositoryPassword, "image-repository-password", "", "Choose a container registry password")
}

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
	"fmt"

	"github.com/99nil/diplomat/core/assistant"
	mgtAgent "github.com/99nil/diplomat/core/mgt/agent"
	mgtServer "github.com/99nil/diplomat/core/mgt/server"
	"github.com/99nil/diplomat/pkg/k8s"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/gopkg/ctr"
	"github.com/99nil/gopkg/server"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/flowcontrol"
)

func NewAssistant() *cobra.Command {
	use := "assistant"
	opt := NewOption(use)
	cmd := &cobra.Command{
		Use:          opt.Module,
		Short:        "Assistant",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := assistant.Environ(opt.EnvPrefix)
			if err := server.ParseConfigWithEnv(opt.Config, cfg, opt.EnvPrefix); err != nil {
				return fmt.Errorf("parse config failed: %v", err)
			}
			cfg.Complete()
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}
			loggerIns := logr.NewLogrusInstance(&cfg.Logger)
			logr.SetDefault(loggerIns)

			logr.Infof("logger level: %v", logr.Level())
			logr.Debugf("%#v", cfg)

			dockerClient, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				return err
			}

			return assistant.Run(cfg, dockerClient)
		},
	}

	opt.CompleteFlags(cmd)
	return cmd
}

func NewMgtServer() *cobra.Command {
	use := "mgt-server"
	opt := NewOption(use)
	cmd := &cobra.Command{
		Use:          opt.Module,
		Short:        "Management Server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mgtServer.Environ(opt.EnvPrefix)
			if err := server.ParseConfigWithEnv(opt.Config, cfg, opt.EnvPrefix); err != nil {
				return fmt.Errorf("parse config failed: %v", err)
			}
			cfg.Complete()
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}
			loggerIns := logr.NewLogrusInstance(&cfg.Logger)
			logr.SetDefault(loggerIns)
			ctr.InitLogger(loggerIns)
			logr.Infof("logger level: %v", logr.Level())
			logr.Debugf("%#v", cfg)

			restConfig, err := k8s.NewRestConfig(cfg.Kubernetes)
			if err != nil {
				return fmt.Errorf("init kubernetes rest config failed: %v", err)
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

			return mgtServer.Run(cfg, kubeClient, dynamicClient)
		},
	}

	opt.CompleteFlags(cmd)
	return cmd
}

func NewMgtAgent() *cobra.Command {
	use := "mgt-agent"
	opt := NewOption(use)
	cmd := &cobra.Command{
		Use:          opt.Module,
		Short:        "Management Agent",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := mgtAgent.Environ(opt.EnvPrefix)
			if err := server.ParseConfigWithEnv(opt.Config, cfg, opt.EnvPrefix); err != nil {
				return fmt.Errorf("parse config failed: %v", err)
			}
			cfg.Complete()
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}
			loggerIns := logr.NewLogrusInstance(&cfg.Logger)
			logr.SetDefault(loggerIns)
			logr.Infof("logger level: %v", logr.Level())
			logr.Debugf("%#v", cfg)

			return mgtAgent.Run(cfg)
		},
	}

	opt.CompleteFlags(cmd)
	return cmd
}

func NewProxyServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "proxy-server",
		Short:        "Proxy Server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO 解析配置
			// TODO 校验配置
			// TODO proxy server 是个有状态的服务
			// TODO 需要保证 边缘Node 资源的 InternalIP 为 proxy server 地址，kubeletEndpoint 端口为 proxy server 监听端口
			// TODO 当 边缘节点 发起长连接时，更新node资源的 InternalIP 为对应 proxy server 地址
			// TODO 开放接口供 k8s APIServer 调用 logs/exec/metrics 等，并透传至边缘节点获取数据
			return nil
		},
	}
	return cmd
}

func NewProxyAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "proxy-agent",
		Short:        "Proxy Agent",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO 解析配置
			// TODO 校验配置
			// TODO 向 proxy server 发起请求
			// TODO 根据 proxy server 的请求，发送数据
			return nil
		},
	}
	return cmd
}

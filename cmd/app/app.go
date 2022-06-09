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
	"os"
	"strings"

	mgtServer "github.com/99nil/diplomat/core/mgt/server"
	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/k8s"
	"github.com/99nil/gopkg/server"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type MgtServerOption struct {
	Config string
}

func NewMgtServer() *cobra.Command {
	prefix := constants.ProjectName + "_manage_server"
	opt := &MgtServerOption{}
	cmd := &cobra.Command{
		Use:          "mgt-server",
		Short:        "Management Server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &mgtServer.Config{}
			if err := server.ParseConfigWithEnv(opt.Config, cfg, strings.ToUpper(prefix)); err != nil {
				return fmt.Errorf("parse config failed: %v", err)
			}
			cfg.Complete()
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}

			restConfig, err := k8s.NewRestConfig(cfg.Kubernetes)
			if err != nil {
				return fmt.Errorf("init kubernetes rest config failed: %v", err)
			}
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

	cfgPathEnv := os.Getenv(strings.ToUpper(prefix + "_config"))
	if cfgPathEnv == "" {
		cfgPathEnv = "config/config.yaml"
	}
	cmd.Flags().StringVarP(&opt.Config, "config", "c", cfgPathEnv,
		"config file (default is $HOME/config.yaml)")
	return cmd
}

func NewMgtAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "mgt-agent",
		Short:        "Management Agent",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO 解析配置
			// TODO 校验配置
			// TODO 启动 健康检查管理
			// TODO 启动 同步服务，并通过 dsync manifest 同步数据到本地存储
			// TODO 启动 APIServer（轻量化 k8s APIServer），供节点直接调用，获取本地存储中的资源，增/改/删 操作需要透传至云端
			return nil
		},
	}
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

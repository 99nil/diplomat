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

	"github.com/99nil/diplomat/core/assistant"
	"github.com/99nil/diplomat/pkg/common"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
)

func NewAssistantEdgeResetCommand(globalOpt *common.GlobalOption) *cobra.Command {
	opt := &common.ResetOption{}
	cmd := &cobra.Command{
		Use:          "reset-edge",
		Short:        "Reset diplomat edge side",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ins := assistant.NewEdgeResetInstance(globalOpt, opt)
			return ins.Run(context.Background())
		},
	}

	cmd.Flags().BoolVar(&opt.Force, "force", false, "Reset the node without prompting for confirmation")
	return cmd
}

func NewAssistantCloudResetCommand(globalOpt *common.GlobalOption) *cobra.Command {
	opt := &common.ResetOption{}
	cmd := &cobra.Command{
		Use:          "reset-cloud",
		Short:        "Reset diplomat cloud side",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig, err := clientcmd.BuildConfigFromFlags("", globalOpt.KubeConfig)
			if err != nil {
				return err
			}
			restConfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)
			kubeClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return err
			}

			ins := assistant.NewCloudResetInstance(globalOpt, opt, kubeClient)
			return ins.Run(context.Background())
		},
	}

	cmd.Flags().BoolVar(&opt.Force, "force", false, "Reset the node without prompting for confirmation")
	return cmd
}

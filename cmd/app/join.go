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

	"github.com/99nil/diplomat/global/constants"

	"github.com/99nil/diplomat/core/assistant"

	"github.com/99nil/diplomat/pkg/common"
	"github.com/spf13/cobra"
)

const (
	edgeJoinLongDescription = `
"assistant join" command bootstraps Edge's worker node (at the edge) component.
It will also connect with cloud component to receive
further instructions and forward telemetry data from
devices to cloud
`
	edgeJoinExample = `
assistant join --cloudcore-ipport=<ip:port address> --edgenode-name=<unique string as edge identifier>

  - For this command --cloudcore-ipport flag is a required option
  - This command will download and install the default version of pre-requisites and Edge

assistant join --cloudcore-ipport=10.20.30.40:10000 --edgenode-name=testing123
`
)

func NewAssistantJoinCommand(globalOpt *common.GlobalOption) *cobra.Command {
	opt := newOption()
	cmd := &cobra.Command{
		Use:          "join",
		Short:        "Bootstraps edge component. Checks and install (if required) the pre-requisites. Execute it on any edge node machine you wish to join",
		Long:         edgeJoinLongDescription,
		Example:      edgeJoinExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ins := assistant.NewJoinInstance(globalOpt, opt)
			return ins.Run(context.Background())
		},
	}
	addJoinOtherFlags(cmd, globalOpt, opt)
	return cmd
}

func addJoinOtherFlags(cmd *cobra.Command, globalOpt *common.GlobalOption, opt *common.JoinOption) {
	cmd.Flags().StringVarP(&opt.CloudCoreIPPort, constants.CloudCoreIPPort, "e", opt.CloudCoreIPPort,
		"IP:Port address of Edge CloudCore")

	if err := cmd.MarkFlagRequired(constants.CloudCoreIPPort); err != nil {
		fmt.Printf("mark flag required failed with error: %v\n", err)
	}

	cmd.Flags().StringVarP(&opt.EdgeNodeName, constants.EdgeNodeName, "i", opt.EdgeNodeName,
		"Edge Node unique identification string, If flag not used then the command will generate a unique id on its own")

	cmd.Flags().StringVar(&opt.Namespace, "namespace", opt.Namespace,
		"Set to the edge node of the specified namespace")

	cmd.Flags().StringSliceVar(&opt.Labels, constants.Labels, opt.Labels,
		"Set to the edge node of the specified labels")

	cmd.Flags().StringVar(&opt.CGroupDriver, constants.CGroupDriver, opt.CGroupDriver,
		"CGroupDriver that uses to manipulate cgroups on the host (cgroupfs or systemd), the default value is cgroupfs")

	cmd.Flags().StringVar(&opt.CertPath, constants.CertPath, opt.CertPath,
		fmt.Sprintf("The certPath used by edgecore, the default value is %s", constants.DefaultCertPath))

	cmd.Flags().StringVarP(&opt.RuntimeType, constants.RuntimeType, "r", opt.RuntimeType,
		"Container runtime type")

	cmd.Flags().StringVarP(&opt.RemoteRuntimeEndpoint, constants.RemoteRuntimeEndpoint, "p", opt.RemoteRuntimeEndpoint,
		"KubeEdge Edge Node RemoteRuntimeEndpoint string, If flag not set, it will use unix:///var/run/dockershim.sock")

	cmd.Flags().StringVarP(&opt.Token, constants.Token, "t", opt.Token,
		"Used for edge to apply for the certificate")

	cmd.Flags().StringVarP(&opt.CertPort, constants.CertPort, "s", opt.CertPort,
		"The port where to apply for the edge certificate")

	cmd.Flags().BoolVar(&opt.WithMQTT, "with-mqtt", opt.WithMQTT,
		`Use this key to set whether to install and start MQTT Broker by default`)

	cmd.Flags().StringVar(&opt.ImageRepository, constants.ImageRepository, opt.ImageRepository,
		`Use this key to decide which image repository to pull images from.`,
	)
}

func newOption() *common.JoinOption {
	joinOptions := &common.JoinOption{}
	joinOptions.WithMQTT = true
	joinOptions.CGroupDriver = constants.CGroupDriverCGroupFS
	joinOptions.CertPath = constants.DefaultCertPath
	joinOptions.RuntimeType = constants.DockerContainerRuntime
	return joinOptions
}

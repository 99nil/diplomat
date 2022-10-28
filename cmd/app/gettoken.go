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
	"github.com/99nil/diplomat/pkg/common"

	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
)

// NewAssistantGetTokenCommand Get token to join edge node to cloud
func NewAssistantGetTokenCommand(globalOpt *common.GlobalOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gettoken",
		Short: "Get token to join edge node to cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			secret, err := kubeClient.CoreV1().Secrets(constants.SystemNamespace).
				Get(context.Background(), constants.TokenSecretName, metaV1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get token, err: %v", err)
			}
			fmt.Println(string(secret.Data[constants.TokenDataName]))
			return nil
		},
	}
	return cmd
}

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
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/common"
	"github.com/99nil/diplomat/pkg/exec"

	"github.com/zc2638/aide"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
)

func CheckPort(resources [][]unstructured.Unstructured, opt *common.InitOption, portSet *common.PortSet) aide.StepFunc {
	return func(sc *aide.StepContext) {
		sc.Log("Start to check port conflict")
		if opt.Env == common.EnvDev {
			sc.Log("Check port conflict Skipped")
			return
		}

		// 读取所有service yaml
		if err := generateAllPortsByResource(resources, opt, portSet); err != nil {
			sc.Errorf("Generate All ports failed: %v", err)
		}

		format := "[ \"$(netstat -nl | awk '{if($4 ~ /^.*:%v/){print $0;}}')\" == '' ]"
		portSet.Range(func(name string, port int32) bool {
			if err := sc.Bash(fmt.Sprintf(format, port)); err != nil {
				sc.Logfl(aide.WarnLevel, "Check %s service port %d conflict, random port will be used", name, port)
				portSet.Remove(name)
			} else {
				sc.Logf("Check %s service port %d success", name, port)
			}
			return true
		})
		sc.Log("Check port conflict completed")
	}
}

func CheckK8s() aide.StepFunc {
	return func(sc *aide.StepContext) {
		sc.Log("Start to check k8s status")
		if err := sc.Shell("kubectl get node"); err != nil {
			sc.Error(err)
		}
		sc.Log("K8s is running")
	}
}

func generateAllPortsByResource(src [][]unstructured.Unstructured, opt *common.InitOption, portSet *common.PortSet) error {
	for _, resources := range src {
		if err := fillPort(resources, opt, portSet); err != nil {
			return err
		}
	}
	return nil
}

func fillPort(resources []unstructured.Unstructured, opt *common.InitOption, portSet *common.PortSet) error {
	for _, obj := range resources {
		if obj.GetKind() != constants.KindService {
			continue
		}
		var svc coreV1.Service
		if err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &svc); err != nil {
			return fmt.Errorf("convert unstructured object to Service failed: %v", err)
		}

		// 禁用NodePort
		isSpecial := common.IsSpecialName(svc.Name)
		for _, port := range svc.Spec.Ports {
			if opt.Env != common.EnvDev && !isSpecial {
				portSet.Remove(port.Name)
				continue
			}
			if port.NodePort > 0 {
				portSet.Add(port.Name, port.NodePort)
			}
		}
	}
	return nil
}

func CheckCloud(kubeClient kubernetes.Interface) aide.StepFunc {
	return func(sc *aide.StepContext) {
		// check cloudcore for container deploy
		_, err := kubeClient.AppsV1().Deployments(constants.SystemNamespace).
			Get(sc.Context(), constants.CloudComponent, metaV1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				sc.Errorf("Check cloud container failed: %v", err)
			}
		} else {
			sc.Errorf("KubeEdge cloudcore container is already running on this node, please run reset to clean up first")
		}

		// check cloudcore process
		running, err := isProcessRunning(constants.CloudComponent)
		if err != nil {
			sc.Errorf("Check cloud process failed: %v", err)
		}
		if running {
			sc.Errorf("KubeEdge cloudcore is already running on this node, please run reset to clean up first")
		}
		sc.Log("Check cloud process successful")
	}
}

// isProcessRunning checks if the given process is running or not
func isProcessRunning(proc string) (bool, error) {
	cmd := exec.NewCommand(fmt.Sprintf("pidof %s 2>&1", proc))
	err := cmd.Exec()

	if cmd.ExitCode == 0 {
		return true, nil
	} else if cmd.ExitCode == 1 {
		return false, nil
	}
	return false, err
}

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
	"os"
	"time"

	"github.com/99nil/diplomat/global/constants"

	"github.com/AlecAivazis/survey/v2"
	keutil "github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/util"
	"github.com/zc2638/aide"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	utilruntime "k8s.io/kubernetes/cmd/kubeadm/app/util/runtime"
	utilsexec "k8s.io/utils/exec"
)

// ChooseResetConfirm reset之前先确认
func ChooseResetConfirm(isEdgeNode bool) aide.StepFunc {
	return func(sc *aide.StepContext) {
		if isEdgeNode {
			sc.Log("This is an edge side")
		} else {
			sc.Log("This is a cloud side")
		}
		var isReset bool
		prompt := &survey.Confirm{
			Message: "[reset] WARNING: Changes made to this host by 'init' will be reverted.\n  [reset] Are you sure you want to proceed? (default: yes)",
			Default: true,
		}
		if err := survey.AskOne(prompt, &isReset); err != nil {
			sc.Error(err)
		}
		if !isReset {
			sc.Log("reset was cancelled!")
			sc.Exit()
		}
	}
}

func ResetCloud(kubeClient kubernetes.Interface) aide.StepFunc {
	return func(sc *aide.StepContext) {
		// For the cloudcore deploy on the container: When deleting the namespace, cloudcore is deleted as well
		// 停止服务
		systemdExist := keutil.HasSystemd()
		if systemdExist {
			// remove the system service.
			serviceFilePath := fmt.Sprintf("/etc/systemd/system/%s.service", constants.CloudComponent)
			serviceFileRemoveExec := fmt.Sprintf("&& sudo rm %s", serviceFilePath)
			if _, err := os.Stat(serviceFilePath); err != nil && os.IsNotExist(err) {
				serviceFileRemoveExec = ""
			}
			command := fmt.Sprintf("sudo systemctl stop %s.service && sudo systemctl disable %s.service %s && sudo systemctl daemon-reload",
				constants.CloudComponent, constants.CloudComponent, serviceFileRemoveExec)
			if err := sc.Shell(command); err != nil {
				sc.Logfl(aide.WarnLevel, "Stop diplomat cloud failed: %v", err)
			} else {
				sc.Log("Stop diplomat cloud successful")
			}
		} else {
			if err := sc.Shell(fmt.Sprintf("pkill %s", constants.CloudComponent)); err != nil {
				sc.Logfl(aide.WarnLevel, "Stop diplomat cloud failed: %v", err)
			} else {
				sc.Log("Stop diplomat cloud successful")
			}
		}

		// 清理目录
		if err := os.RemoveAll(constants.LogDir); err != nil {
			sc.Logfl(aide.WarnLevel, "Reset diplomat management directory `%s` failed: %v", constants.LogDir, err)
		}
		if err := os.RemoveAll(constants.RootDir); err != nil {
			sc.Logfl(aide.WarnLevel, "Reset diplomat management directory `%s` failed: %v", constants.RootDir, err)
		}
		if err := os.RemoveAll(constants.ResourceDir); err != nil {
			sc.Logfl(aide.WarnLevel, "Reset diplomat management directory `%s` failed: %v", constants.RootDir, err)
		}

		// 清理k8s资源
		rqt, err := labels.NewRequirement("diplomat", selection.Equals, []string{"diplomat"})
		if err != nil {
			sc.Errorf("Failed to create k8s requirement, err: %s", err)
		}
		selector := labels.NewSelector().Add(*rqt)
		sc.Log("Start to clean resource")
		if err := kubeClient.RbacV1().ClusterRoleBindings().DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat clusterrolebinding: %s\n", err)
		}
		if err := kubeClient.RbacV1().ClusterRoles().DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat clusterrole: %s\n", err)
		}
		if err := kubeClient.PolicyV1beta1().PodSecurityPolicies().DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat policy: %s\n", err)
		}
		if err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat mutating webhook config: %s\n", err)
		}
		if err := kubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat validating webhook config: %s\n", err)
		}
		// delete kube-system namespace raven-controller resource
		if err := kubeClient.RbacV1().RoleBindings(constants.RavenControllerNamespace).DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat rolebinding: %s\n", err)
		}
		if err := kubeClient.RbacV1().Roles(constants.RavenControllerNamespace).DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat role: %s\n", err)
		}
		if err := kubeClient.CoreV1().ServiceAccounts(constants.RavenControllerNamespace).DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat serviceAccount: %s\n", err)
		}
		if err := kubeClient.CoreV1().Secrets(constants.RavenControllerNamespace).DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat secret: %s\n", err)
		}
		if err := kubeClient.AppsV1().Deployments(constants.RavenControllerNamespace).DeleteCollection(
			sc.Context(),
			metaV1.DeleteOptions{},
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			}); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Failed to delete diplomat deployment: %s\n", err)
		}
		result, err := kubeClient.CoreV1().Services(constants.RavenControllerNamespace).List(
			sc.Context(),
			metaV1.ListOptions{
				LabelSelector: selector.String(),
			})
		if err == nil {
			for _, svc := range result.Items {
				if err := kubeClient.CoreV1().Services(constants.RavenControllerNamespace).Delete(
					sc.Context(),
					svc.Name,
					metaV1.DeleteOptions{}); err != nil {
					sc.Logfl(aide.WarnLevel, "Failed to delete diplomat service: %s\n", err)
				}
			}
		} else {
			if !apierrors.IsNotFound(err) {
				sc.Logfl(aide.WarnLevel, "Failed to list diplomat service: %s\n", err)
			}
		}

		if err := kubeClient.CoreV1().Namespaces().Delete(
			sc.Context(), constants.DefaultNamespace, *metaV1.NewDeleteOptions(0)); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Clean kubernetes resources failed: %v", err)
		}
		if err := kubeClient.CoreV1().Namespaces().Delete(
			sc.Context(), constants.SystemNamespace, *metaV1.NewDeleteOptions(0)); err != nil && !apierrors.IsNotFound(err) {
			sc.Logfl(aide.WarnLevel, "Clean kubernetes resources failed: %v", err)
		}
		if err := checkNS(sc.Context(), kubeClient, constants.DefaultNamespace); err != nil {
			sc.Logfl(aide.WarnLevel, "delete namespace %s failed: %v", err)
		}
		if err := checkNS(sc.Context(), kubeClient, constants.SystemNamespace); err != nil {
			sc.Logfl(aide.WarnLevel, "delete namespace %s failed: %v", err)
		}
		sc.Log("Reset diplomat edge node successful")
	}
}

func checkNS(ctx context.Context, kubeClient kubernetes.Interface, namespace string) error {
	for i := 0; i < 20; i++ {
		_, err := kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metaV1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 5)
	}
	return fmt.Errorf("delete namespace %s timeout", namespace)
}

func ResetEdge() aide.StepFunc {
	return func(sc *aide.StepContext) {
		// 停止服务
		systemdExist := keutil.HasSystemd()
		if systemdExist {
			// remove the system service.
			serviceFilePath := fmt.Sprintf("/etc/systemd/system/%s.service", constants.EdgeComponent)
			serviceFileRemoveExec := fmt.Sprintf("&& sudo rm %s", serviceFilePath)
			if _, err := os.Stat(serviceFilePath); err != nil && os.IsNotExist(err) {
				serviceFileRemoveExec = ""
			}
			command := fmt.Sprintf("sudo systemctl stop %s.service && sudo systemctl disable %s.service %s && sudo systemctl daemon-reload",
				constants.EdgeComponent, constants.EdgeComponent, serviceFileRemoveExec)
			if err := sc.Shell(command); err != nil {
				sc.Logfl(aide.WarnLevel, "Stop diplomat edge failed: %v", err)
			} else {
				sc.Log("Stop diplomat edge successful")
			}
		} else {
			if err := sc.Shell(fmt.Sprintf("pkill %s", constants.EdgeComponent)); err != nil {
				sc.Logfl(aide.WarnLevel, "Stop diplomat edge failed: %v", err)
			} else {
				sc.Log("Stop diplomat edge successful")
			}
		}

		// 清理MQTT容器
		if err := sc.Shell(fmt.Sprintf("docker rm -f %s", constants.MQTTName)); err != nil {
			sc.Logfl(aide.WarnLevel, "Remove MQTT service container failed: %v", err)
		} else {
			sc.Log("Remove MQTT service successful")
		}

		// 边缘端需要清理运行中的容器
		if err := removeK8sContainers(); err != nil {
			sc.Logfl(aide.WarnLevel, "Clean kubernetes containers failed: %v", err)
		} else {
			sc.Log("Clean kubernetes containers successful")
		}

		// 清理目录
		if err := os.RemoveAll(constants.LogDir); err != nil {
			sc.Logfl(aide.WarnLevel, "Reset diplomat management directory `%s` failed: %v", constants.LogDir, err)
		}
		if err := os.RemoveAll(constants.RootDir); err != nil {
			sc.Logfl(aide.WarnLevel, "Reset diplomat management directory `%s` failed: %v", constants.RootDir, err)
		}
		if err := os.RemoveAll(constants.ResourceDir); err != nil {
			sc.Logfl(aide.WarnLevel, "Reset diplomat management directory `%s` failed: %v", constants.RootDir, err)
		}
		sc.Log("Reset diplomat diplomat node successful")
	}
}

func removeK8sContainers() error {
	criSocketPath, err := utilruntime.DetectCRISocket()
	if err != nil {
		return err
	}
	containerRuntime, err := utilruntime.NewContainerRuntime(utilsexec.New(), criSocketPath)
	if err != nil {
		return err
	}
	containers, err := containerRuntime.ListKubeContainers()
	if err != nil {
		return err
	}
	return containerRuntime.RemoveContainers(containers)
}

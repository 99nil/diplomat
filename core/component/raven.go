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

package component

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/common"
	"github.com/99nil/diplomat/pkg/exec"
	"github.com/99nil/diplomat/pkg/util"
	"github.com/99nil/diplomat/static"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	coreV1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

var (
	cptFlannel = []string{"diplomat-flannel-edge", "diplomat-flannel"}
	cptRaven   = []string{"diplomat-raven-agent", "diplomat-raven-agent-edge"}
)

var cptName = append(cptFlannel, cptRaven...)

var FlannelVersion = "v0.7.4-diplomat"

var RavenVersion = "v0.0.1-diplomat"

type RavenInstallTool struct {
	Ctx context.Context

	Resources     [][]unstructured.Unstructured
	KubeClient    kubernetes.Interface
	DynamicClient dynamic.Interface
}

// PreInstall check whether the environment meets installation requirements.
func (t *RavenInstallTool) PreInstall(ctx context.Context) error {
	// check ns
	if _, err := t.KubeClient.CoreV1().Namespaces().
		Get(t.Ctx, constants.DefaultNamespace, metaV1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			if _, err = t.KubeClient.CoreV1().Namespaces().
				Create(
					t.Ctx,
					&coreV1.Namespace{
						ObjectMeta: metaV1.ObjectMeta{
							Name: constants.DefaultNamespace,
						},
					},
					metaV1.CreateOptions{}); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	// network plugins
	if err := t.check(cptFlannel); err != nil {
		return err
	}

	// raven
	if err := t.check(cptRaven); err != nil {
		return err
	}

	return nil
}

func (t *RavenInstallTool) Install(ctx context.Context) error {
	// apply yaml
	gr, err := restmapper.GetAPIGroupResources(t.KubeClient.Discovery())
	if err != nil {
		return fmt.Errorf("[raven install] get API group resources failed: %v", err)
	}
	mapping := restmapper.NewDiscoveryRESTMapper(gr)

	// The KubeEdge permissions required by the Raven component had been applied
	for _, resources := range t.Resources {
		wg, ctx := errgroup.WithContext(t.Ctx)
		for _, obj := range resources {
			currentObj := obj.DeepCopy()
			wg.Go(func() error {
				gvk := currentObj.GroupVersionKind()
				logrus.Debugf("[raven install] apply resource %s %s", gvk, currentObj.GetName())

				restMapping, err := mapping.RESTMapping(gvk.GroupKind(), gvk.Version)
				if err != nil {
					return err
				}

				var resourceInter dynamic.ResourceInterface
				if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
					if currentObj.GetNamespace() == "" {
						currentObj.SetNamespace(constants.DefaultNamespace)
					}
					resourceInter = t.DynamicClient.Resource(restMapping.Resource).Namespace(currentObj.GetNamespace())
				} else {
					resourceInter = t.DynamicClient.Resource(restMapping.Resource)
				}

				return common.ApplyResource(ctx, resourceInter, currentObj)
			})
		}
		if err = wg.Wait(); err != nil {
			return fmt.Errorf("[raven install] apply resource failed: %v", err)
		}
	}

	// install custom specified architecture binary 'host-local'
	if err = util.ExistsAndCreateDir(constants.CNIBinPath); err != nil {
		return fmt.Errorf("[raven install] create dir failed, err: %v", err)
	}

	fileName := "cni-plugins-" + runtime.GOARCH + "-" + FlannelVersion + ".tgz"
	filePath := static.FlannelBin + "/" + fileName
	if !util.IsFile(filePath) {
		return fmt.Errorf("[raven install] failed to locate file %s", fileName)
	}

	readFile, err := static.EmbedResource.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileHostPath := "/tmp/diplomat/" + fileName
	if err = util.ExistsAndCreateDir("/tmp/diplomat/"); err != nil {
		return fmt.Errorf("[raven install] create dir failed, err: %v", err)
	}
	if err = os.WriteFile(fileHostPath, readFile, 0666); err != nil {
		return err
	}

	if err = exec.NewCommand(
		fmt.Sprintf("tar -zxvf %s -C %s", fileHostPath, constants.CNIBinPath),
	).Exec(); err != nil {
		return fmt.Errorf("[raven install] failed to exec, err: %v", err)
	}
	return nil
}

func (t *RavenInstallTool) Uninstall(ctx context.Context) error {
	rq, err := labels.NewRequirement("app", selection.In, cptName)
	if err != nil {
		return err
	}

	return t.KubeClient.
		AppsV1().
		Deployments(constants.DefaultNamespace).
		DeleteCollection(t.Ctx,
			metaV1.DeleteOptions{},
			metaV1.ListOptions{LabelSelector: labels.NewSelector().Add(*rq).String()},
		)
}

func (t *RavenInstallTool) Rollback() error {
	// TODO
	return nil
}

func (t *RavenInstallTool) check(label []string) error {
	if len(label) == 0 {
		return errors.New("label must not be null")
	}

	rq, err := labels.NewRequirement("app", selection.In, label)
	if err != nil {
		return err
	}

	list, err := t.KubeClient.
		AppsV1().
		DaemonSets(constants.DefaultNamespace).
		List(t.Ctx, metaV1.ListOptions{
			LabelSelector: labels.NewSelector().Add(*rq).String(),
		})
	if err != nil {
		return err
	}

	if len(list.Items) > 0 {
		b := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Found %s, will be replaced by diplomat custom version. Confirm?", strings.Join(label, ",")),
		}
		if err := survey.AskOne(prompt, &b); err != nil {
			return err
		}
		if !b {
			logrus.Infof("exit Raven pre install check, confirm exit %s.\n", strings.Join(label, ","))
			return fmt.Errorf("user cofirm exit")
		}
	}

	return nil
}

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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	keutil "github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/util"

	"github.com/99nil/diplomat/pkg/edgeconfig"

	"github.com/99nil/diplomat/pkg/util"

	"github.com/99nil/diplomat/pkg/exec"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/common"
	"github.com/99nil/diplomat/static"

	"github.com/zc2638/aide"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

func Advance() aide.StepFunc {
	return func(sc *aide.StepContext) {
		// 创建管理目录
		if err := os.MkdirAll(constants.RootDir, os.ModePerm); err != nil {
			sc.Errorf("Create diplomat management directory `%s` failed: %v", constants.RootDir, err)
		}
		if err := os.MkdirAll(constants.ResourceDir, os.ModePerm); err != nil {
			sc.Errorf("Create diplomat management directory `%s` failed: %v", constants.ResourceDir, err)
		}
		if err := os.MkdirAll(constants.ConfigDir, os.ModePerm); err != nil {
			sc.Errorf("Create diplomat management directory `%s` failed: %v", constants.ConfigDir, err)
		}
		if err := os.MkdirAll(constants.DefaultCertPath, os.ModePerm); err != nil {
			sc.Errorf("Create diplomat management directory `%s` failed: %v", constants.DefaultCertPath, err)
		}
		if err := os.MkdirAll(constants.LogDir, os.ModePerm); err != nil {
			sc.Errorf("Create diplomat management directory `%s` failed: %v", constants.LogDir, err)
		}
		sc.Log("Init diplomat management directory successful.")
	}
}

func AdvanceNS(kubeClient kubernetes.Interface) aide.StepFunc {
	return func(sc *aide.StepContext) {
		ns := &coreV1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.DefaultNamespace,
			},
		}
		_, err := kubeClient.CoreV1().Namespaces().Create(sc.Context(), ns, metav1.CreateOptions{})
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				sc.Errorf("Init default namespace failed: %v", err)
			}
			sc.Log("Default namespace already exist, skipped")
		} else {
			sc.Log("Init default namespace successful")
		}

		ns.Name = constants.SystemNamespace
		_, err = kubeClient.CoreV1().Namespaces().Create(sc.Context(), ns, metav1.CreateOptions{})
		if err == nil {
			sc.Log("Init system namespace successful")
			return
		}
		if apierrors.IsAlreadyExists(err) {
			sc.Log("System namespace already exist, skipped")
			return
		}

		sc.Errorf("Init system namespace failed: %v", err)
	}
}

func compileNewImage(src, registry string) (string, error) {
	ss := strings.Split(src, "/")
	if len(ss) < 3 {
		return "", fmt.Errorf("split image path error:%s", src)
	}
	return registry + "/" + ss[2], nil
}

func buildImgPullSecretData(registry, username, password string) map[string]string {
	auth := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Auth     string `json:"auth"`
	}{
		Username: username,
		Password: password,
		Auth:     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
	}
	encode, err := json.Marshal(auth)
	if err != nil {
		return nil
	}

	arr := strings.Split(registry, "/")
	data := make(map[string]string)
	src := "{\"auths\":{\"" + arr[0] + "\":" + string(encode) + "}}"
	data[".dockerconfigjson"] = src
	return data
}

func AdvanceSecret(kubeClient kubernetes.Interface, opt *common.InitOption, imageSet map[string]string) aide.StepFunc {
	return func(sc *aide.StepContext) {
		sc.Log("Start to init secret")
		if opt.ImageRepository == "" {
			sc.Log("init secret successful")
			return
		}

		for k, v := range imageSet {
			image, err := compileNewImage(v, opt.ImageRepository)
			if err != nil {
				sc.Error(err)
				return
			}
			imageSet[k] = image
		}

		if opt.ImageRepositoryUsername == "" || opt.ImageRepositoryPassword == "" {
			sc.Logf("Use custom image repository (%s), generate secret skipped", opt.ImageRepository)
			return
		}

		secret := &coreV1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: constants.DefaultNamespace,
				Name:      opt.ImageRepository,
			},
			Data:       nil,
			StringData: buildImgPullSecretData(opt.ImageRepository, opt.ImageRepositoryUsername, opt.ImageRepositoryPassword),
			Type:       coreV1.SecretTypeDockerConfigJson,
		}
		_, err := kubeClient.CoreV1().Secrets(secret.Namespace).Create(sc.Context(), secret, metav1.CreateOptions{})
		if err != nil {
			if !apierrors.IsAlreadyExists(err) {
				sc.Errorf("Create image secret failed: %s", err)
			}
			sc.Logl(aide.WarnLevel, "Custom image secret %s has already exists", secret.Name)
			return
		}
		sc.Logf("Use custom image repository (%s), generate secret successful", opt.ImageRepository)
	}
}

func GenerateCerts(opt *common.InitOption) aide.StepFunc {
	return func(sc *aide.StepContext) {
		// admission cert
		targetPath := filepath.Join(constants.RootDir, "gen-admission-cert.sh")
		if err := os.WriteFile(targetPath, static.AdmissionCertScript, os.ModePerm); err != nil {
			sc.Errorf("Generate admission certificate script failed: %v", err)
		}
		if err := sc.Shell(targetPath); err != nil {
			sc.Errorf("Generate admission certificate failed: %v", err)
		}
		sc.Log("Generate admission certificate successful")

		// cloudcore cert
		targetPath = filepath.Join(constants.RootDir, "certgen.sh")
		if err := os.WriteFile(targetPath, static.CoreCertScript, os.ModePerm); err != nil {
			sc.Errorf("Generate cloudcore certificate script failed: %v", err)
		}
		if err := sc.Shell(fmt.Sprintf("%s buildCloudcoreSecret -i %s", targetPath, opt.AdvertiseAddress)); err != nil {
			sc.Errorf("Generate cloudcore certificate failed: %v", err)
		}
		sc.Log("Generate cloudcore certificate successful")

		// cloud stream cert
		advAddr := strings.Join(strings.Split(opt.AdvertiseAddress, ","), " ")
		var domain string
		cmd := fmt.Sprintf(
			"CLOUDCOREIPS=\"%s\" %s stream -c true -p %s",
			advAddr, targetPath, opt.K8sCertPath)
		if opt.Domain != "" {
			domains := strings.Split(opt.Domain, ",")
			domain = domains[0]
			if len(domains) > 1 {
				for i := 1; i < len(domains); i++ {
					domain += " " + domains[i]
				}
			}
			cmd = fmt.Sprintf("CLOUDCORE_DOMAINS=\"%s\" "+cmd, domain)
		}
		sc.Logfl(aide.InfoLevel, "cloud stream cert cmd: %s", cmd)
		if err := sc.Shell(cmd); err != nil {
			sc.Errorf("Generate stream certificate failed: %v", err)
		}
		sc.Log("Generate cloud stream certificate successful")
	}
}

func ApplyResource(
	globalOpt *common.GlobalOption,
	opt *common.InitOption,
	resourceSet [][]unstructured.Unstructured,
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
	portSet *common.PortSet,
	imageOption map[string]string,
) aide.StepFunc {
	return func(sc *aide.StepContext) {
		// TODO 检查版本适应

		gr, err := restmapper.GetAPIGroupResources(kubeClient.Discovery())
		if err != nil {
			sc.Errorf("Get API group resources failed: %v", err)
		}
		mapping := restmapper.NewDiscoveryRESTMapper(gr)

		for _, resources := range resourceSet {
			wg, ctx := errgroup.WithContext(sc.Context())
			for _, obj := range resources {
				currentObj := obj.DeepCopy()
				wg.Go(func() error {
					gvk := currentObj.GroupVersionKind()
					sc.Logf("Apply resource %s %s", gvk, currentObj.GetName())

					restMapping, err := mapping.RESTMapping(gvk.GroupKind(), gvk.Version)
					if err != nil {
						return err
					}

					var resourceInter dynamic.ResourceInterface
					if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
						if currentObj.GetNamespace() == "" {
							currentObj.SetNamespace(constants.DefaultNamespace)
						}
						resourceInter = dynamicClient.Resource(restMapping.Resource).Namespace(currentObj.GetNamespace())
					} else {
						resourceInter = dynamicClient.Resource(restMapping.Resource)
					}

					// 自定义修改
					switch currentObj.GetKind() {
					case constants.KindService:
						if currentObj.GetName() == constants.CloudComponent && opt.Type != common.TypeContainer {
							return nil
						}
						return ApplyService(ctx, kubeClient, currentObj, portSet, opt)
					case constants.KindConfigMap:
						if currentObj.GetName() == constants.CloudComponent && opt.Type != common.TypeContainer {
							return nil
						}
						return applyConfigMap(ctx, kubeClient, resourceInter, currentObj, portSet, opt.CurrentIP, opt.AdvertiseAddress)
					case constants.KindDeployment:
						if currentObj.GetName() == constants.CloudComponent && opt.Type != common.TypeContainer {
							return nil
						}
						return applyDeployment(ctx, kubeClient, resourceInter, currentObj, opt, imageOption)
					case constants.KindDaemonSet:
						return applyDaemonSet(ctx, kubeClient, resourceInter, currentObj, portSet, imageOption, opt.CurrentIP)
					default:
						return common.ApplyResource(ctx, resourceInter, currentObj)
					}
				})
			}
			if err := wg.Wait(); err != nil {
				sc.Errorf("Apply resource failed: %v", err)
			}
		}
		sc.Log("Apply resource successful")
	}
}

func applyConfigMap(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	resourceInter dynamic.ResourceInterface,
	obj *unstructured.Unstructured,
	portSet *common.PortSet,
	hostIP string,
	advertiseAddress string,
) error {
	var cm coreV1.ConfigMap
	if err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &cm); err != nil {
		return fmt.Errorf("convert unstructured object %s to ConfigMap failed: %v", obj.GetName(), err)
	}

	switch obj.GetName() {
	case "cloudcore":
		data, ok := cm.Data["cloudcore.yaml"]
		if !ok {
			return fmt.Errorf("meta resource ConfigMap %s not found config.yaml", obj.GetName())
		}
		var cmMap map[string]interface{}
		if err := yaml.Unmarshal([]byte(data), &cmMap); err != nil {
			return err
		}

		// TODO 优化，获取字段有点恶心
		v1, ok := getMap(cmMap, "modules")
		if !ok {
			return fmt.Errorf("configmap field not found")
		}
		v2, ok := getMap(v1, "cloudHub")
		if !ok {
			return fmt.Errorf("configmap field not found")
		}
		// for testing
		//values, ok := v2["advertiseAddress"].([]interface{})
		//if !ok {
		//	return fmt.Errorf("configmap field not found")
		//}
		//values = []interface{}{strings.Split(advertiseAddress, ",")}

		v2["advertiseAddress"] = strings.Split(advertiseAddress, ",")
		v1["cloudHub"] = v2
		cmMap["modules"] = v1

		marshal, err := yaml.Marshal(cmMap)
		if err != nil {
			return err
		}
		cm.Data["cloudcore.yaml"] = string(marshal)
	}

	toUnstructured, err := pkgruntime.DefaultUnstructuredConverter.ToUnstructured(&cm)
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(toUnstructured)
	return common.ApplyResource(ctx, resourceInter, obj)
}

func setCurrentToMap(data map[string]interface{}, current map[string]interface{}, keys ...string) {
	for _, k := range keys {
		if v, ok := current[k]; ok {
			data[k] = v
		}
	}
}

func getMap(src map[string]interface{}, name string) (map[string]interface{}, bool) {
	if m, ok := src[name]; ok {
		if v, ok := m.(map[string]interface{}); ok {
			return v, true
		}
	}
	return nil, false
}

func ApplyService(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	obj *unstructured.Unstructured,
	portSet *common.PortSet,
	opt *common.InitOption,
) error {
	var svc coreV1.Service
	if err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &svc); err != nil {
		return fmt.Errorf("convert unstructured object %s to Service failed: %v", obj.GetName(), err)
	}

	// 禁用NodePort
	isSpecial := common.IsSpecialName(svc.Name)
	if opt.Env != common.EnvDev && !isSpecial {
		svc.Spec.Type = coreV1.ServiceTypeClusterIP
	}

	ports := make([]coreV1.ServicePort, 0, len(svc.Spec.Ports))
	for _, port := range svc.Spec.Ports {
		switch svc.Spec.Type {
		case coreV1.ServiceTypeNodePort:
			port.NodePort = portSet.Get(port.Name)
		case coreV1.ServiceTypeClusterIP:
			port.NodePort = 0
		default:
		}
		ports = append(ports, port)
	}
	svc.Spec.Ports = ports

	var current *coreV1.Service
	serviceInter := kubeClient.CoreV1().Services(obj.GetNamespace())
	found, err := serviceInter.Get(ctx, svc.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		current, err = serviceInter.Create(ctx, &svc, metav1.CreateOptions{})
	} else {
		// 更新操作的话，使用运行中的ports，不影响原有配置
		currentPorts := make([]coreV1.ServicePort, 0, len(svc.Spec.Ports))
		for _, port := range ports {
			var existPort *coreV1.ServicePort
			for _, v := range found.Spec.Ports {
				if v.Name == port.Name {
					existPort = &v
					break
				}
			}
			if existPort != nil {
				port = *existPort
			}
			currentPorts = append(currentPorts, port)
		}
		found.Spec.Ports = currentPorts

		rv, _ := strconv.ParseInt(found.GetResourceVersion(), 10, 64)
		found.SetResourceVersion(strconv.FormatInt(rv, 10))
		current, err = serviceInter.Update(ctx, found, metav1.UpdateOptions{})
	}
	if err != nil {
		return err
	}
	setCurrentPorts(portSet, current)
	return nil
}

func setCurrentPorts(portSet *common.PortSet, svc *coreV1.Service) {
	for _, port := range svc.Spec.Ports {
		portSet.Add(port.Name, port.NodePort)
	}
}

func applyDeployment(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	resourceInter dynamic.ResourceInterface,
	obj *unstructured.Unstructured,
	opt *common.InitOption,
	imageOption map[string]string,
) error {
	name := obj.GetName()
	// 修改镜像版本
	var deploy appsV1.Deployment
	err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &deploy)
	if err != nil {
		return fmt.Errorf("convert unstructured object %s to Deployment failed: %v", obj.GetName(), err)
	}

	image, ok := imageOption[name]
	if !ok {
		image = deploy.Spec.Template.Spec.Containers[0].Image
	}
	deploy.Spec.Template.Spec.Containers[0].Image = image

	current, err := kubeClient.AppsV1().Deployments(deploy.GetNamespace()).Get(ctx, deploy.GetName(), metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		current.Spec.Template.Spec.Containers[0].Image = image
		deploy = *current
	}

	toUnstructured, err := pkgruntime.DefaultUnstructuredConverter.ToUnstructured(&deploy)
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(toUnstructured)
	return common.ApplyResource(ctx, resourceInter, obj)
}

func applyDaemonSet(
	ctx context.Context,
	kubeClient kubernetes.Interface,
	resourceInter dynamic.ResourceInterface,
	obj *unstructured.Unstructured,
	portSet *common.PortSet,
	imageSet map[string]string,
	hostIP string,
) error {
	var ds appsV1.DaemonSet
	if err := pkgruntime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &ds); err != nil {
		return fmt.Errorf("convert unstructured object %s to DaemonSet failed: %v", obj.GetName(), err)
	}

	if len(ds.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("meta resource DaemonSet %s not found container", obj.GetName())
	}
	container := ds.Spec.Template.Spec.Containers[0]
	if image, ok := imageSet[ds.Name]; ok {
		container.Image = image
	}

	current, err := kubeClient.AppsV1().DaemonSets(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		// 存在情况：只更新镜像版本及不存在的环境变量
		currentContainer := current.Spec.Template.Spec.Containers[0]
		var envs []coreV1.EnvVar
		for _, v := range container.Env {
			var existEnv *coreV1.EnvVar
			for _, cv := range currentContainer.Env {
				if v.Name == cv.Name {
					existEnv = &cv
					break
				}
			}
			if existEnv != nil {
				envs = append(envs, *existEnv)
			} else {
				envs = append(envs, v)
			}
		}
		currentContainer.Image = container.Image
		current.Spec.Template.Spec.Containers[0] = currentContainer
		ds = *current
	}

	toUnstructured, err := pkgruntime.DefaultUnstructuredConverter.ToUnstructured(&ds)
	if err != nil {
		return err
	}
	obj.SetUnstructuredContent(toUnstructured)
	return common.ApplyResource(ctx, resourceInter, obj)
}

func KubeEdgeInstall(
	globalOpt *common.GlobalOption,
	opt *common.InitOption,
) aide.StepFunc {
	return func(sc *aide.StepContext) {
		switch opt.Type {
		case common.TypeContainer:
			sc.Log("Already apply cloudcore resource")
		case common.TypeBinary:
			// 创建默认配置
			cfg := edgeconfig.NewDefaultCloudCoreConfig(constants.RootDir, constants.ResourceDir)
			cfg.Modules.CloudHub.AdvertiseAddress = strings.Split(opt.AdvertiseAddress, ",")

			if globalOpt.KubeConfig != "" {
				cfg.KubeAPIConfig.KubeConfig = globalOpt.KubeConfig
			}
			if err := util.WriteToFile(filepath.Join(constants.ConfigDir, constants.CloudComponent+".yaml"), cfg); err != nil {
				sc.Errorf("Init Diplomat Cloud config failed: %v", err)
			}
			// 下载云端二进制执行程序
			// TODO 需要提供一个下载地址
			downloadCloudResource(sc, "https://github.com/kubeedge/kubeedge/releases/download/v1.11.2")

			// 检查是否存在systemd
			systemdExists := keutil.HasSystemd()
			// 创建systemd管理文件，并启动程序
			if systemdExists {
				sc.Log("Systemd found")

				if err := util.GenerateServiceFile(constants.ConfigDir, constants.ExecDir, constants.CloudComponent); err != nil {
					sc.Errorf("Init diplomat Cloud Systemd failed: %v", err)
				}
				if err := sc.Shell(fmt.Sprintf("sudo systemctl daemon-reload && sudo systemctl enable %s && sudo systemctl start %s",
					constants.CloudComponent, constants.CloudComponent)); err != nil {
					sc.Errorf("Init diplomat Cloud Systemd failed: %v", err)
				}
				sc.Log("Diplomat Cloud is running, For logs visit: journalctl -u cloudcore.service -b")
			} else {
				sc.Logl(aide.WarnLevel, "Systemd not found")

				logPath := filepath.Join(constants.LogDir, "cloudcore.log")
				if err := sc.Shell(fmt.Sprintf("%s > %s 2>&1 &",
					filepath.Join(constants.ExecDir, constants.CloudComponent), logPath)); err != nil {
					sc.Errorf("Init diplomat Cloud Systemd failed: %v", err)
				}
				sc.Log("Diplomat Cloud is running, For logs visit: ", logPath)
			}
		default:
			sc.Errorf("invalid install type %s", opt.Type)
		}
	}
}

func downloadCloudResource(sc *aide.StepContext, addr string) {
	// 适配amd64、arm环境的下载
	//edgeFileCleanName := "kubeedge-linux-" + runtime.GOARCH
	edgeFileCleanName := "kubeedge-v1.11.2-linux-" + runtime.GOARCH
	edgeFileName := edgeFileCleanName + ".tar.gz"

	// https://github.com/kubeedge/kubeedge/releases/download/v1.11.2 / kubeedge-v1.11.2-linux-amd64.tar.gz
	// 下载资源包
	if err := sc.Shell(fmt.Sprintf(
		"wget -k --no-check-certificate --progress=bar:force %s/%s -O %s",
		addr, edgeFileName, filepath.Join(constants.RootDir, edgeFileName))); err != nil {
		sc.Errorf("Download diplomat Cloud resource failed: %v", err)
	}

	// 解压并将执行程序移动到exec目录
	if err := sc.Shell(fmt.Sprintf(
		"cd %s && tar -C %s -xvzf %s && cp -f %s %s/",
		constants.RootDir, constants.RootDir, edgeFileName,
		filepath.Join(constants.RootDir, edgeFileCleanName, "cloud", "cloudcore", constants.CloudComponent), constants.ExecDir)); err != nil {
		sc.Errorf("Process diplomat Cloud executor file failed: %v", err)
	}
}

func HealthCheck(t string) aide.StepFunc {
	return func(sc *aide.StepContext) {
		var isRunning bool
		switch t {
		case common.TypeContainer:
			// TODO 检查容器部署的cloudcore
			isRunning = true
			sc.Log("please use `kubectl get pod -n kubeedge cloudcore' to check pod is it not healthy.")
		case common.TypeBinary:
			for i := 0; i < 5; i++ {
				cmd := exec.NewCommand("pidof cloudcore 2>&1")
				err := cmd.Exec()
				if cmd.ExitCode == 0 {
					isRunning = true
				} else if cmd.ExitCode == 1 {
					isRunning = false
				}
				if err != nil {
					isRunning = false
					sc.Log(aide.WarnLevel, err)
				}
				time.Sleep(time.Second * 2)
			}
		default:
			sc.Errorf("invalid install type %s", t)
		}

		if !isRunning {
			sc.Errorf("Service startup exception, please contact the administrator for troubleshooting")
		}
		sc.Log("Health is ok")
	}
}

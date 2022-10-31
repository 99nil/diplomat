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
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/common"
	"github.com/99nil/diplomat/pkg/edgeconfig"
	"github.com/99nil/diplomat/pkg/exec"
	"github.com/99nil/diplomat/pkg/hash"
	"github.com/99nil/diplomat/pkg/util"

	keutil "github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/util"
	"github.com/sirupsen/logrus"
	"github.com/zc2638/aide"
	"sigs.k8s.io/yaml"
)

func CheckEdge() aide.StepFunc {
	return func(sc *aide.StepContext) {
		if err := sc.Shell("docker version"); err != nil {
			sc.Errorf("Check docker failed: %v", err)
		}
		sc.Log("Check container runtime successful")

		running, err := isProcessRunning(constants.EdgeComponent)
		if err != nil {
			sc.Errorf("Check edge process failed: %v", err)
		}
		if running {
			sc.Errorf("Diplomat edgecore is already running on this node, please run reset to clean up first")
		}
		sc.Log("Check edge process successful")

		port := 1883
		format := "[ \"$(netstat -nl | awk '{if($4 ~ /^.*:%v/){print $0;}}')\" == '' ]"
		if err := sc.Bash(fmt.Sprintf(format, port)); err != nil {
			sc.Errorf("Check MQTT service port %d conflict, please check it. error: %v", port, err)
		}
		sc.Logf("Check MQTT service port %d successful", port)
	}
}

func Request(opt *common.JoinOption, imageSet map[string]string) aide.StepFunc {
	return func(sc *aide.StepContext) {
		images := make([]string, 0, len(imageSet))
		for _, v := range imageSet {
			images = append(images, v)
		}

		ctnRuntime, err := keutil.NewContainerRuntime(opt.RuntimeType, opt.RemoteRuntimeEndpoint)
		if err != nil {
			sc.Errorf("Create container runtime client failed, err: %v", err)
		}

		sc.Log("Pull Images")
		if err := ctnRuntime.PullImages(images); err != nil {
			sc.Errorf("Pull Images failed: %v", err)
		}

		sc.Log("Copy resources from the image to the management directory")
		dirs := map[string]string{
			constants.RootDir: filepath.Join(constants.TmpPath, "data"),
		}
		files := map[string]string{
			filepath.Join(constants.UsrBinPath, constants.EdgeComponent): filepath.Join(constants.TmpPath, "bin", constants.EdgeComponent),
		}
		if err := ctnRuntime.CopyResources(imageSet[common.InstallationPackageName], dirs, files); err != nil {
			sc.Errorf("copy resources failed: %v", err)
		}
	}
}

func StartMQTT(opt *common.JoinOption, imageSet map[string]string) aide.StepFunc {
	return func(sc *aide.StepContext) {
		if !opt.WithMQTT {
			sc.Log("set without MQTT tag, continue to init MQTT service")
			return
		}
		sc.Log("Start to init MQTT service")

		content := `persistence true
persistence_location /mosquitto/data
log_dest file /mosquitto/log/mosquitto.log
`
		dir := filepath.Join(constants.RootDir, "mosquitto", "config")
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			sc.Errorf("Create MQTT config dir %s failed: %s", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "mosquitto.conf"), []byte(content), os.ModePerm); err != nil {
			sc.Errorf("Create MQTT config file failed: %s", err)
		}

		image, ok := imageSet[constants.MQTTName]
		if !ok {
			image = "eclipse-mosquitto:1.6.15"
		}

		if err := sc.Shell(fmt.Sprintf(
			"docker run --name %s -d -p 1883:1883 -v /etc/diplomat/mosquitto:/mosquitto --restart unless-stopped %s",
			constants.MQTTName,
			image,
		)); err != nil {
			sc.Errorf("Start MQTT service failed: %s", err)
		}
		sc.Log("Start MQTT service successful")
	}
}

func JoinEdge(opt *common.JoinOption) aide.StepFunc {
	return func(sc *aide.StepContext) {
		// use 'installation-package' image to install Edge
		sc.Log("Generate EdgeCore default configuration")
		if err := createEdgeConfigFiles(opt); err != nil {
			sc.Errorf("Create edge config file failed: %v", err)
		}

		sc.Log("Generate systemd service file")
		if err := util.GenerateServiceFile(constants.ConfigDir, constants.UsrBinPath, constants.EdgeComponent); err != nil {
			sc.Errorf("Create systemd service file failed: %v", err)
		}

		sc.Log("Run EdgeCore daemon")
		if err := runEdgeCore(); err != nil {
			sc.Errorf("Start EdgeCore failed: %v", err)
		}
	}
}

func customEdgeConfig(opt *common.JoinOption, cfg *edgeconfig.EdgeCore) error {
	host, _, err := net.SplitHostPort(opt.CloudCoreIPPort)
	if err != nil {
		return err
	}
	cfg.Modules.EdgeHub.HTTPServer = "https://" + net.JoinHostPort(host, "10002")
	cfg.Modules.EdgeStream.TunnelServer = net.JoinHostPort(host, "10004")
	cfg.Modules.EdgeHub.Token = opt.Token
	cfg.Modules.EdgeHub.WebSocket.Server = opt.CloudCoreIPPort

	// 如果未定义节点名称，默认为自动注册
	if opt.EdgeNodeName != "" {
		cfg.Modules.Edged.HostnameOverride = opt.EdgeNodeName
		cfg.Modules.Edged.RegisterNode = false
		return nil
	}

	// 名称长度不能超过63个字符，mac地址默认占用18位，hash占用10位，限制34位
	hostnameOverride := cfg.Modules.Edged.HostnameOverride
	if len(hostnameOverride) > 34 {
		hostnameOverride = hostnameOverride[:33]
	}
	interfaces, err := net.Interfaces()
	var hardware string
	if err == nil {
		for _, inter := range interfaces {
			if inter.Name != "docker0" {
				continue
			}
			if inter.HardwareAddr.String() != "" {
				hardware = inter.HardwareAddr.String()
				break
			}
		}
	}
	if hardware != "" {
		hostnameOverride += "." + strings.ReplaceAll(hardware, ":", "-")
	}
	cfg.Modules.Edged.HostnameOverride = hash.BuildNodeName(opt.Namespace, hostnameOverride)
	cfg.Modules.Edged.Annotations = map[string]string{
		"diplomat.99nil.com/ether": hardware,
	}
	cfg.Modules.Edged.Labels = map[string]string{
		"diplomat.99nil.com/auto":      "node",
		"app-offline.kubeedge.io":      "autonomy",
		"diplomat.99nil.com/namespace": opt.Namespace,
		"diplomat.99nil.com/edge":      hostnameOverride,
	}
	for _, v := range opt.Labels {
		va := strings.Split(v, "=")
		if len(va) != 2 {
			continue
		}
		cfg.Modules.Edged.Labels[va[0]] = va[1]
	}
	return nil
}

func createEdgeConfigFiles(opt *common.JoinOption) error {
	var edgeCoreConfig *edgeconfig.EdgeCore

	configFilePath := filepath.Join(constants.ConfigDir, "edgecore.yaml")
	_, err := os.Stat(configFilePath)
	if err == nil || os.IsExist(err) {
		// Read existing configuration file
		b, err := os.ReadFile(configFilePath)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(b, &edgeCoreConfig); err != nil {
			return err
		}
	}
	if edgeCoreConfig == nil {
		// The configuration does not exist or the parsing fails, and the default configuration is generated
		edgeCoreConfig = edgeconfig.NewDefaultEdgeCoreConfig(constants.RootDir, constants.ResourceDir)
	}

	if err = customEdgeConfig(opt, edgeCoreConfig); err != nil {
		return fmt.Errorf("config edge custom configuration failed, err: %v", err)
	}

	edgeCoreConfig.Modules.EdgeHub.WebSocket.Server = opt.CloudCoreIPPort
	if opt.Token != "" {
		edgeCoreConfig.Modules.EdgeHub.Token = opt.Token
	}
	if opt.EdgeNodeName != "" {
		edgeCoreConfig.Modules.Edged.HostnameOverride = opt.EdgeNodeName
	}
	if opt.RuntimeType != "" {
		edgeCoreConfig.Modules.Edged.RuntimeType = opt.RuntimeType
	}

	switch opt.CGroupDriver {
	case constants.CGroupDriverSystemd:
		edgeCoreConfig.Modules.Edged.CGroupDriver = constants.CGroupDriverSystemd
	case constants.CGroupDriverCGroupFS:
		edgeCoreConfig.Modules.Edged.CGroupDriver = constants.CGroupDriverCGroupFS
	default:
		return fmt.Errorf("unsupported CGroupDriver: %s", opt.CGroupDriver)
	}
	edgeCoreConfig.Modules.Edged.CGroupDriver = opt.CGroupDriver

	if opt.RemoteRuntimeEndpoint != "" {
		edgeCoreConfig.Modules.Edged.RemoteRuntimeEndpoint = opt.RemoteRuntimeEndpoint
		edgeCoreConfig.Modules.Edged.RemoteImageEndpoint = opt.RemoteRuntimeEndpoint
	}

	host, _, err := net.SplitHostPort(opt.CloudCoreIPPort)
	if err != nil {
		return fmt.Errorf("get current host and port failed: %v", err)
	}
	if opt.CertPort != "" {
		edgeCoreConfig.Modules.EdgeHub.HTTPServer = "https://" + net.JoinHostPort(host, opt.CertPort)
	} else {
		edgeCoreConfig.Modules.EdgeHub.HTTPServer = "https://" + net.JoinHostPort(host, "10002")
	}
	edgeCoreConfig.Modules.EdgeStream.TunnelServer = net.JoinHostPort(host, "10004")

	if len(opt.Labels) > 0 {
		labelsMap := make(map[string]string)
		for _, label := range opt.Labels {
			arr := strings.SplitN(label, "=", 2)
			if arr[0] == "" {
				continue
			}

			if len(arr) > 1 {
				labelsMap[arr[0]] = arr[1]
			} else {
				labelsMap[arr[0]] = ""
			}
		}
		edgeCoreConfig.Modules.Edged.Labels = labelsMap
	}

	return util.Write2File(configFilePath, edgeCoreConfig)
}

func runEdgeCore() error {
	systemdExist := keutil.HasSystemd()

	var binExec, tip string
	if systemdExist {
		tip = fmt.Sprintf("KubeEdge edgecore is running, For logs visit: journalctl -u %s.service -xe", constants.EdgeComponent)
		binExec = fmt.Sprintf(
			"sudo systemctl daemon-reload && sudo systemctl enable %s && sudo systemctl start %s",
			constants.EdgeComponent, constants.EdgeComponent)
	} else {
		tip = fmt.Sprintf("KubeEdge edgecore is running, For logs visit: %s/%s.log", constants.LogDir, constants.EdgeComponent)
		binExec = fmt.Sprintf(
			"%s > %s/edge/%s.log 2>&1 &",
			filepath.Join(constants.UsrBinPath, constants.EdgeComponent),
			constants.LogDir,
			constants.EdgeComponent,
		)
	}

	cmd := exec.NewCommand(binExec)
	if err := cmd.Exec(); err != nil {
		return err
	}
	logrus.Infoln(cmd.GetStdOut())
	logrus.Infoln(tip)
	return nil
}

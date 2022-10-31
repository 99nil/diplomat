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

package common

import (
	"os"
	"strings"
	"sync"

	"github.com/99nil/diplomat/global/constants"
	"github.com/spf13/cobra"
)

var (
	AdmissionImage           = "docker.io/a526102465/admission:v1.11.2-diplomat"
	CloudcoreImage           = "docker.io/a526102465/cloudcore:v1.11.2-diplomat"
	InstallationPackageImage = "docker.io/a526102465/installation-package:v1.11.2-diplomat"
	RavenAgentImage          = "docker.io/a526102465/raven-agent:v0.0.1-diplomat"
	RavenControllerImage     = "docker.io/a526102465/raven-controller:v0.0.1-diplomat"
	FlannelImage             = "docker.io/a526102465/flannel:v0.14.0-diplomat"
)

const (
	AdmissionName           = "admission"
	CloudcoreName           = "cloudcore"
	InstallationPackageName = "installationpackage"
	RavenAgentName          = "ravenagent"
	RavenControllerName     = "ravencontroller"
	FlannelName             = "flannel"
)

func NewCloudImageSet() map[string]string {
	// 以static下的yaml内的版本为主，这里不要添加默认的
	// 否则将会以imageSet中的版本为优先
	imageSet := make(map[string]string)
	imageSet[AdmissionName] = AdmissionImage
	imageSet[CloudcoreName] = CloudcoreImage
	imageSet[RavenAgentName] = RavenAgentImage
	imageSet[RavenControllerName] = RavenControllerImage
	imageSet[FlannelName] = FlannelImage
	return imageSet
}

func NewEdgeImageSet() map[string]string {
	// 以static下的yaml内的版本为主，这里不要添加默认的
	// 否则将会以imageSet中的版本为优先
	imageSet := make(map[string]string)
	imageSet[InstallationPackageName] = InstallationPackageImage
	return imageSet
}

type PortSet struct {
	set sync.Map
}

func (s *PortSet) Has(name string) bool {
	_, ok := s.set.Load(name)
	return ok
}

func (s *PortSet) Get(name string) int32 {
	val, ok := s.set.Load(name)
	if ok {
		return val.(int32)
	}
	return 0
}

func (s *PortSet) Add(name string, port int32) {
	s.set.Store(name, port)
}

func (s *PortSet) Remove(name string) {
	s.set.Delete(name)
}

func (s *PortSet) Range(f func(name string, port int32) bool) {
	s.set.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(int32))
	})
}

func (s *PortSet) Reset() {
	s.set = sync.Map{}
}

func NewPortOption() *PortSet {
	ps := new(PortSet)
	//ps.Add(BackendName, 31181)
	//ps.Add(FrontendName, 31180)
	//ps.Add(MysqlName, 31186)
	//ps.Add(KeycloakName, 31187)
	//ps.Add(ManualName, 31189)
	//
	//ps.Add(PrometheusOperator, 31190)
	//ps.Add(Prometheus, 31191)
	//ps.Add(Alertmanager, 31192)
	//
	//ps.Add(LokiName, 31300)
	//ps.Add(GrafanaName, 31301)
	return ps
}

type GlobalOption struct {
	Module    string
	EnvPrefix string
	Config    string
	// 显示详细
	Verbose bool
	// k8s配置文件路径
	KubeConfig string
}

func (o *GlobalOption) CompleteFlags(cmd *cobra.Command) {
	cfgPathEnv := os.Getenv(o.EnvPrefix + "_CONFIG")
	if cfgPathEnv == "" {
		cfgPathEnv = "config/config.yaml"
	}
	cmd.Flags().StringVarP(&o.Config, "config", "c", cfgPathEnv,
		"config file (default is $HOME/config.yaml)")
}

func NewGlobalOption(module string) *GlobalOption {
	opt := &GlobalOption{Module: module}
	module = strings.ReplaceAll(module, "-", "_")
	opt.EnvPrefix = strings.ToUpper(constants.ProjectName + "_" + module)
	return opt
}

const (
	EnvDev  = "dev"
	EnvProd = "prod"
)

const (
	TypeBinary    = "binary"
	TypeContainer = "container"
)

type InitOption struct {
	// 安装类型：二进制binary、容器部署container
	Type string
	// 服务端可用的IP地址，如果有多个，则以逗号分隔，主要用于证书签名
	AdvertiseAddress string
	// 服务端可用的域名，如果有多个，则以逗号分隔，主要用于证书签名
	Domain string
	// k8s的证书路径，用于(cloud-stream)日志流服务的证书签名
	K8sCertPath string
	// 环境: dev、prod
	Env string
	// 需要被禁用的组件，如果有多个则以逗号分隔，例如 cloudStream,deviceController
	Disable []string
	// 指定镜像仓库 e.g. k8s.gcr.io  docker.io/99nil
	ImageRepository string
	// 指定镜像仓库用户名称
	ImageRepositoryUsername string
	// 指定镜像仓库用户密码
	ImageRepositoryPassword string
	// 安装过程使用的正确IP，从advertiseAddress解析出第一个，如果没有则读取hostIP
	CurrentIP string
}

type JoinOption struct {
	CloudCoreIPPort       string
	EdgeNodeName          string
	Namespace             string
	CGroupDriver          string
	CertPath              string
	RuntimeType           string
	RemoteRuntimeEndpoint string
	Token                 string
	CertPort              string
	WithMQTT              bool
	ImageRepository       string
	Labels                []string
}

type ResetOption struct {
	Force bool
}

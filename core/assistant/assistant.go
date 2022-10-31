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

	"github.com/99nil/diplomat/core/component"
	"github.com/99nil/diplomat/pkg/common"

	"github.com/zc2638/aide"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func Run(cfg *Config) error {
	// TODO 支持插件的升级及重启
	// TODO 支持自升级
	// TODO 监听插件的健康状态
	return nil
}

func NewInitInstance(
	globalOpt *common.GlobalOption,
	opt *common.InitOption,
	cfg *Config,
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
) *aide.Instance {
	ri := component.RavenInstallTool{
		Ctx:           context.Background(),
		Resources:     cfg.RavenResources,
		KubeClient:    kubeClient,
		DynamicClient: dynamicClient,
	}
	imageSet := common.NewCloudImageSet()
	portSet := common.NewPortOption()

	ins := aide.New(aide.WithVerboseOption(globalOpt.Verbose))
	checkStage := aide.NewStage("check")
	{
		checkStage.AddStepFunc("port check", CheckPort(append(cfg.KubeEdgeResources, cfg.RavenResources...), opt, portSet))
		checkStage.AddStepFunc("k8s check", CheckK8s())
		checkStage.AddStepFunc("cloud check", CheckCloud(kubeClient))
		checkStage.AddStepFunc("raven check", func(sc *aide.StepContext) {
			if err := ri.PreInstall(ri.Ctx); err != nil {
				sc.Errorf("Check raven failed, err: %v", err)
			}
		})
	}

	initStage := aide.NewStage("init resource")
	{
		initStage.AddStepFunc("prepare in advance", Advance())
		initStage.AddStepFunc("namespace in advance", AdvanceNS(kubeClient))
		initStage.AddStepFunc("secret in advance", AdvanceSecret(kubeClient, opt, imageSet))
		initStage.AddStepFunc("generate certs", GenerateCerts(opt))
		initStage.AddStepFunc("apply kubeedge resource", ApplyResource(globalOpt, opt, cfg.KubeEdgeResources, kubeClient, dynamicClient, portSet, imageSet))
		initStage.AddStepFunc("kubeedge install", KubeEdgeInstall(globalOpt, opt))
		initStage.AddStepFunc("raven install", func(sc *aide.StepContext) {
			if err := ri.Install(ri.Ctx); err != nil {
				sc.Errorf("install raven failed, err: %v", err)
			}
		})
	}

	healthStage := aide.NewStage("health check")
	{
		healthStage.AddStepFunc("cloud health", HealthCheck(opt.Type))
	}

	ins.AddStages(checkStage, initStage, healthStage)

	return ins
}

func NewJoinInstance(
	globalOpt *common.GlobalOption,
	opt *common.JoinOption,
) *aide.Instance {
	imageSet := common.NewEdgeImageSet()
	ins := aide.New(aide.WithVerboseOption(globalOpt.Verbose))

	checkStage := aide.NewStage("check").AddSteps(
		CheckEdge().Step("check dependency"),
	)
	joinStage := aide.NewStage("join").AddSteps(
		Advance().Step("prepare in advance"),
		Request(opt, imageSet).Step("prepare images"),
		StartMQTT(opt, imageSet).Step("start MQTT"),
		JoinEdge(opt).Step("join cloud"),
	)

	ins.AddStages(checkStage, joinStage)
	return ins
}

func NewUpgradeInstance(
	globalOpt *common.GlobalOption,
	opt *common.InitOption,
	cfg *Config,
	kubeClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
) *aide.Instance {
	// TODO cloudcore, raven, edgecore...
	return nil
}

func NewCloudResetInstance(
	globalOpt *common.GlobalOption,
	opt *common.ResetOption,
	kubeClient kubernetes.Interface,
) *aide.Instance {
	ins := aide.New(aide.WithVerboseOption(globalOpt.Verbose))
	if !opt.Force {
		chooseStage := aide.NewStage("choose")
		chooseStage.AddStepFunc("reset confirm", ChooseResetConfirm(false))
		ins.AddStages(chooseStage)
	}

	resetStage := aide.NewStage("reset")
	resetStage.AddStepFunc("reset cloudcore", ResetCloud(kubeClient))
	ins.AddStages(resetStage)

	return ins
}

func NewEdgeResetInstance(
	globalOpt *common.GlobalOption,
	opt *common.ResetOption,
) *aide.Instance {
	ins := aide.New(aide.WithVerboseOption(globalOpt.Verbose))
	if !opt.Force {
		chooseStage := aide.NewStage("choose")
		chooseStage.AddStepFunc("reset confirm", ChooseResetConfirm(true))
		ins.AddStages(chooseStage)
	}

	resetStage := aide.NewStage("reset")
	resetStage.AddStepFunc("reset edgecore", ResetEdge())
	ins.AddStages(resetStage)

	return ins
}

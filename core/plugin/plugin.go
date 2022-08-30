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

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/shirou/gopsutil/process"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

type Interface interface {
	Name() string
	Kind() string
	Options() []Option
}

type Stats struct {
	Name        string  `json:"name"`
	CPUUsage    uint64  `json:"cpu_usage"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryStats uint64  `json:"memory_stats"`
}

type Engine interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Stats(ctx context.Context) ([]Stats, error)
}

type Option struct {
	Name string
	Args []string

	// docker options
	Image string
}

func NewDockerEngine(client dockerclient.APIClient, opts []Option) Engine {
	return &DockerEngine{client: client, opts: opts}
}

type DockerEngine struct {
	client dockerclient.APIClient
	opts   []Option
}

func (e *DockerEngine) Name() string {
	return "docker"
}

func (e *DockerEngine) Start(ctx context.Context) error {
	hostCfg := &dockercontainer.HostConfig{
		NetworkMode:   dockercontainer.NetworkMode("host"),
		RestartPolicy: dockercontainer.RestartPolicy{Name: "unless-stopped"},
	}

	for _, opt := range e.opts {
		cfg := &dockercontainer.Config{
			Hostname: opt.Name,
			Image:    opt.Image,
			Labels:   nil,
		}

		// TODO 判断镜像是否存在，不存在拉取

		if _, err := e.client.ContainerCreate(ctx, cfg, hostCfg, nil, nil, opt.Name); err != nil {
			return fmt.Errorf("container(%s) create failed: %v", opt.Name, err)
		}
		if err := e.client.ContainerStart(ctx, opt.Name, dockertypes.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("container(%s) start failed: %v", opt.Name, err)
		}
	}
	return nil
}

func (e *DockerEngine) Stop(ctx context.Context) error {
	for _, opt := range e.opts {
		if err := e.client.ContainerStop(ctx, opt.Name, nil); err != nil {
			return fmt.Errorf("container(%s) stop failed: %v", opt.Name, err)
		}
	}
	return nil
}

func (e *DockerEngine) Stats(ctx context.Context) ([]Stats, error) {
	statsSet := make([]Stats, 0, len(e.opts))
	for _, opt := range e.opts {
		resp, err := e.client.ContainerStats(ctx, opt.Name, false)
		if err != nil {
			return nil, fmt.Errorf("container stats failed: %v", err)
		}

		var stats dockertypes.Stats
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return nil, fmt.Errorf("container stats decode failed: %v", err)
		}
		resp.Body.Close()

		statsSet = append(statsSet, Stats{
			Name:        opt.Name,
			CPUUsage:    stats.CPUStats.CPUUsage.TotalUsage,
			MemoryStats: stats.MemoryStats.Usage,
		})
	}
	return statsSet, nil
}

func NewProcessEngine(opts []Option) Engine {
	return &ProcessEngine{opts: opts}
}

type ProcessEngine struct {
	opts []Option
	cmds []*exec.Cmd
}

func (e *ProcessEngine) Name() string {
	return "process"
}

func (e *ProcessEngine) Start(ctx context.Context) error {
	for _, opt := range e.opts {
		cmd := exec.Command(opt.Name, opt.Args...)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start process(%s) failed: %v", opt.Name, err)
		}
		e.cmds = append(e.cmds, cmd)
	}
	return nil
}

func (e *ProcessEngine) Stop(ctx context.Context) error {
	for _, cmd := range e.cmds {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("stop process(%d) failed: %v", cmd.Process.Pid, err)
		}
	}
	return nil
}

func (e *ProcessEngine) Stats(ctx context.Context) ([]Stats, error) {
	statsSet := make([]Stats, 0, len(e.cmds))
	for _, cmd := range e.cmds {
		p, err := process.NewProcessWithContext(ctx, int32(cmd.Process.Pid))
		if err != nil {
			return nil, err
		}
		memInfo, err := p.MemoryInfo()
		if err != nil {
			return nil, err
		}
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			return nil, err
		}

		statsSet = append(statsSet, Stats{
			CPUPercent:  cpuPercent,
			MemoryStats: memInfo.Data,
		})
	}
	return statsSet, nil
}

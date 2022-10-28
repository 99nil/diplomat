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

package edgeconfig

import (
	"context"
	"fmt"
	"path/filepath"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"k8s.io/klog/v2"
)

// CopyResources copies binary and configuration file from the image to the host.
// The same way as func (runtime *DockerRuntime) CopyResources
func CopyResources(ctx context.Context, image string, dirs map[string]string, files map[string]string) error {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return fmt.Errorf("init docker client failed: %v", err)
	}

	cli.NegotiateAPIVersion(ctx)

	if len(files) == 0 && len(dirs) == 0 {
		return fmt.Errorf("no resources need copying")
	}

	copyCmd := copyResourcesCmd(dirs, files)

	config := &dockercontainer.Config{
		Image: image,
		Cmd: []string{
			"/bin/sh",
			"-c",
			copyCmd,
		},
	}
	var binds []string
	for origin, bind := range dirs {
		binds = append(binds, origin+":"+bind)
	}
	for origin, bind := range files {
		binds = append(binds, filepath.Dir(origin)+":"+filepath.Dir(bind))
	}

	hostConfig := &dockercontainer.HostConfig{
		Binds: binds,
	}

	// Randomly generate container names to prevent duplicate names.
	container, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	defer func() {
		if err := cli.ContainerRemove(ctx, container.ID, dockertypes.ContainerRemoveOptions{}); err != nil {
			klog.V(3).ErrorS(err, "Remove container failed", "containerID", container.ID)
		}
	}()

	if err := cli.ContainerStart(ctx, container.ID, dockertypes.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("container start failed: %v", err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, container.ID, "")
	select {
	case err := <-errCh:
		klog.Errorf("container wait error %v", err)
	case <-statusCh:
	}
	return nil
}

func copyResourcesCmd(dirs map[string]string, files map[string]string) string {
	var copyCmd string
	first := true
	for origin, bind := range dirs {
		if first {
			copyCmd = copyCmd + fmt.Sprintf("cp -r %s %s", origin, bind)
		} else {
			copyCmd = copyCmd + fmt.Sprintf(" && cp -r %s %s", origin, bind)
		}
		first = false
	}
	for origin, bind := range files {
		if first {
			copyCmd = copyCmd + fmt.Sprintf("cp %s %s", origin, bind)
		} else {
			copyCmd = copyCmd + fmt.Sprintf(" && cp %s %s", origin, bind)
		}
		first = false
	}
	return copyCmd
}

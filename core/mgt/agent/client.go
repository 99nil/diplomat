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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/99nil/diplomat/global/constants"
	"github.com/99nil/diplomat/pkg/logr"
	"github.com/99nil/diplomat/pkg/sse"
	"github.com/99nil/dsync/suid"
)

func NewClient(host, nodeName string) *Client {
	host = strings.TrimSuffix(host, "/")
	client := &http.Client{}
	return &Client{client: client, host: host, node: nodeName}
}

type Client struct {
	client   *http.Client
	host     string
	node     string
	instance string
}

func (c *Client) Manifest(ctx context.Context, state suid.UID) (*suid.AssembleManifest, error) {
	uri := c.host + "/api/v1/manifest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header = make(http.Header)
	req.Header.Set("node", c.node)
	req.Header.Set("state", state.String())

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	instance := res.Header.Get(fmt.Sprintf("%s-mgt-server-instance", constants.ProjectName))
	logr.Debugf("Request manifest from mgt-server-instance: %s", instance)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var manifest *suid.AssembleManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}
	c.instance = instance
	return manifest, nil
}

func (c *Client) Data(ctx context.Context, manifest *suid.AssembleManifest, fn func(msg *sse.Message) error) error {
	b, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	uri := c.host + "/api/v1/data"
	req, err := sse.NewRequest(uri)
	if err != nil {
		return err
	}
	req.Header.Set("node", c.node)
	req.Header.Set(fmt.Sprintf("%s-mgt-server-instance", constants.ProjectName), c.instance)
	req.Header.Set("manifest", string(b))

	res, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	reader := sse.NewEventStreamReader(res.Body, 1024*1024)
	for {
		content, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		message, err := sse.TransferMessage(content)
		if err != nil {
			return err
		}
		if *message == *sse.EOFMessage {
			return nil
		}
		if message.Event == "error" {
			return errors.New(message.Data)
		}
		if err := fn(message); err != nil {
			return err
		}
	}
}

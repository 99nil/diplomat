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

package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/yaml"
)

func ParseResourceVersion(resourceVersion string) uint64 {
	if resourceVersion == "" || resourceVersion == "0" {
		return 0
	}
	version, _ := strconv.ParseUint(resourceVersion, 10, 64)
	return version
}

// CompareResourceVersion returns an integer comparing two resource version strings.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
func CompareResourceVersion(a, b string) int {
	an := ParseResourceVersion(a)
	bn := ParseResourceVersion(b)
	if an == bn {
		return 0
	}
	if an < bn {
		return -1
	}
	return +1
}

func WriteToFile(path string, data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}

func GetLocalIP(hostName string) (string, error) {
	var ipAddr net.IP
	var err error

	// If looks up host failed, will use utilnet.ChooseHostInterface() below,
	// So ignore the error here
	addrs, _ := net.LookupIP(hostName)
	for _, addr := range addrs {
		if err := ValidateNodeIP(addr); err != nil {
			continue
		}
		if addr.To4() != nil {
			ipAddr = addr
			break
		}
		if ipAddr == nil && addr.To16() != nil {
			ipAddr = addr
		}
	}

	if ipAddr == nil {
		ipAddr, err = utilnet.ChooseHostInterface()
		if err != nil {
			return "", err
		}
	}
	return ipAddr.String(), nil
}

// ValidateNodeIP validates given node IP belongs to the current host
func ValidateNodeIP(nodeIP net.IP) error {
	// Honor IP limitations set in setNodeStatus()
	if nodeIP.To4() == nil && nodeIP.To16() == nil {
		return fmt.Errorf("nodeIP must be a valid IP address")
	}
	if nodeIP.IsLoopback() {
		return fmt.Errorf("nodeIP can't be loopback address")
	}
	if nodeIP.IsMulticast() {
		return fmt.Errorf("nodeIP can't be a multicast address")
	}
	if nodeIP.IsLinkLocalUnicast() {
		return fmt.Errorf("nodeIP can't be a link-local unicast address")
	}
	if nodeIP.IsUnspecified() {
		return fmt.Errorf("nodeIP can't be an all zeros address")
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && ip.Equal(nodeIP) {
			return nil
		}
	}
	return fmt.Errorf("node IP: %q not found in the host's network interfaces", nodeIP.String())
}

func InStringSlice(ss []string, str string) (index int, exists bool) {
	for k, v := range ss {
		if str == v {
			return k, true
		}
	}
	return -1, false
}

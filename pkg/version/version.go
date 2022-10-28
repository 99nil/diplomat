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

package version

import (
	"fmt"
	"runtime"
)

var (
	ver          string
	assistantVer string
	kubeedgeVer  string
	ravenVer     string
)

type Version struct {
	Version          string
	AssistantVersion string
	KubeedgeVersion  string
	RavenVersion     string
	GoVersion        string
	Compiler         string
	Platform         string
}

func Get() Version {
	return Version{
		Version:          ver,
		AssistantVersion: assistantVer,
		KubeedgeVersion:  kubeedgeVer,
		RavenVersion:     ravenVer,
		GoVersion:        runtime.Version(),
		Compiler:         runtime.Compiler,
		Platform:         fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func (v Version) String() string {
	return fmt.Sprintf(`version: %s
assistant: %s
kubeedge: %s
raven: %s
go version: %s
go compiler: %s
platform: %s
`,
		v.Version,
		v.AssistantVersion,
		v.KubeedgeVersion,
		v.RavenVersion,
		v.GoVersion,
		v.Compiler,
		v.Platform,
	)
}

func (v Version) EdgeString() string {
	if v.Version == "" {
		v.Version = "dev"
	}

	return fmt.Sprintf(`version: %s
kubeedge: %s
go version: %s
go compiler: %s
platform: %s
`,
		v.AssistantVersion,
		v.KubeedgeVersion,
		v.GoVersion,
		v.Compiler,
		v.Platform,
	)
}

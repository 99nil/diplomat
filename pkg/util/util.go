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
	"strconv"
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

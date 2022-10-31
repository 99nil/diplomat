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

package hash

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

const alphaNums = "bcdfghjklmnpqrstvwxz2456789"

func SafeEncodeString(s string) string {
	r := make([]byte, len(s))
	for i, b := range []rune(s) {
		r[i] = alphaNums[(int(b) % len(alphaNums))]
	}
	return string(r)
}

func ComputeHash(data interface{}) string {
	h := fnv.New32a()
	fmt.Fprintf(h, "%#v", data)
	return SafeEncodeString(strconv.FormatUint(uint64(h.Sum32()), 10))
}

func BuildNodeName(namespace, name string) string {
	return name + "-" + ComputeHash(namespace)
}

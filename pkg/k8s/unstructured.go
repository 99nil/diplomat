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

package k8s

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/99nil/diplomat/static"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	serializeryaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

var serializer = serializeryaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func ParseAllYamlToObject(dirPath string) ([][]unstructured.Unstructured, error) {
	files, err := static.EmbedResource.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("read resource dir(%s) failed: %v", dirPath, err)
	}

	var objs [][]unstructured.Unstructured
	for _, file := range files {
		name := file.Name()
		currentPath := filepath.Join(dirPath, name)

		if file.IsDir() {
			current, err := ParseAllYamlToObject(currentPath)
			if err != nil {
				return nil, err
			}
			objs = append(objs, current...)
			continue
		}

		ext := filepath.Ext(name)
		isYamlFile := ext == ".yaml" || ext == ".yml"
		if !isYamlFile {
			continue
		}
		result, err := parseUnstructuredObject(currentPath)
		if err != nil {
			return nil, err
		}
		objs = append(objs, result)
	}
	return objs, nil
}

func parseUnstructuredObject(filePath string) ([]unstructured.Unstructured, error) {
	data, err := static.EmbedResource.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed: %v", filePath, err)
	}
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 256)

	var result []unstructured.Unstructured
	for {
		var rawObj pkgruntime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode data to raw object failed: %v", err)
		}

		var obj unstructured.Unstructured
		if err := pkgruntime.DecodeInto(serializer, rawObj.Raw, &obj); err != nil {
			return nil, fmt.Errorf("decode raw object failed: %v", err)
		}
		result = append(result, obj)
	}
	return result, nil
}

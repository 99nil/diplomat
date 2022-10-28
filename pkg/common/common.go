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
	"context"
	"fmt"
	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func IsSpecialName(name string) bool {
	// TODO 指定svc暴露
	return true
}

func ApplyResource(
	ctx context.Context,
	resourceInter dynamic.ResourceInterface,
	obj *unstructured.Unstructured,
) error {
	current, err := resourceInter.Get(ctx, obj.GetName(), metaV1.GetOptions{
		TypeMeta: metaV1.TypeMeta{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
		},
	})
	if err == nil {
		rv, _ := strconv.ParseInt(current.GetResourceVersion(), 10, 64)
		obj.SetResourceVersion(strconv.FormatInt(rv, 10))
		if _, err = resourceInter.Update(ctx, obj, metaV1.UpdateOptions{}); err != nil {
			err = fmt.Errorf("update %s %s failed: %v", obj.GetKind(), obj.GetName(), err)
		}
		return err
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	if _, err = resourceInter.Create(ctx, obj, metaV1.CreateOptions{}); err != nil {
		err = fmt.Errorf("create %s %s failed: %v", obj.GetKind(), obj.GetName(), err)
	}
	return err
}

/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"encoding/json"
	"fmt"

	api "kubevault.dev/operator/apis/engine/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned/typed/engine/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchMySQLRole(c cs.EngineV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.MySQLRole) *api.MySQLRole) (*api.MySQLRole, kutil.VerbType, error) {
	cur, err := c.MySQLRoles(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating MySQLRole %s/%s.", meta.Namespace, meta.Name)
		out, err := c.MySQLRoles(meta.Namespace).Create(transform(&api.MySQLRole{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindMySQLRole,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchMySQLRole(c, cur, transform)
}

func PatchMySQLRole(c cs.EngineV1alpha1Interface, cur *api.MySQLRole, transform func(*api.MySQLRole) *api.MySQLRole) (*api.MySQLRole, kutil.VerbType, error) {
	return PatchMySQLRoleObject(c, cur, transform(cur.DeepCopy()))
}

func PatchMySQLRoleObject(c cs.EngineV1alpha1Interface, cur, mod *api.MySQLRole) (*api.MySQLRole, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching MySQLRole %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.MySQLRoles(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateMySQLRole(c cs.EngineV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.MySQLRole) *api.MySQLRole) (result *api.MySQLRole, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.MySQLRoles(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.MySQLRoles(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update MySQLRole %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update MySQLRole %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateMySQLRoleStatus(
	c cs.EngineV1alpha1Interface,
	in *api.MySQLRole,
	transform func(*api.MySQLRoleStatus) *api.MySQLRoleStatus,
) (result *api.MySQLRole, err error) {
	apply := func(x *api.MySQLRole) *api.MySQLRole {
		return &api.MySQLRole{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(in.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur := in.DeepCopy()
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.MySQLRoles(in.Namespace).UpdateStatus(apply(cur))
		if kerr.IsConflict(e2) {
			latest, e3 := c.MySQLRoles(in.Namespace).Get(in.Name, metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update status of MySQLRole %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
	}
	return
}

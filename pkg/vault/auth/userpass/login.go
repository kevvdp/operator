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
package userpass

import (
	"encoding/json"
	"fmt"

	"kubevault.dev/operator/apis"
	vsapi "kubevault.dev/operator/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/auth/types"
	vaultuitl "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type auth struct {
	vClient *vaultapi.Client
	user    string
	pass    string
	path    string
}

func New(vApp *appcat.AppBinding, secret *core.Secret) (*auth, error) {
	cfg, err := vaultuitl.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client")
	}

	user, ok := secret.Data[core.BasicAuthUsernameKey]
	if !ok {
		return nil, errors.New("username is missing")
	}
	pass, ok := secret.Data[core.BasicAuthPasswordKey]
	if !ok {
		return nil, errors.New("password is missing")
	}

	authPath := string(vsapi.AuthTypeUserPass)
	if val, ok := secret.Annotations[apis.AuthPathKey]; ok && len(val) > 0 {
		authPath = val
	}

	return &auth{
		vClient: vc,
		user:    string(user),
		pass:    string(pass),
		path:    authPath,
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login/%s", a.path, a.user)
	req := a.vClient.NewRequest("POST", path)
	payload := map[string]interface{}{
		"password": a.pass,
	}
	if err := req.SetJSONBody(payload); err != nil {
		return "", err
	}

	resp, err := a.vClient.RawRequest(req)
	if err != nil {
		return "", err
	}

	var loginResp types.AuthLoginResponse
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}
	return loginResp.Auth.ClientToken, nil
}

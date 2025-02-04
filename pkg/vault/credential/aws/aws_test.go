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
package aws

import (
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestAWSCredManager_ParseCredential(t *testing.T) {
	awsCM := &AWSCredManager{}

	testData := []struct {
		name      string
		data      map[string]interface{}
		expectErr bool
	}{
		{
			name: "success, 'security_key' is nil",
			data: map[string]interface{}{
				"access_key":   "hi",
				"secret_key":   "hello",
				"security_key": nil,
			},
			expectErr: false,
		},
		{
			name: "success, 'security_key' is string",
			data: map[string]interface{}{
				"access_key":   "hi",
				"secret_key":   "hello",
				"security_key": "bye",
			},
			expectErr: false,
		},
		{
			name: "failed, 'security_key' is struct",
			data: map[string]interface{}{
				"access_key": "hi",
				"secret_key": "hello",
				"security_key": struct {
				}{},
			},
			expectErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			data, err := awsCM.ParseCredential(&vaultapi.Secret{
				Data: test.data,
			})
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {
					for key, val := range test.data {
						if val == nil {
							assert.Nil(t, data[key])
						} else {
							assert.Equal(t, val.(string), string(data[key]))
						}
					}
				}
			}
		})
	}
}

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
package v1alpha1

import (
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindGCPAccessKeyRequest = "GCPAccessKeyRequest"
	ResourceGCPAccessKeyRequest     = "gcpaccesskeyrequest"
	ResourceGCPAccessKeyRequests    = "gcpaccesskeyrequests"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=gcpaccesskeyrequests,singular=gcpaccesskeyrequest,categories={vault,appscode,all}
// +kubebuilder:subresource:status
type GCPAccessKeyRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GCPAccessKeyRequestSpec   `json:"spec,omitempty"`
	Status            GCPAccessKeyRequestStatus `json:"status,omitempty"`
}

// Link:
//  - https://www.vaultproject.io/api/secret/gcp/index.html#generate-secret-iam-service-account-creds-oauth2-access-token
//  - https://www.vaultproject.io/api/secret/gcp/index.html#generate-secret-iam-service-account-creds-service-account-key

// GCPAccessKeyRequestSpec contains information to request for vault gcp credentials
type GCPAccessKeyRequestSpec struct {
	// Contains vault gcp role info
	// +required
	RoleRef RoleRef `json:"roleRef"`

	// Contains a reference to the object or user identities the role binding is applied to
	// +required
	Subjects []rbac.Subject `json:"subjects"`

	// Specifies the algorithm used to generate key.
	// Defaults to 2k RSA key.
	// Accepted values: KEY_ALG_UNSPECIFIED, KEY_ALG_RSA_1024, KEY_ALG_RSA_2048
	// +optional
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`

	// Specifies the private key type to generate.
	// Defaults to JSON credentials file
	// Accepted values: TYPE_UNSPECIFIED, TYPE_PKCS12_FILE, TYPE_GOOGLE_CREDENTIALS_FILE
	// +optional
	KeyType string `json:"keyType,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type GCPAccessKeyRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of GCPAccessKeyRequest objects
	Items []GCPAccessKeyRequest `json:"items,omitempty"`
}

type GCPAccessKeyRequestStatus struct {
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []GCPAccessKeyRequestCondition `json:"conditions,omitempty"`

	// Name of the secret containing GCPCredential
	Secret *core.LocalObjectReference `json:"secret,omitempty"`

	// Contains lease info
	Lease *Lease `json:"lease,omitempty"`
}

type GCPAccessKeyRequestCondition struct {
	// request approval state, currently Approved or Denied.
	Type RequestConditionType `json:"type"`

	// brief reason for the request state
	// +optional
	Reason string `json:"reason,omitempty"`

	// human readable message with details about the request state
	// +optional
	Message string `json:"message,omitempty"`

	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	ResourceKindMySQLRole = "MySQLRole"
	ResourceMySQLRole     = "mysqlrole"
	ResourceMySQLRoles    = "mysqlroles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mysqlroles,singular=mysqlrole,categories={vault,appscode,all}
// +kubebuilder:subresource:status
type MySQLRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MySQLRoleSpec   `json:"spec,omitempty"`
	Status            MySQLRoleStatus `json:"status,omitempty"`
}

// MySQLRoleSpec contains connection information, mysql role info etc
type MySQLRoleSpec struct {
	// VaultRef is the name of a AppBinding referencing to a Vault Server
	VaultRef core.LocalObjectReference `json:"vaultRef"`

	// DatabaseRef specifies the database appbinding reference in any namespace.
	DatabaseRef *appcat.AppReference `json:"databaseRef,omitempty"`

	// Specifies the database name under which the role will be created
	DatabaseName string `json:"databaseName,omitempty"`

	// Specifies the path where secret engine is enabled
	Path string `json:"path,omitempty"`

	// links:
	// 	- https://www.vaultproject.io/api/secret/databases/index.html
	//	- https://www.vaultproject.io/api/secret/databases/mysql-maria.html

	// Specifies the TTL for the leases associated with this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds.
	// Defaults to system/engine default TTL time
	DefaultTTL string `json:"defaultTTL,omitempty"`

	// Specifies the maximum TTL for the leases associated with this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds.
	// Defaults to system/engine default TTL time.
	MaxTTL string `json:"maxTTL,omitempty"`

	// https://www.vaultproject.io/api/secret/databases/mysql-maria.html#creation_statements
	// Specifies the database statements executed to create and configure a user.
	CreationStatements []string `json:"creationStatements"`

	// https://www.vaultproject.io/api/secret/databases/mysql-maria.html#revocation_statements
	// Specifies the database statements to be executed to revoke a user.
	RevocationStatements []string `json:"revocationStatements,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MySQLRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of MySQLRole objects
	Items []MySQLRole `json:"items,omitempty"`
}

type MySQLRolePhase string

type MySQLRoleStatus struct {
	Phase MySQLRolePhase `json:"phase,omitempty"`

	// ObservedGeneration is the most recent generation observed for this MySQLRole. It corresponds to the
	// MySQLRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a MySQLRole current state.
	Conditions []MySQLRoleCondition `json:"conditions,omitempty"`
}

// MySQLRoleCondition describes the state of a MySQLRole at a certain point.
type MySQLRoleCondition struct {
	// Type of MySQLRole condition.
	Type string `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

/*
Copyright 2023.

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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TenantAccessSpec defines the desired state of TenantRBAC
type TenantAccessSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Iam is a list of IAM permissions
	Iam []IamSpec `json:"iam,omitempty"`
}

type IamSpec struct {
	Name     string              `json:"name"`
	Rules    []rbacv1.PolicyRule `json:"rules"`
	Subjects []rbacv1.Subject    `json:"subjects,omitempty"`
}

// TenantAccessStatus defines the observed state of TenantAccess
type TenantAccessStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TenantAccess is the Schema for the tenantaccess API
type TenantAccess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantAccessSpec   `json:"spec,omitempty"`
	Status TenantAccessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantAccessList contains a list of TenantAccess
type TenantAccessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantAccess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantAccess{}, &TenantAccessList{})
}

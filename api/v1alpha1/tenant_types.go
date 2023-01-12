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

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Assign subjects to the admin rbac role
	Admins []rbacv1.Subject `json:"admins,omitempty"`
	// Assign subjects to the viewer rbac role
	Viewers []rbacv1.Subject `json:"viewers,omitempty"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Namespace NamespaceStatus `json:"namespace"`
	Admins    string          `json:"admins,omitempty"`
	Viewers   string          `json:"viewers,omitempty"`
}

type NamespaceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:JSONPath=".status.namespace.name",name="Namespace Name",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.namespace.status",name="Namespace Status",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.admins",name="Namespace Admins",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.viewers",name="Namespace Viewers",type="string"
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=tenants,scope=Cluster

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}

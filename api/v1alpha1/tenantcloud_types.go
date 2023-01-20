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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TenantCloudSpec defines the desired state of TenantCloud
type TenantCloudSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Configure GCP credentials
	GCP GcpSpec `json:"gcp,omitempty"`
}

type GcpSpec struct {
	// Enable Google Cloud authentication resources
	Enabled bool `json:"enabled,omitempty"`
	// GCP service account email address
	ServiceAccount string `json:"serviceAccount"`
	// Generate service account access token via cronjob
	GenerateAccessToken bool `json:"generateAccessToken,omitempty"`
}

// TenantCloudStatus defines the observed state of TenantCloud
type TenantCloudStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Service account
	ServiceAccount ServiceAccountSpec `json:"serviceAccount"`
}

type ServiceAccountSpec struct {
	// Service account name
	Name string `json:"name"`
	// Service account namespace
	Namespace string `json:"namespace"`
	// GCP service account
	GcpServicAccount string `json:"gcpServiceAccount,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:JSONPath=".status.serviceAccount.name",name="ServiceAccount Name",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.serviceAccount.namespace",name="ServiceAccount Namespace",type="string"
//+kubebuilder:printcolumn:JSONPath=".status.serviceAccount.gcpServiceAccount",name="GCP ServiceAccount",type="string"
//+kubebuilder:subresource:status

// TenantCloud is the Schema for the tenantclouds API
type TenantCloud struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantCloudSpec   `json:"spec,omitempty"`
	Status TenantCloudStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantCloudList contains a list of TenantCloud
type TenantCloudList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantCloud `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantCloud{}, &TenantCloudList{})
}

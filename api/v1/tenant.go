// +groupName=projectx.mavenwave.dev
package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tenant
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=tenants,scope=Cluster
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

type TenantSpec struct {
	Admins  []Binding `json:"admins,omitempty"`
	Viewers []Binding `json:"viewers,omitempty"`
}

type Binding struct {
	ApiGroup  string `json:"apiGroup"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type TenantStatus struct {
	NamepaceConditions []v1.NamespaceCondition `json:"namespaceConditions,omitempty"`
	Admins             []Binding               `json:"admins,omitempty"`
	Viewers            []Binding               `json:"viewers,omitempty"`
}

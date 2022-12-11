// +groupName=projectx.mavenwave.dev
package v1

import (
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	// Namespace admin users
	Admins []rbacv1.Subject `json:"admins,omitempty"`
	// Namespace viewer users
	Viewers []rbacv1.Subject `json:"viewers,omitempty"`
}

type TenantStatus struct {
	NamepaceConditions []v1.NamespaceCondition `json:"namespaceConditions,omitempty"`
	// Namespace admin users
	Admins []rbacv1.Subject `json:"admins,omitempty"`
	// Namespace viewer users
	Viewers []rbacv1.Subject `json:"viewers,omitempty"`
}

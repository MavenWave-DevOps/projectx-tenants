// +groupName=projectx.mavenwave.dev
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Infrastructure
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=infrastructure,scope=Cluster
type Infrastructure struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   InfrastructureSpec   `json:"spec,omitempty"`
	Status InfrastructureStatus `json:"status,omitempty"`
}

type InfrastructureSpec struct {
	// Google Cloud provisioning configuration
	Google GoogleSpec `json:"google,omitempty"`
}

type GoogleSpec struct {
	// Create Google Cloud Provisioning resources
	Enabled bool `json:"enabled,omitempty"`
}

type InfrastructureStatus struct {
}

package controller

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Children struct {
	Namespace      []unstructured.Unstructured `json:"v1.Namespace"`
	Roles          []unstructured.Unstructured `json:"rbac.authorization.k8s.io/v1.Role"`
	RoleBindings   []unstructured.Unstructured `json:"rbac.authorization.k8s.io/v1.RoleBinding"`
	ServiceAccount []unstructured.Unstructured `json:"v1.ServiceAccount"`
	CronJob        []unstructured.Unstructured `json:"batch/v1.CronJob"`
	// GcpProviderConfig            []unstructured.Unstructured `json:"gcp.crossplane.io/v1beta1.ProviderConfig"`
	GcpUpboundProviderConfig     []unstructured.Unstructured `json:"gcp.upbound.io/v1beta1.ProviderConfig"`
	GcpServiceAccount            []unstructured.Unstructured `json:"cloudplatform.gcp.upbound.io/v1beta1.ServiceAccount"`
	GcpServiceAccountIAMPolicies []unstructured.Unstructured `json:"cloudplatform.gcp.upbound.io/v1beta1.ServiceAccountIAMMember"`
	GcpProjectIAMPolices         []unstructured.Unstructured `json:"cloudplatform.gcp.upbound.io/v1beta1.ProjectIAMMember"`
}

type SyncRequest struct {
	Parent   unstructured.Unstructured `json:"parent"`
	Children Children                  `json:"children,omitempty"`
}

type SyncResponse struct {
	Children []unstructured.Unstructured `json:"children"`
}

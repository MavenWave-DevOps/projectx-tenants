package infrastructure

import (
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Children struct {
	ServiceAccount            []unstructured.Unstructured `json:"v1.ServiceAccount"`
	Roles                     []unstructured.Unstructured `json:"rbac.authorization.k8s.io/v1.Role"`
	RoleBindings              []unstructured.Unstructured `json:"rbac.authorization.k8s.io/v1.RoleBinding"`
	CronJob                   []unstructured.Unstructured `json:"batch/v1.CronJob"`
	GcpProviderConfig         []unstructured.Unstructured `json:"gcp.crossplane.io/v1beta1.ProviderConfig"`
	GcpUpboundProviderConfig  []unstructured.Unstructured `json:"gcp.upbound.io/v1beta1.ProviderConfig"`
	GcpServiceAccount         []unstructured.Unstructured `json:"iam.gcp.crossplane.io/v1alpha1.ServiceAccount"`
	GcpServiceAccountPolicies []unstructured.Unstructured `json:"iam.gcp.crossplane.io/v1alpha1.ServiceAccountPolicy"`
}

type SyncRequest struct {
	Parent   unstructured.Unstructured `json:"parent"`
	Children Children                  `json:"children,omitempty"`
}

type SyncResponse struct {
	Children []unstructured.Unstructured `json:"children"`
}

func addAnnotations(req *SyncRequest) map[string]string {
	annotations := make(map[string]string)
	gcp, ok, err := unstructured.NestedBool(req.Parent.Object, "spec", "google", "enabled")
	if err != nil {
		log.Println(err)
	}
	if ok && gcp {
		proj, ok, err := unstructured.NestedString(req.Parent.Object, "spec", "google", "project")
		if err != nil {
			log.Println(err)
		}
		if ok {
			annotations["iam.gke.io/gcp-service-account"] = fmt.Sprintf("crossplane-%s@%s.iam.gserviceaccount.com", req.Parent.GetName(), proj)
		}
	}
	return annotations
}

func createServiceAccount(req *SyncRequest) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":        fmt.Sprintf("crossplane-%s", req.Parent.GetName()),
				"annotations": addAnnotations(req),
			},
		},
	}
}

func createRole(req *SyncRequest) unstructured.Unstructured {
	role := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "Role",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-sync", req.Parent.GetName()),
				"namespace": req.Parent.GetName(),
			},
			"rules": map[string]interface{}{
				"apiGroups": []interface{}{""},
				"resources": []interface{}{"secrets"},
				"verbs":     []interface{}{"get", "create", "patch"},
			},
		},
	}
	return role
}

func createRoleBinding(req *SyncRequest) unstructured.Unstructured {
	// subs := getSubjects(req, name)
	rbs := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "RoleBinding",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-sync-rb", req.Parent.GetName()),
				"namespace": req.Parent.GetName(),
			},
			"subjects": []map[string]interface{}{
				{
					"kind": "ServiceAccount",
					"name": fmt.Sprintf("crossplane-%s", req.Parent.GetName()),
				},
			},
			"roleRef": map[string]interface{}{
				"kind":     "Role",
				"name":     fmt.Sprintf("%s-sync", req.Parent.GetName()),
				"apiGroup": "",
			},
		},
	}
	return rbs
}

func createCronJob(req *SyncRequest) unstructured.Unstructured {
	cronjob := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "CronJob",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-credentials-sync", req.Parent.GetName()),
				"namespace": req.Parent.GetName(),
			},
			"spec": map[string]interface{}{
				"suspend":                    false,
				"schedule":                   "*/45 * * * *",
				"failedJobsHistoryLimit":     1,
				"successfulJobsHistoryLimit": 1,
				"concurrencyPolicy":          "Forbid",
				"startingDeadlineSeconds":    1800,
				"jobTemplate": map[string]interface{}{
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"serviceAccountName": fmt.Sprintf("crossplane-%s", req.Parent.GetName()),
								"restartPolicy":      "Never",
								"containers": []map[string]interface{}{
									{
										"image": "google/cloud-sdk:debian_component_based",
										"name":  "create-access-token",
									},
								},
							},
						},
					},
				},
				//             imagePullPolicy: IfNotPresent
				//             livenessProbe:
				//               exec:
				//                 command:
				//                 - gcloud
				//                 - version
				//             readinessProbe:
				//               exec:
				//                 command:
				//                 - gcloud
				//                 - version
				//             env:
				//               - name: SECRET_NAME
				//                 value: gcp-creds
				//               - name: SECRET_KEY
				//                 value: credentials
				//             command:
				//               - /bin/bash
				//               - -ce
				//               - |-
				//                 kubectl create secret generic $SECRET_NAME \
				//                   --dry-run=client \
				//                   --from-literal=$SECRET_KEY=\$(gcloud auth print-access-token) \
				//                   -o yaml | kubectl apply -f -
				//             resources:
				//               requests:
				//                 cpu: 250m
				//                 memory: 256Mi
				//               limits:
				//                 cpu: 500m
				//                 memory: 512Mi
			},
		},
	}
	return cronjob
}

func getProviderConfigRef(req *SyncRequest) string {
	confRef, ok, err := unstructured.NestedString(req.Parent.Object, "spec", "providerConfigRef", "name")
	if err != nil {
		log.Println(err)
		return "notfound"
	}
	if !ok {
		return "notfound"
	}
	return confRef
}

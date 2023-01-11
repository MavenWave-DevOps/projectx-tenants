package rbac

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	serviceAccount = "ServiceAccount"
)

type RoleBinding struct {
	Name      string
	Namespace string
	RoleRef   Role
	Subjects  []rbacv1.Subject
}

func CreateSubjects(rb *RoleBinding, l []interface{}) []map[string]interface{} {
	subjects := make([]map[string]interface{}, len(l))
	for i := range l {
		v := l[i].(map[string]interface{})
		if v["kind"] == serviceAccount {
			if _, ok := v["namespace"]; !ok {
				v["namespace"] = rb.Namespace
			}
		}
		subjects[i] = v
	}
	return subjects
}

func CreateRoleBinding(rb *RoleBinding, subjects []map[string]interface{}) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "RoleBinding",
			"metadata": map[string]interface{}{
				"name":      rb.Name,
				"namespace": rb.Namespace,
			},
			"roleRef": map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "Role",
				"name":     rb.RoleRef.Name,
			},
			"subjects": subjects,
		},
	}
}

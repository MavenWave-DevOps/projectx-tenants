package rbac

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Role struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	Rules       []Rule
}

type Rule struct {
	ApiGroups []string
	Resources []string
	Verbs     []string
}

func createRules(r []Rule) []map[string]interface{} {
	rules := make([]map[string]interface{}, len(r))
	for i, v := range r {
		m := map[string]interface{}{
			"apiGroups": v.ApiGroups,
			"resources": v.Resources,
			"verbs":     v.Verbs,
		}
		rules[i] = m
	}
	return rules
}

func CreateRole(r *Role) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "Role",
			"metadata": map[string]interface{}{
				"name":      r.Name,
				"namespace": r.Namespace,
			},
			"rules": createRules(r.Rules),
		},
	}
}

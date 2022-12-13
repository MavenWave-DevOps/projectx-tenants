package namespace

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Namespace struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
}

func CreateNs(n *Namespace) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name":        n.Name,
				"annotations": n.Annotations,
				"labels":      n.Labels,
			},
		},
	}
}

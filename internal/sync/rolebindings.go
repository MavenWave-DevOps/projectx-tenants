package sync

import (
	"fmt"
	"projectx-tenants/internal/controller"
	"projectx-tenants/internal/rbac"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func createRoleBindings(req *controller.SyncRequest, res *controller.SyncResponse, roles []rbac.Role) (*controller.SyncResponse, error) {
	for _, v := range roles {
		d := rbac.RoleBinding{
			Name:      fmt.Sprintf("%s-rb", v.Name),
			Namespace: req.Parent.GetName(),
			RoleRef:   v,
		}
		subs, err := addSubjects(&d, req, v.Name)
		if err != nil {
			return res, err
		}
		rb := rbac.CreateRoleBinding(&d, subs)
		res.Children = append(res.Children, rb)
	}
	return res, nil
}

func addSubjects(rb *rbac.RoleBinding, req *controller.SyncRequest, name string) ([]map[string]interface{}, error) {
	subs, ok, err := unstructured.NestedSlice(req.Parent.UnstructuredContent(), "spec", name)
	out := make([]map[string]interface{}, 0)
	if err != nil {
		return nil, err
	}

	if ok {
		out = rbac.CreateSubjects(rb, subs)
	}

	return out, nil
}

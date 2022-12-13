package sync

import (
	"projectx-tenants/internal/controller"
	"projectx-tenants/internal/rbac"
)

func createRoles(req *controller.SyncRequest, res *controller.SyncResponse) (*controller.SyncResponse, []rbac.Role) {
	roles := make([]rbac.Role, 0)
	adminRole := rbac.Role{
		Name:      "admins",
		Namespace: req.Parent.GetName(),
		Rules: []rbac.Rule{
			{
				ApiGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}
	roles = append(roles, adminRole)
	admins := rbac.CreateRole(&adminRole)
	res.Children = append(res.Children, admins)
	viewerRole := rbac.Role{
		Name:      "viewers",
		Namespace: req.Parent.GetName(),
		Rules: []rbac.Rule{
			{
				ApiGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	roles = append(roles, viewerRole)
	viewers := rbac.CreateRole(&viewerRole)
	res.Children = append(res.Children, viewers)
	return res, roles
}

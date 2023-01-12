package role

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"projectx.mavenwave.dev/internal/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Role struct {
	Name      string
	Namespace string
	Rules     []rbacv1.PolicyRule
	Owner     *projectxv1alpha1.Tenant
	Client    client.Client
	Scheme    *runtime.Scheme
}

func Create(ctx context.Context, r *Role) (*rbacv1.Role, error) {
	var err error
	role := &rbacv1.Role{}
	role.Name = r.Name
	role.Namespace = r.Namespace
	role.Rules = r.Rules
	foundRole := &rbacv1.Role{}
	if err := controllerutil.SetControllerReference(r.Owner, role, r.Scheme); err != nil {
		return role, err
	}
	err = r.Client.Get(ctx, types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating role", "role", role.GetName())
		err = r.Client.Create(ctx, role)
	} else if err == nil {
		if ok := comparePolicies(foundRole.Rules, r.Rules); !ok {
			role.Rules = r.Rules
			log.Log.Info("updating role", "role", role.GetName())
			err = r.Client.Update(ctx, role)
		}
	}
	return foundRole, err
}

func comparePolicies(a, b []rbacv1.PolicyRule) bool {
	for _, v := range a {
		for _, x := range b {
			if ok := common.CompareSliceString(v.Resources, x.Resources); !ok {
				return false
			}
			if ok := common.CompareSliceString(v.APIGroups, x.APIGroups); !ok {
				return false
			}
			if ok := common.CompareSliceString(v.Verbs, x.Verbs); !ok {
				return false
			}
		}
	}
	return true
}

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
}

func (r *Role) Create(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner *projectxv1alpha1.Tenant) (*rbacv1.Role, error) {
	var err error
	role := &rbacv1.Role{}
	role.Name = r.Name
	role.Namespace = r.Namespace
	role.Rules = r.Rules
	foundRole := &rbacv1.Role{}
	if err := controllerutil.SetControllerReference(owner, role, scheme); err != nil {
		return role, err
	}
	err = client.Get(ctx, types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating role", "tenant", owner.GetName(), "role", role.GetName())
		err = client.Create(ctx, role)
	} else if err == nil {
		if role.String() != foundRole.String() {
			log.Log.Info("updating role", "tenant", owner.GetName(), "role", role.GetName())
			err = client.Update(ctx, role)
		}
	}
	return foundRole, err
}

func (r *Role) Delete(ctx context.Context, client client.Client, owner *projectxv1alpha1.Tenant) error {
	role := &rbacv1.Role{}
	role.Name = r.Name
	role.Namespace = r.Namespace
	foundRole := &rbacv1.Role{}
	if err := client.Get(ctx, types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, foundRole); err == nil {
		log.Log.Info("deleting role", "tenant", owner.GetName(), "role", role.GetName())
		if err := client.Delete(ctx, foundRole); err != nil {
			return err
		}
	}
	return nil
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

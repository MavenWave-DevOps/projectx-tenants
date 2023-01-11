package role

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Role struct {
	Name      string
	Namespace string
	Rules     []rbacv1.PolicyRule
	Owner       *projectxv1alpha1.Tenant
	Client      client.Client
	Scheme      *runtime.Scheme
}

func Create(ctx context.Context, r *Role) (*rbacv1.Role, error) {
	if err := controllerutil.SetControllerReference(ns.Owner, namespace, ns.Scheme); err != nil {
		return namespace, err
	}
	// role := &rbacv1.Role{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      r.Name,
	// 		Namespace: r.Namespace,
	// 	},
	// 	Rules: r.Rules,
	// }
	// foundRole := &rbacv1.Role{}
	// err := client.Get(ctx, types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, foundRole)
	// if err != nil && errors.IsNotFound(err) {
	// 	log.Log.Info("Creating Role", "role", role.GetName())
	// 	if err := client.Create(ctx, role); err != nil {
	// 		log.Log.Info("failed to create role", "role", role.GetName())
	// 	}
		// } else if err == nil {
		// 	if err := client.Update(ctx, role); err != nil {
		// 		log.Log.Error(err, "role not updated", "role", role.GetName())
		// 	}
	}
	return foundRole, nil
}

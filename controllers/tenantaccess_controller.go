/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
)

// TenantAccessReconciler reconciles a TenantAccess object
type TenantAccessReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenantaccess,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenantaccess/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenantaccess/finalizers,verbs=update
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="*",resources="*",verbs="*"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TenantAccess object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *TenantAccessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var err error
	var tenantAccess projectxv1alpha1.TenantAccess
	err = r.Get(ctx, req.NamespacedName, &tenantAccess)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("unable to fetch TenantAccess", "error", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, v := range tenantAccess.Spec.Iam {
		role := &rbacv1.Role{}
		role.Name = v.Name
		role.Namespace = tenantAccess.GetNamespace()
		if role.GetAnnotations() == nil {
			role.SetAnnotations(make(map[string]string))
		}
		role.SetAnnotations(addToMap(role.GetAnnotations(), tenantAccess.GetAnnotations()))
		if role.GetLabels() == nil {
			role.SetLabels(make(map[string]string))
		}
		role.SetLabels(addToMap(role.GetLabels(), tenantAccess.GetLabels()))
		role.Rules = v.Rules
		foundRole := &rbacv1.Role{}
		if err := controllerutil.SetControllerReference(&tenantAccess, role, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = r.Get(ctx, types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, foundRole)
		if err != nil && errors.IsNotFound(err) {
			log.Log.Info("creating role", "tenant", tenantAccess.GetName(), "role", role.GetName())
			if err := r.Create(ctx, role); err != nil {
				return ctrl.Result{}, err
			}
		} else if err == nil {
			if role.String() != foundRole.String() {
				log.Log.Info("updating role", "tenant", tenantAccess.GetName(), "role", role.GetName())
				if err := r.Update(ctx, role); err != nil {
					return ctrl.Result{}, err
				}
			}
		}

		rb := &rbacv1.RoleBinding{}
		rb.Name = fmt.Sprintf("%s-rb", v.Name)
		rb.Namespace = tenantAccess.GetNamespace()
		if rb.GetAnnotations() == nil {
			rb.SetAnnotations(make(map[string]string))
		}
		rb.SetAnnotations(addToMap(rb.GetAnnotations(), tenantAccess.GetAnnotations()))
		if rb.GetLabels() == nil {
			rb.SetLabels(make(map[string]string))
		}
		rb.SetLabels(addToMap(rb.GetLabels(), tenantAccess.GetLabels()))
		rb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.GetName(),
		}
		for i := range v.Subjects {
			if v.Subjects[i].Kind == "ServiceAccount" {
				if v.Subjects[i].Namespace == "" {
					v.Subjects[i].Namespace = tenantAccess.GetNamespace()
				}
			}
		}
		rb.Subjects = v.Subjects
		foundRb := &rbacv1.RoleBinding{}
		if err := controllerutil.SetControllerReference(&tenantAccess, rb, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = r.Get(ctx, types.NamespacedName{Name: rb.GetName(), Namespace: rb.GetNamespace()}, foundRb)
		if err != nil && errors.IsNotFound(err) {
			log.Log.Info("creating rolebinding", "tenant", tenantAccess.GetName(), "rolebinding", rb.GetName())
			if err := r.Create(ctx, rb); err != nil {
				return ctrl.Result{}, err
			}
		} else if err == nil {
			if rb.String() != foundRb.String() {
				log.Log.Info("updating rolebinding", "tenant", tenantAccess.GetName(), "rolebinding", rb.GetName())
				if err := r.Update(ctx, rb); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantAccessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectxv1alpha1.TenantAccess{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}

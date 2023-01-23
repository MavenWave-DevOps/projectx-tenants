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

	v1 "k8s.io/api/core/v1"
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

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type role struct {
	name     string
	rules    []rbacv1.PolicyRule
	subjects []rbacv1.Subject
}

//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenants/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespaces/finalizers,verbs=update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles/finalizers,verbs=update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var err error
	var tenant projectxv1alpha1.Tenant
	err = r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("unable to fetch TenantCloud", "error", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ns := &v1.Namespace{}
	ns.Name = tenant.Name
	if ns.GetAnnotations() == nil {
		ns.SetAnnotations(make(map[string]string))
	}
	ns.SetAnnotations(addToMap(ns.GetAnnotations(), tenant.GetAnnotations()))
	ns.SetAnnotations(addToMap(ns.GetAnnotations(), tenant.Spec.Namespace.Annotations))
	if ns.GetLabels() == nil {
		ns.SetLabels(make(map[string]string))
	}
	ns.SetLabels(addToMap(ns.GetLabels(), tenant.GetLabels()))
	ns.SetLabels(addToMap(ns.GetLabels(), tenant.Spec.Namespace.Labels))
	foundNs := &v1.Namespace{}
	err = r.Get(ctx, types.NamespacedName{Name: ns.GetName()}, foundNs)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating namespace", "tenant", tenant.GetName(), "namespace", ns.GetName())
		if err := r.Create(ctx, ns); err != nil {
			return ctrl.Result{}, err
		}
	} else if err == nil {
		if ns.String() != foundNs.String() {
			log.Log.Info("updating namespace", "tenant", tenant.GetName(), "namespace", ns.GetName())
			if err := r.Update(ctx, ns); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	roles := []role{
		{
			name: "admins",
			rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"*"},
					Resources: []string{"*"},
					Verbs:     []string{"*"},
				},
			},
			subjects: tenant.Spec.Admins,
		},
		{
			name: "viewers",
			rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{"*"},
					Resources: []string{"*"},
					Verbs:     []string{"list", "get", "watch"},
				},
			},
			subjects: tenant.Spec.Viewers,
		},
	}

	for _, v := range roles {
		rl := &rbacv1.Role{}
		rl.Name = v.name
		rl.Namespace = tenant.Name
		if rl.GetAnnotations() == nil {
			rl.SetAnnotations(make(map[string]string))
		}
		rl.SetAnnotations(addToMap(rl.GetAnnotations(), tenant.GetAnnotations()))
		if rl.GetLabels() == nil {
			rl.SetLabels(make(map[string]string))
		}
		rl.SetLabels(addToMap(rl.GetLabels(), tenant.GetLabels()))
		rl.Rules = v.rules
		foundRl := &rbacv1.Role{}
		if err := controllerutil.SetControllerReference(&tenant, rl, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = r.Get(ctx, types.NamespacedName{Name: rl.GetName(), Namespace: rl.GetNamespace()}, foundRl)
		if err != nil && errors.IsNotFound(err) {
			log.Log.Info("creating role", "tenant", tenant.GetName(), "role", rl.GetName())
			if err := r.Create(ctx, rl); err != nil {
				return ctrl.Result{}, err
			}
		} else if err == nil {
			if rl.String() != foundRl.String() {
				log.Log.Info("updating role", "tenant", tenant.GetName(), "role", rl.GetName())
				if err := r.Update(ctx, rl); err != nil {
					return ctrl.Result{}, err
				}
			}
		}

		rb := &rbacv1.RoleBinding{}
		rb.Name = fmt.Sprintf("%s-rb", v.name)
		rb.Namespace = tenant.Name
		if rb.GetAnnotations() == nil {
			rb.SetAnnotations(make(map[string]string))
		}
		rb.SetAnnotations(addToMap(rb.GetAnnotations(), tenant.GetAnnotations()))
		if rb.GetLabels() == nil {
			rb.SetLabels(make(map[string]string))
		}
		rb.SetLabels(addToMap(rb.GetLabels(), tenant.GetLabels()))
		rb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     v.name,
		}
		for i := range v.subjects {
			if v.subjects[i].Kind == "ServiceAccount" {
				if v.subjects[i].Namespace == "" {
					v.subjects[i].Namespace = tenant.Name
				}
			}
		}
		rb.Subjects = v.subjects
		foundRb := &rbacv1.RoleBinding{}
		if err := controllerutil.SetControllerReference(&tenant, rb, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = r.Get(ctx, types.NamespacedName{Name: rb.GetName(), Namespace: rb.GetNamespace()}, foundRb)
		if err != nil && errors.IsNotFound(err) {
			log.Log.Info("creating rolebinding", "tenant", tenant.GetName(), "rolebinding", rb.GetName())
			if err := r.Create(ctx, rb); err != nil {
				return ctrl.Result{}, err
			}
		} else if err == nil {
			if rb.String() != foundRb.String() {
				log.Log.Info("updating rolebinding", "tenant", tenant.GetName(), "rolebinding", rb.GetName())
				if err := r.Update(ctx, rb); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectxv1alpha1.Tenant{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}

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

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"projectx.mavenwave.dev/internal/namespace"
	role "projectx.mavenwave.dev/internal/roles"
)

var (
	apiGVStr    = projectxv1alpha1.GroupVersion.String()
	jobOwnerKey = ".tenant.controller"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenants/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespace,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespace/status,verbs=get

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
	var tenant projectxv1alpha1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Error(err, "unable to fetch Tenant")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Create namespace
	ns := &namespace.Namespace{
		Name:        tenant.Name,
		Labels:      tenant.Labels,
		Annotations: tenant.Annotations,
		Owner:       &tenant,
		Client:      r.Client,
		Scheme:      r.Scheme,
	}
	if _, err := namespace.Create(ctx, ns); err != nil {
		return ctrl.Result{}, err
	}
	// Create admin role
	adminRoleReq := &role.Role{
		Name:      "admin",
		Namespace: tenant.GetName(),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}

	adminRole, err := role.Create(ctx, r.Client, adminRoleReq)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(&tenant, adminRole, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	// Create viewer role
	viewerRoleReq := &role.Role{
		Name:      "viewer",
		Namespace: tenant.GetName(),
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	viewerRole, err := role.Create(ctx, r.Client, viewerRoleReq)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := controllerutil.SetControllerReference(&tenant, viewerRole, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1.Namespace{}, jobOwnerKey, func(rawObj client.Object) []string {
	// 	// grab the job object, extract the owner...
	// 	ns := rawObj.(*v1.Namespace)
	// 	owner := metav1.GetControllerOf(ns)
	// 	if owner == nil {
	// 		return nil
	// 	}
	// 	// ...make sure it's a Tenant...
	// 	if owner.APIVersion != apiGVStr || owner.Kind != "Tenant" {
	// 		return nil
	// 	}

	// 	// ...and if so, return it
	// 	return []string{owner.Name}
	// }); err != nil {
	// 	return err
	// }
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectxv1alpha1.Tenant{}).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.Role{}).
		Complete(r)
}

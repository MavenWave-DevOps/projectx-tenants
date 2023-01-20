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

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
)

// TenantCloudReconciler reconciles a TenantCloud object
type TenantCloudReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenantclouds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenantclouds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=projectx.mavenwave.dev,resources=tenantclouds/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts/finalizers,verbs=update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles/finalizers,verbs=update
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="batch",resources=cronjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TenantCloud object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *TenantCloudReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	var tenant projectxv1alpha1.TenantCloud
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("unable to fetch TenantCloud", "error", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if tenant.Spec.GCP.Enabled {
		if err := r.SetupGcp(ctx, req, &tenant); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.Client.Status().Update(ctx, &tenant); err != nil {
		log.Log.Info("failed to update tenant status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantCloudReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&projectxv1alpha1.TenantCloud{}).
		Owns(&v1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}

func (r *TenantCloudReconciler) SetupGcp(ctx context.Context, req ctrl.Request, tenant *projectxv1alpha1.TenantCloud) error {
	var err error
	// Create service account
	sa := &v1.ServiceAccount{}
	sa.Name = tenant.Name
	sa.Namespace = tenant.Namespace
	sa.SetLabels(addToMap(tenant.GetLabels()))
	sa.SetAnnotations(addToMap(tenant.GetAnnotations()))
	sa.Annotations["iam.gke.io/gcp-service-account"] = tenant.Spec.GCP.ServiceAccount
	foundSa := &v1.ServiceAccount{}
	if err := controllerutil.SetControllerReference(tenant, sa, r.Scheme); err != nil {
		return err
	}
	err = r.Get(ctx, types.NamespacedName{Name: sa.GetName(), Namespace: sa.GetNamespace()}, foundSa)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating serviceaccount", "tenant", tenant.GetName(), "serviceaccount", sa.GetName())
		if err := r.Create(ctx, sa); err != nil {
			return err
		}
	} else if err == nil {
		if sa.String() != foundSa.String() {
			log.Log.Info("updating serviceaccount", "tenant", tenant.GetName(), "serviceaccount", sa.GetName())
			if err := r.Update(ctx, sa); err != nil {
				return err
			}
		}
	}
	// tenant.Status.ServiceAccount.Name = foundSa.Name
	// tenant.Status.ServiceAccount.Namespace = foundSa.Namespace
	// tenant.Status.ServiceAccount.GcpServicAccount = foundSa.Annotations["iam.gke.io/gcp-service-account"]
	// Create role
	role := &rbacv1.Role{}
	role.Name = tenant.Name
	role.Namespace = tenant.Namespace
	role.SetLabels(addToMap(tenant.GetLabels()))
	role.SetAnnotations(addToMap(tenant.GetAnnotations()))
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get", "create", "patch"},
		},
	}
	foundRole := &rbacv1.Role{}
	if err := controllerutil.SetControllerReference(tenant, role, r.Scheme); err != nil {
		return err
	}
	err = r.Get(ctx, types.NamespacedName{Name: role.GetName(), Namespace: role.GetNamespace()}, foundRole)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating role", "tenant", tenant.GetName(), "role", role.GetName())
		if err := r.Create(ctx, role); err != nil {
			return err
		}
	} else if err == nil {
		if role.String() != foundRole.String() {
			log.Log.Info("updating role", "tenant", tenant.GetName(), "role", role.GetName())
			if err := r.Update(ctx, role); err != nil {
				return err
			}
		}
	}
	// Create rolebinding
	rb := &rbacv1.RoleBinding{}
	rb.Name = fmt.Sprintf("%s-rb", tenant.Name)
	rb.Namespace = tenant.Namespace
	rb.SetLabels(addToMap(tenant.GetLabels()))
	rb.SetAnnotations(addToMap(tenant.GetAnnotations()))
	rb.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     role.Name,
	}
	rb.Subjects = []rbacv1.Subject{
		{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      sa.Name,
			Namespace: sa.Namespace,
		},
	}
	foundRb := &rbacv1.RoleBinding{}
	if err := controllerutil.SetControllerReference(tenant, rb, r.Scheme); err != nil {
		return err
	}
	err = r.Get(ctx, types.NamespacedName{Name: rb.GetName(), Namespace: rb.GetNamespace()}, foundRb)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating rolebinding", "tenant", tenant.GetName(), "rolebinding", rb.GetName())
		if err := r.Create(ctx, rb); err != nil {
			return err
		}
	} else if err == nil {
		if rb.String() != foundRb.String() {
			log.Log.Info("updating rolebinding", "tenant", tenant.GetName(), "rolebinding", rb.GetName())
			if err := r.Update(ctx, rb); err != nil {
				return err
			}
		}
	}
	// Create cronjob
	cronjob := &batchv1.CronJob{}
	cronjob.Name = tenant.Name
	cronjob.Namespace = tenant.Namespace
	cronjob.SetLabels(addToMap(tenant.GetLabels()))
	cronjob.SetAnnotations(addToMap(tenant.GetAnnotations()))
	cronjob.Spec.Schedule = "*/45 * * * *"
	cronjob.Spec.FailedJobsHistoryLimit = int32Ptr(1)
	cronjob.Spec.SuccessfulJobsHistoryLimit = int32Ptr(1)
	cronjob.Spec.ConcurrencyPolicy = batchv1.ConcurrencyPolicy("Forbid")
	cronjob.Spec.StartingDeadlineSeconds = int64Ptr(1800)
	cronjob.Spec.Suspend = booPtr(!tenant.Spec.GCP.GenerateAccessToken)
	cronjob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = sa.Name
	cronjob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = int64Ptr(600)
	cronjob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = v1.RestartPolicy("Never")
	cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers = []v1.Container{
		{
			Name:            "create-access-token",
			Image:           "google/cloud-sdk:debian_component_based",
			ImagePullPolicy: "IfNotPresent",
			LivenessProbe: &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					Exec: &v1.ExecAction{Command: []string{"gcloud", "version"}},
				},
			},
			ReadinessProbe: &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					Exec: &v1.ExecAction{Command: []string{"gcloud", "version"}},
				},
			},
			Env: []v1.EnvVar{
				{
					Name:  "SECRET_NAME",
					Value: fmt.Sprintf("%s-gcp-credentials", tenant.Name),
				},
				{
					Name:  "SECRET_KEY",
					Value: "credentials",
				},
			},
			Command: []string{"/bin/bash", "-ce", "kubectl create secret generic $SECRET_NAME --dry-run=client --from-literal=$SECRET_KEY=$(gcloud auth print-access-token) -o yaml | kubectl apply -f -"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(250, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(256*1024*1024, resource.BinarySI),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(512*1024*1024, resource.BinarySI),
				},
			},
		},
	}
	foundCronjob := &batchv1.CronJob{}
	if err := controllerutil.SetControllerReference(tenant, cronjob, r.Scheme); err != nil {
		return err
	}
	err = r.Get(ctx, types.NamespacedName{Name: cronjob.GetName(), Namespace: cronjob.GetNamespace()}, foundCronjob)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating cronjob", "tenant", tenant.GetName(), "cronjob", cronjob.GetName())
		if err := r.Create(ctx, cronjob); err != nil {
			return err
		}
	} else if err == nil {
		if cronjob.String() != foundCronjob.String() {
			log.Log.Info("updating cronjob", "tenant", tenant.GetName(), "cronjob", cronjob.GetName())
			if err := r.Update(ctx, cronjob); err != nil {
				return err
			}
		}
	}
	return nil
}

func addToMap(tenant map[string]string) map[string]string {
	obj := make(map[string]string)
	for k, v := range tenant {
		obj[k] = v
	}
	return obj
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func booPtr(b bool) *bool {
	return &b
}

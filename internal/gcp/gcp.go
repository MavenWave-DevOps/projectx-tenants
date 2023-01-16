package gcp

import (
	"context"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"projectx.mavenwave.dev/internal/cronjob"
	"projectx.mavenwave.dev/internal/role"
	"projectx.mavenwave.dev/internal/rolebinding"
	"projectx.mavenwave.dev/internal/serviceaccount"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	saName      = "gcp-auth"
	rbName      = "gcp-auth-rb"
	cronjobName = "gcp-credential-sync"
)

func Create(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner *projectxv1alpha1.Tenant) error {
	sa := serviceaccount.ServiceAccount{
		Name:      saName,
		Namespace: owner.Name,
		Annotations: map[string]string{
			"iam.gke.io/gcp-service-account": owner.Spec.Infrastructure.GCP.ServiceAccount,
		},
	}
	if _, err := sa.Create(ctx, client, scheme, owner); err != nil {
		return err
	}
	role := role.Role{
		Name:      saName,
		Namespace: owner.Name,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "create", "patch"},
			},
		},
	}
	if _, err := role.Create(ctx, client, scheme, owner); err != nil {
		return err
	}
	rb := rolebinding.Rolebinding{
		Name:      rbName,
		Namespace: owner.Name,
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: "",
				Kind:     "ServiceAccount",
				Name:     sa.Name,
			},
		},
	}
	if _, err := rb.Create(ctx, client, scheme, owner); err != nil {
		return err
	}
	cronjob := cronjob.CronJob{
		Name:                       cronjobName,
		Namespace:                  owner.Name,
		Schedule:                   "*/45 * * * *",
		FailedJobsHistoryLimit:     1,
		SuccessfulJobsHistoryLimit: 1,
		ConcurrencyPolicy:          "Forbid",
		StartingDeadlineSeconds:    1800,
		ActiveDeadlineSeconds:      600,
		Suspend:                    !owner.Spec.Infrastructure.GCP.GenerateAccessToken,
		ServiceAccountName:         sa.Name,
		RestartPolicy:              "Never",
		Containers: []v1.Container{
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
						Value: "gcp-credentials",
					},
					{
						Name:  "SECRET_KEY",
						Value: "credentials",
					},
				},
				Command: []string{"/bin/bash", "-ce", "TOKEN=$(gcloud auth print-access-token); kubectl create secret generic $SECRET_NAME --dry-run=client --from-literal=$SECRET_KEY=$TOKEN -o yaml | kubectl apply -f -"},
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
		},
	}
	if _, err := cronjob.Create(ctx, client, scheme, owner); err != nil {
		return err
	}
	return nil
}

func Delete(ctx context.Context, client client.Client, owner *projectxv1alpha1.Tenant) error {
	sa := serviceaccount.ServiceAccount{
		Name:      saName,
		Namespace: owner.Name,
	}
	if err := sa.Delete(ctx, client, owner); err != nil {
		return err
	}
	role := role.Role{
		Name:      saName,
		Namespace: owner.Name,
	}
	if err := role.Delete(ctx, client, owner); err != nil {
		return err
	}
	rb := rolebinding.Rolebinding{
		Name:      saName,
		Namespace: owner.Name,
	}
	if err := rb.Delete(ctx, client, owner); err != nil {
		return err
	}
	cronjob := cronjob.CronJob{
		Name:      cronjobName,
		Namespace: owner.Name,
	}
	if err := cronjob.Delete(ctx, client, owner); err != nil {
		return err
	}
	return nil
}

package cronjob

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CronJob struct {
	Name                       string
	Namespace                  string
	Schedule                   string
	FailedJobsHistoryLimit     int32
	SuccessfulJobsHistoryLimit int32
	ConcurrencyPolicy          string
	StartingDeadlineSeconds    int64
	Suspend                    bool
	ServiceAccountName         string
	ActiveDeadlineSeconds      int64
	RestartPolicy              string
	Containers                 []v1.Container
}

func (c *CronJob) Create(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner *projectxv1alpha1.Tenant) (*batchv1.CronJob, error) {
	var err error
	cronjob := &batchv1.CronJob{}
	cronjob.Name = c.Name
	cronjob.Namespace = c.Namespace
	cronjob.Spec.Schedule = c.Schedule
	cronjob.Spec.FailedJobsHistoryLimit = int32Ptr(c.FailedJobsHistoryLimit)
	cronjob.Spec.SuccessfulJobsHistoryLimit = int32Ptr(c.SuccessfulJobsHistoryLimit)
	cronjob.Spec.ConcurrencyPolicy = batchv1.ConcurrencyPolicy(c.ConcurrencyPolicy)
	cronjob.Spec.StartingDeadlineSeconds = int64Ptr(c.StartingDeadlineSeconds)
	cronjob.Spec.Suspend = booPtr(c.Suspend)
	cronjob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = c.ServiceAccountName
	cronjob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = int64Ptr(c.ActiveDeadlineSeconds)
	cronjob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = v1.RestartPolicy(c.RestartPolicy)
	cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers = c.Containers
	foundCronjob := &batchv1.CronJob{}
	if err := controllerutil.SetControllerReference(owner, cronjob, scheme); err != nil {
		return cronjob, err
	}
	err = client.Get(ctx, types.NamespacedName{Name: cronjob.GetName(), Namespace: cronjob.GetNamespace()}, foundCronjob)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating cronjob", "tenant", owner.GetName(), "cronjob", cronjob.GetName())
		err = client.Create(ctx, cronjob)
	} else if err == nil {
		if cronjob.String() != foundCronjob.String() {
			log.Log.Info("updating cronjob", "tenant", owner.GetName(), "cronjob", cronjob.GetName())
			err = client.Update(ctx, cronjob)
		}
	}
	return foundCronjob, err
}

func (c *CronJob) Delete(ctx context.Context, client client.Client, owner *projectxv1alpha1.Tenant) error {
	cronjob := &batchv1.CronJob{}
	cronjob.Name = c.Name
	cronjob.Namespace = c.Namespace
	foundCronjob := &batchv1.CronJob{}
	if err := client.Get(ctx, types.NamespacedName{Name: cronjob.GetName(), Namespace: cronjob.GetNamespace()}, foundCronjob); err == nil {
		log.Log.Info("deleting cronjob", "tenant", owner.GetName(), "cronjob", cronjob.GetName())
		if err := client.Delete(ctx, foundCronjob); err != nil {
			return err
		}
	}
	return nil
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

package serviceaccount

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ServiceAccount struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

func (s *ServiceAccount) Create(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner *projectxv1alpha1.Tenant) (*v1.ServiceAccount, error) {
	var err error
	sa := &v1.ServiceAccount{}
	sa.Name = s.Name
	sa.Namespace = s.Namespace
	sa.Labels = s.Labels
	sa.Annotations = s.Annotations
	foundSa := &v1.ServiceAccount{}
	if err := controllerutil.SetControllerReference(owner, sa, scheme); err != nil {
		return sa, err
	}
	err = client.Get(ctx, types.NamespacedName{Name: sa.GetName(), Namespace: sa.GetNamespace()}, foundSa)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating serviceaccount", "tenant", owner.GetName(), "serviceaccount", sa.GetName())
		err = client.Create(ctx, sa)
	} else if err == nil {
		if sa.String() != foundSa.String() {
			log.Log.Info("updating serviceaccount", "tenant", owner.GetName(), "serviceaccount", sa.GetName())
			err = client.Update(ctx, sa)
		}
	}
	return foundSa, err
}

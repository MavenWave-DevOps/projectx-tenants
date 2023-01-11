package namespace

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

type Namespace struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Owner       *projectxv1alpha1.Tenant
	Client      client.Client
	Scheme      *runtime.Scheme
}

func Create(ctx context.Context, ns *Namespace) (*v1.Namespace, error) {
	namespace := &v1.Namespace{}
	namespace.Name = ns.Name
	if err := controllerutil.SetControllerReference(ns.Owner, namespace, ns.Scheme); err != nil {
		return namespace, err
	}
	foundNamespace := &v1.Namespace{}
	err := ns.Client.Get(ctx, types.NamespacedName{Name: namespace.GetName()}, foundNamespace)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating namespace", "namespace", namespace.GetName())
		err = ns.Client.Create(ctx, namespace)
	} else if err == nil {
		namespace.Labels = ns.Labels
		namespace.Annotations = ns.Annotations
		log.Log.Info("updating namespace", "namespace", namespace.GetName())
		err = ns.Client.Update(ctx, namespace)
	}
	return namespace, err
}

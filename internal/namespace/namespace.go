package namespace

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Namespace struct {
	Name        string
	Annotations map[string]string
	Labels      map[string]string
}

func (ns *Namespace) Create(ctx context.Context, client client.Client, owner *projectxv1alpha1.Tenant) (*v1.Namespace, error) {
	namespace := &v1.Namespace{}
	namespace.Name = ns.Name
	for k, v := range owner.Annotations {
		ns.Annotations[k] = v
	}
	namespace.SetAnnotations(ns.Annotations)
	for k, v := range owner.Labels {
		ns.Labels[k] = v
	}
	namespace.SetLabels(ns.Labels)
	foundNamespace := &v1.Namespace{}
	err := client.Get(ctx, types.NamespacedName{Name: namespace.GetName()}, foundNamespace)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating namespace", "tenant", owner.GetName(), "namespace", namespace.GetName())
		err = client.Create(ctx, namespace)
	} else if err == nil {
		if namespace.String() != foundNamespace.String() {
			log.Log.Info("updating namespace", "tenant", owner.GetName(), "namespace", namespace.GetName())
			err = client.Update(ctx, namespace)
		}
	}
	return foundNamespace, err
}

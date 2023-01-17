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

func (ns *Namespace) SetAnnotations(annotations map[string]string) {
	if len(ns.Annotations) == 0 {
		ns.Annotations = make(map[string]string)
	}
	for k, v := range annotations {
		ns.Annotations[k] = v
	}
}

func (ns *Namespace) SetLabels(labels map[string]string) {
	if len(ns.Labels) == 0 {
		ns.Labels = make(map[string]string)
	}
	for k, v := range labels {
		ns.Labels[k] = v
	}
}

func (ns *Namespace) Create(ctx context.Context, client client.Client, owner *projectxv1alpha1.Tenant) (*v1.Namespace, error) {
	namespace := &v1.Namespace{}
	namespace.Name = ns.Name
	ns.SetAnnotations(owner.Annotations)
	namespace.SetAnnotations(ns.Annotations)
	ns.SetLabels(owner.Labels)
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

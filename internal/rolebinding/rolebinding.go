package rolebinding

import (
	"context"
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	projectxv1alpha1 "projectx.mavenwave.dev/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	serviceAccount = "ServiceAccount"
)

type Rolebinding struct {
	Name      string
	Namespace string
	RoleRef   rbacv1.RoleRef
	Subjects  []rbacv1.Subject
}

func (r *Rolebinding) Create(ctx context.Context, client client.Client, scheme *runtime.Scheme, owner *projectxv1alpha1.Tenant) (*rbacv1.RoleBinding, error) {
	var err error
	rb := &rbacv1.RoleBinding{}
	rb.Name = r.Name
	rb.Namespace = r.Namespace
	rb.SetAnnotations(owner.Annotations)
	rb.SetLabels(owner.Labels)
	rb.RoleRef = r.RoleRef
	rb.Subjects = r.Subjects
	for i, v := range rb.Subjects {
		if v.Kind == serviceAccount {
			if v.Namespace == "" {
				rb.Subjects[i].Namespace = r.Namespace
			}
		}
	}
	foundRb := &rbacv1.RoleBinding{}
	if err := controllerutil.SetControllerReference(owner, rb, scheme); err != nil {
		return rb, err
	}
	err = client.Get(ctx, types.NamespacedName{Name: rb.GetName(), Namespace: rb.GetNamespace()}, foundRb)
	if err != nil && errors.IsNotFound(err) {
		log.Log.Info("creating rolebinding", "tenant", owner.GetName(), "rolebinding", rb.GetName())
		err = client.Create(ctx, rb)
	} else if err == nil {
		if rb.String() != foundRb.String() {
			log.Log.Info("updating rolebinding", "tenant", owner.GetName(), "rolebinding", rb.GetName())
			err = client.Update(ctx, rb)
		}
	}
	return foundRb, err
}

func (r *Rolebinding) Delete(ctx context.Context, client client.Client, owner *projectxv1alpha1.Tenant) error {
	rb := &rbacv1.RoleBinding{}
	rb.Name = r.Name
	rb.Namespace = r.Namespace
	foundRb := &rbacv1.RoleBinding{}
	if err := client.Get(ctx, types.NamespacedName{Name: rb.GetName(), Namespace: rb.GetNamespace()}, foundRb); err == nil {
		log.Log.Info("deleting rolebinding", "tenant", owner.GetName(), "rolebinding", rb.GetName())
		if err := client.Delete(ctx, foundRb); err != nil {
			return err
		}
	}
	return nil
}

func ListSubjectsStr(s []rbacv1.Subject) string {
	out := make([]string, len(s))
	for i, v := range s {
		if v.Kind == serviceAccount {
			out[i] = fmt.Sprintf("%s/%s/%s", v.Kind, v.Namespace, v.Name)
		} else {
			out[i] = fmt.Sprintf("%s/%s", v.Kind, v.Name)
		}
	}
	return strings.Join(out, ",")
}

func compareRoleRef(a, b rbacv1.RoleRef) bool {
	if a.APIGroup != b.APIGroup {
		return false
	}
	if a.Kind != b.Kind {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	return true
}

func compareSubjects(a, b []rbacv1.Subject) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v.APIGroup != b[i].APIGroup {
			return false
		}
		if v.Kind != b[i].Kind {
			return false
		}
		if v.Name != b[i].Name {
			return false
		}
		if v.Namespace != b[i].Namespace {
			return false
		}
	}
	return true
}

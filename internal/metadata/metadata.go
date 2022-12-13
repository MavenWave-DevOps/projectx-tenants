package metadata

import "projectx-tenants/internal/controller"

func GetLabels(req *controller.SyncRequest) map[string]string {
	l := req.Parent.GetLabels()
	if l != nil {
		return l
	}
	return map[string]string{}
}

func GetAnnotations(req *controller.SyncRequest) map[string]string {
	a := req.Parent.GetAnnotations()
	if a != nil {
		return a
	}
	return map[string]string{}
}

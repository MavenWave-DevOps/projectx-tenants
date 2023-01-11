package sync

import (
	"projectx-tenants/internal/controller"
	"projectx-tenants/internal/metadata"
	"projectx-tenants/internal/namespace"
)

func createNs(req *controller.SyncRequest, res *controller.SyncResponse) *controller.SyncResponse {
	reqNs := namespace.Namespace{
		Name:        req.Parent.GetName(),
		Annotations: metadata.GetAnnotations(req),
		Labels:      metadata.GetLabels(req),
	}
	ns := namespace.CreateNs(&reqNs)
	res.Children = append(res.Children, ns)
	return res
}

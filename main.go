package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Children struct {
	Namespace []unstructured.Unstructured `json:"v1.Namespace"`
	Roles     []unstructured.Unstructured `json:"rbac.authorization.k8s.io/v1.Role"`
}

type SyncRequest struct {
	Parent   unstructured.Unstructured `json:"parent"`
	Children Children                  `json:"children,omitempty"`
}

type SyncResponse struct {
	Children []unstructured.Unstructured `json:"children"`
}

func getLabels(req *SyncRequest) map[string]string {
	l := req.Parent.GetLabels()
	if l != nil {
		return l
	}
	return map[string]string{}
}

func getAnnotations(req *SyncRequest) map[string]string {
	a := req.Parent.GetAnnotations()
	if a != nil {
		return a
	}
	return map[string]string{}
}

func createNs(req *SyncRequest) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name":        req.Parent.GetName(),
				"annotations": getAnnotations(req),
				"labels":      getLabels(req),
			},
		},
	}
}

func createRole(req *SyncRequest, name string, rules []map[string]interface{}) unstructured.Unstructured {
	role := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "Role",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": req.Parent.GetName(),
			},
			"rules": rules,
		},
	}
	return role
}

func getSubjects(req *SyncRequest, name string) []map[string]interface{} {
	subs, _, _ := unstructured.NestedSlice(req.Parent.UnstructuredContent(), "spec", name)
	esubs := make([]map[string]interface{}, len(subs))
	for i := range subs {
		v := subs[i].(map[string]interface{})
		esubs[i] = v
		if v["kind"] == "ServiceAccount" {
			if _, ok := v["namespace"]; !ok {
				esubs[i]["namespace"] = req.Parent.GetName()
			}
		}
	}
	return esubs
}

func createRoleBinding(req *SyncRequest, name string) unstructured.Unstructured {
	subs := getSubjects(req, name)
	rbs := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "RoleBinding",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-rb", name),
				"namespace": req.Parent.GetName(),
			},
			"roleRef": map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "Role",
				"name":     name,
			},
			"subjects": subs,
		},
	}
	return rbs
}

func sync(req *SyncRequest) *SyncResponse {
	var res SyncResponse
	ns := createNs(req)
	res.Children = append(res.Children, ns)
	var rules []map[string]interface{}
	rules = []map[string]interface{}{
		{
			"apiGroups": []interface{}{
				"*",
			},
			"resources": []interface{}{
				"*",
			},
			"verbs": []interface{}{
				"*",
			},
		},
	}
	admins := createRole(req, "admins", rules)
	res.Children = append(res.Children, admins)
	rules = []map[string]interface{}{
		{
			"apiGroups": []interface{}{
				"*",
			},
			"resources": []interface{}{
				"*",
			},
			"verbs": []interface{}{
				"get",
				"watch",
				"list",
			},
		},
	}
	viewers := createRole(req, "viewers", rules)
	res.Children = append(res.Children, viewers)
	adminrb := createRoleBinding(req, "admins")
	res.Children = append(res.Children, adminrb)
	viewerrb := createRoleBinding(req, "viewers")
	res.Children = append(res.Children, viewerrb)
	return &res
}

func main() {
	http.HandleFunc("/sync", syncHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	var req SyncRequest
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(b, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := sync(&req)
	body, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

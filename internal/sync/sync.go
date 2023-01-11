package sync

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"projectx-tenants/internal/controller"
)

func sync(req *controller.SyncRequest) (*controller.SyncResponse, error) {
	res := new(controller.SyncResponse)
	var err error
	createNs(req, res)
	_, roles := createRoles(req, res)
	if _, err = createRoleBindings(req, res, roles); err != nil {
		return nil, err
	}
	return res, nil
}

func SyncHandler(w http.ResponseWriter, r *http.Request) {
	var req controller.SyncRequest
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(b, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := sync(&req)
	if err != nil {
		log.Printf("ERROR Could not sync: %s", err)
	}
	body, err := json.Marshal(&res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

package main

import (
	"log"
	"net/http"
	"projectx-tenants/internal/sync"
)

func main() {
	http.HandleFunc("/sync", sync.SyncHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"log"
	"net/http"
	"projectx-tenants/internal/namespace"
)

func main() {
	http.HandleFunc("/tenant", namespace.SyncHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

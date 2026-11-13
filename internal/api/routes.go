package api

import (
	"net/http"

	"hardware-monitoring-control-platform/internal/core"
)

// HTTP yönlendiricisini (ServeMux) kurma ve handler'ları core state ile bağlama.
func NewRouter(state *core.ServerState) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/sensors", EnableCORS(HandleGetSensors(state)))
	mux.HandleFunc("/api/relay", EnableCORS(HandlePostRelay(state)))
	mux.HandleFunc("/api/relays", EnableCORS(HandleGetRelays(state)))

	return mux
}

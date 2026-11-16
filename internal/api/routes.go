package api

import (
	"net/http"

	"hardware-monitoring-control-platform/internal/core"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(state *core.ServerState) http.Handler {
	mux := http.NewServeMux()

	// JWT gerektirmeyen
	mux.HandleFunc("/api/login", EnableCORS(HandleLoginProxy))
	mux.Handle("/metrics", promhttp.Handler())

	// JWT gerektiren
	mux.HandleFunc("/api/devices", EnableCORS(AuthMiddleware(HandleGetDevices(state))))
	mux.HandleFunc("/api/sensors", EnableCORS(AuthMiddleware(HandleGetSensors(state))))
	mux.HandleFunc("/api/relays", EnableCORS(AuthMiddleware(HandleGetRelays(state))))
	mux.HandleFunc("/api/relay", EnableCORS(AuthMiddleware(HandlePostRelay(state))))

	return mux
}

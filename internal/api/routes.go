package api

import (
	"net/http"

	"hardware-monitoring-control-platform/internal/core"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(state *core.ServerState) http.Handler {
	mux := http.NewServeMux()

	// JWT gerektirmeyen
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/api/login", EnableCORS(PrometheusMiddleware("/api/login", HandleLoginProxy)))

	// 3. JWT gerektiren endpoint'ler
	mux.HandleFunc("/api/devices", EnableCORS(AuthMiddleware(PrometheusMiddleware("/api/devices", HandleGetDevices(state)))))
	mux.HandleFunc("/api/sensors", EnableCORS(AuthMiddleware(PrometheusMiddleware("/api/sensors", HandleGetSensors(state)))))
	mux.HandleFunc("/api/relays", EnableCORS(AuthMiddleware(PrometheusMiddleware("/api/relays", HandleGetRelays(state)))))
	mux.HandleFunc("/api/relay", EnableCORS(AuthMiddleware(PrometheusMiddleware("/api/relay", HandlePostRelay(state)))))
	return mux
}

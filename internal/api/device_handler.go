package api

import (
	"encoding/json"
	"hardware-monitoring-control-platform/internal/core"
	"net/http"
)

func HandleGetDevices(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state.AppConfig.Devices)
	}
}

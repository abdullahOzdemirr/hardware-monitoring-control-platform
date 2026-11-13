package api

import (
	"encoding/json"
	"net/http"
	"time"

	"hardware-monitoring-control-platform/internal/core"
)

// anlık sıcaklık ve nem verilerini HTTP üzerinden servis etme.
func HandleGetSensors(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state.DataMu.RLock()
		temp := state.LatestTemp
		hum := state.LatestHum
		state.DataMu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"temperature": temp,
			"humidity":    hum,
			"timestamp":   time.Now().Format("15:04:05"),
		})
	}
}

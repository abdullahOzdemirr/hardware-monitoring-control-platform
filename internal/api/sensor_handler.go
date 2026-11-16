package api

import (
	"encoding/json"
	"hardware-monitoring-control-platform/internal/core"
	"net/http"
	"time"
)

func HandleGetSensors(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state.DataMu.RLock()
		sensorsCopy := make(map[int]map[uint16]float64)
		for dID, sMap := range state.DeviceSensors {
			sensorsCopy[dID] = make(map[uint16]float64)
			for addr, val := range sMap {
				sensorsCopy[dID][addr] = val
			}
		}
		state.DataMu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"devices":   sensorsCopy,
			"timestamp": time.Now().Format("15:04:05"),
		})
	}
}

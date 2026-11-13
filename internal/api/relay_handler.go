package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"hardware-monitoring-control-platform/internal/core"
)

// sistemde tanımlı olan röle konfigürasyonlarını HTTP üzerinden dönme.
func HandleGetRelays(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state.AppConfig.Relays)
	}
}

// röle kontrol HTTP isteğini karşılama, CanOperate kuralını denetleme ve Modbus'a yazma.
func HandlePostRelay(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "İzin verilmeyen HTTP metodu", http.StatusMethodNotAllowed)
			return
		}

		var req RelayRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Geçersiz istek gövdesi (JSON parse hatası): "+err.Error(), http.StatusBadRequest)
			return
		}

		var targetRelay *core.RelayConfig
		for _, rel := range state.AppConfig.Relays {
			if rel.ID == req.ID {
				targetRelay = &rel
				break
			}
		}

		if targetRelay == nil {
			http.Error(w, "Belirtilen ID'ye sahip röle bulunamadı", http.StatusNotFound)
			return
		}

		if !state.CanOperate() {
			http.Error(w, "Sistem şu anda işlem yapmaya uygun değil: Modbus bağlantısı aktif değil", http.StatusServiceUnavailable)
			return
		}

		ac := state.SafeClient.Get()

		state.SafeClient.Mu.Lock()

		err := ac.WriteCoil(targetRelay.CoilAddress, req.State)
		if err != nil {
			state.SafeClient.Mu.Unlock()
			http.Error(w, "Modbus cihaza yazma başarısız: "+err.Error(), http.StatusInternalServerError)
			return
		}

		actualState, err := ac.ReadCoils(targetRelay.CoilAddress, 1)
		state.SafeClient.Mu.Unlock()

		if err != nil {
			http.Error(w, "Donanım durum doğrulaması (Read-back) başarısız: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if actualState[0] != req.State {
			http.Error(w, "Donanım alarmı: Röle mekanik olarak sıkıştı, komut uygulanamadı!", http.StatusConflict)
			return
		}

		state.RelayMu.Lock()
		state.RelayStates[req.ID] = req.State
		state.RelayMu.Unlock()

		fmt.Printf("\n[HTTP KONTROL BAŞARILI] Röle #%d (%s) -> Fiziksel olarak doğrulandı: %v\n> ", req.ID, targetRelay.Name, req.State)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      req.ID,
			"state":   req.State,
		})
	}
}

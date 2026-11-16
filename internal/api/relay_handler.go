package api

import (
	"encoding/json"
	"fmt"
	"hardware-monitoring-control-platform/internal/core"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type RelayRequest struct {
	DeviceID int  `json:"device_id"`
	ID       int  `json:"id"`
	State    bool `json:"state"`
}

func HandleGetRelays(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var allRelays []core.RelayConfig
		for _, dev := range state.AppConfig.Devices {
			for _, r := range dev.Relays {
				allRelays = append(allRelays, r)
			}
		}
		json.NewEncoder(w).Encode(allRelays)
	}
}
func HandlePostRelay(state *core.ServerState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "İzin verilmeyen HTTP metodu", http.StatusMethodNotAllowed)
			return
		}

		// --- 1. JWT & YETKİ KONTROLÜ ---
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Yetkilendirme başlığı eksik", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Geçersiz token formatı", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Geçersiz veya süresi dolmuş token", http.StatusUnauthorized)
			return
		}

		var req RelayRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Geçersiz JSON gövdesi: "+err.Error(), http.StatusBadRequest)
			return
		}

		// --- 2. DİNAMİK CİHAZ VE RÖLE YETKİ KONTROLÜ ---
		hasPermission := false
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if isSuper, ok := claims["is_superuser"].(bool); ok && isSuper {
				hasPermission = true
			} else if perms, ok := claims["permissions"].([]interface{}); ok {
				requiredPerm := fmt.Sprintf("accounts.device%d_relay%d_change", req.DeviceID, req.ID)
				for _, p := range perms {
					if pStr, ok := p.(string); ok && pStr == requiredPerm {
						hasPermission = true
						break
					}
				}
			}
		}

		if !hasPermission {
			http.Error(w, fmt.Sprintf("Bu röleyi değiştirmek için yetkiniz yok (Gerekli: accounts.device%d_relay%d_change)", req.DeviceID, req.ID), http.StatusForbidden)
			return
		}

		// --- 3. HEDEF CİHAZ VE RÖLEYİ BULMA ---
		var targetDevice *core.DeviceConfig
		for _, dev := range state.AppConfig.Devices {
			if dev.ID == req.DeviceID {
				targetDevice = &dev
				break
			}
		}

		if targetDevice == nil {
			http.Error(w, "Belirtilen ID'ye sahip cihaz bulunamadı", http.StatusNotFound)
			return
		}

		var targetRelay *core.RelayConfig
		for _, rel := range targetDevice.Relays {
			if rel.ID == req.ID {
				targetRelay = &rel
				break
			}
		}

		if targetRelay == nil {
			http.Error(w, "Belirtilen cihazda bu ID'ye sahip röle bulunamadı", http.StatusNotFound)
			return
		}

		// --- 4. MODBUS BAĞLANTI VE YAZMA İŞLEMİ ---
		safeClient := state.GetClient(req.DeviceID)
		if safeClient == nil || safeClient.Get() == nil {
			http.Error(w, "Cihaz için aktif Modbus bağlantısı yok", http.StatusServiceUnavailable)
			return
		}

		ac := safeClient.Get()
		safeClient.Mu.Lock()
		err = ac.WriteCoil(targetRelay.CoilAddress, req.State)
		if err != nil {
			safeClient.Mu.Unlock()
			http.Error(w, "Modbus yazma başarısız: "+err.Error(), http.StatusInternalServerError)
			return
		}

		actualState, err := ac.ReadCoils(targetRelay.CoilAddress, 1)
		safeClient.Mu.Unlock()

		if err != nil {
			http.Error(w, "Donanım durum doğrulaması başarısız: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if actualState[0] != req.State {
			http.Error(w, "Donanım alarmı: Röle mekanik olarak sıkıştı!", http.StatusConflict)
			return
		}

		// State güncelleme
		state.RelayMu.Lock()
		if state.RelayStates[req.DeviceID] == nil {
			state.RelayStates[req.DeviceID] = make(map[int]bool)
		}
		state.RelayStates[req.DeviceID][req.ID] = req.State
		state.RelayMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"device_id": req.DeviceID,
			"id":        req.ID,
			"state":     req.State,
		})
	}
}

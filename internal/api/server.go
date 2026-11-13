package api

import (
	"net/http"
)

// istemciden gelen röle kontrol komutlarının HTTP veri yapısıdır.
type RelayRequest struct {
	ID    int  `json:"id"`
	State bool `json:"state"`
}

// CORS politikalarını uygulayan ara yazılımdır.
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

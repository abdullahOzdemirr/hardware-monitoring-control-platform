package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleLoginProxy, kullanıcı giriş isteklerini Django'ya yönlendirir
func HandleLoginProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "İzin verilmeyen HTTP metodu", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON gövdesi", http.StatusBadRequest)
		return
	}

	// Django servisinin URL'si (Docker Compose kullanıyorsan servis adıyla yazılabilir)
	// Django servisinin URL'si (Yerel test için)
	djangoURL := "http://127.0.0.1:8000/api/auth/login/"

	jsonData, _ := json.Marshal(req)
	resp, err := http.Post(djangoURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Django kimlik doğrulama servisine ulaşılamadı: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Django'dan dönen yanıtı (JWT veya hata mesajı) doğrudan istemciye (React) ilet
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

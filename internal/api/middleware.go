package api

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"hardware-monitoring-control-platform/internal/metrics"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var JwtSecret []byte

func init() {
	_ = godotenv.Load()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("KRİTİK HATA: .env dosyasında JWT_SECRET tanımlanmamış! Django ile Go haberleşemez.")
	}
	JwtSecret = []byte(secret)
}

// Status kodunu yakalayabilmek için ResponseWriter sarmalayıcısı
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORS izinleri
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// --- 1. PROMETHEUS ---
func PrometheusMiddleware(endpoint string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(rw, r)

		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(rw.statusCode)

		metrics.HttpRequestsTotal.WithLabelValues(r.Method, endpoint, statusStr).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
	}
}

// korumalı endpoint'lere gelen JWT'leri denetleme
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Yetkilendirme başlığı (Authorization header) eksik", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Geçersiz token formatı. Format: Bearer <token>", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Geçersiz veya süresi dolmuş token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

/*
func RequirePermission(requiredPermission string, next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		tokenStr := parts[1]

		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		})

		// Token içerisinden kullanıcı adını da alalım (Loglama için gerekli)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Geçersiz Token Claims", http.StatusUnauthorized)
			return
		}

		username, _ := claims["username"].(string) // Token'daki kullanıcı adı alanı (anahtarı projene göre değişebilir)
		if username == "" {
			username = "BilinmeyenKullanici"
		}

		// 1. Durum: Süper kullanıcı kontrolü
		if isSuper, ok := claims["is_superuser"].(bool); ok && isSuper {
			pkg.LogAction(username, "RELAY_PERMISSION_GRANTED", fmt.Sprintf("Superuser yetkisiyle izin verildi. Gerekli: %s", requiredPermission))
			next(w, r)
			return
		}

		// 2. Durum: İzin listesi kontrolü
		if perms, ok := claims["permissions"].([]interface{}); ok {
			for _, p := range perms {
				if pStr, ok := p.(string); ok && pStr == requiredPermission {
					pkg.LogAction(username, "RELAY_PERMISSION_GRANTED", fmt.Sprintf("İzin eşleşti: %s", requiredPermission))

					next(w, r)
					return
				}
			}
		}

		http.Error(w, "Bu röleyi değiştirmek için yetkiniz yok (Forbidden)", http.StatusForbidden)
	})
}
*/

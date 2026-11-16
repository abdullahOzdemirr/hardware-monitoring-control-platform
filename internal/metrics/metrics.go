package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// --- 1. MODBUS & DONANIM METRİKLERİ ---

	// Modbus TCP bağlantı durumu (1: Bağlı, 0: Kopuk)
	ModbusConnectionStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "scada_modbus_connection_status",
			Help: "Modbus TCP donanım bağlantı durumu (1: Bağlı, 0: Kopuk).",
		},
	)

	// Toplam Modbus hata sayısı
	ModbusErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "scada_modbus_errors_total",
			Help: "Modbus haberleşmesi sırasında oluşan toplam hata sayısı.",
		},
	)

	// Modbus yanıt süresi (Gecikme - Latency)
	ModbusLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "scada_modbus_latency_seconds",
			Help:    "Modbus cihazlarından veri okuma/yazma gecikme süresi.",
			Buckets: prometheus.DefBuckets,
		},
	)

	// --- 2. RÖLE & KONTROL METRİKLERİ ---

	// Hangi röle başarıyla veya hatayla tetiklendi?
	RelayTriggerCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scada_relay_triggers_total",
			Help: "Toplam röle tetikleme komut sayısı.",
		},
		[]string{"relay_id", "status"},
	)

	// --- 3. API GATEWAY & HTTP METRİKLERİ ---

	// Gelen toplam HTTP istek sayısı
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scada_http_requests_total",
			Help: "API Gateway'e gelen toplam HTTP istek sayısı.",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP istek süreleri (Latency)
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scada_http_request_duration_seconds",
			Help:    "API endpoint yanıt süreleri (saniye).",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

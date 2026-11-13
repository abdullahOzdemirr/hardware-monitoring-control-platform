package core

import (
	"sync"

	"hardware-monitoring-control-platform/internal/client"
)

// konfigürasyon dosyasından okunan röle kanal parametrelerini tanımlama.
type RelayConfig struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CoilAddress uint16 `json:"coil_address"`
}

// tüm sistem konfigürasyon yapısını temsil etme.
type Config struct {
	DeviceURL string        `json:"device_url"`
	Relays    []RelayConfig `json:"relays"`
}

// eşzamanlı çalışan goroutine'ler arası veri paylaşımını ve donanım state'ini yönetme.
type ServerState struct {
	AppConfig   Config
	SafeClient  *client.SafeClient
	LatestTemp  float64
	LatestHum   float64
	DataMu      sync.RWMutex
	RelayStates map[int]bool
	RelayMu     sync.Mutex
}

func (s *ServerState) CanOperate() bool {
	ac := s.SafeClient.Get()
	if ac == nil {
		return false
	}
	// buraya acil stop, bakım modu vb. ek iş kuralları ekleyelim.
	return true
}

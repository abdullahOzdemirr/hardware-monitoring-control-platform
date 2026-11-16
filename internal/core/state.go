package core

import (
	"hardware-monitoring-control-platform/internal/client"
	"sync"
)

type SensorConfig struct {
	Name    string `json:"name"`
	Address uint16 `json:"address"`
	Unit    string `json:"unit"`
}

type RelayConfig struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CoilAddress uint16 `json:"coil_address"`
}

type DeviceConfig struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	URL     string         `json:"url"`
	Sensors []SensorConfig `json:"sensors"`
	Relays  []RelayConfig  `json:"relays"`
}

type Config struct {
	Devices []DeviceConfig `json:"devices"`
}

type ServerState struct {
	AppConfig Config

	Clients  map[int]*client.SafeClient
	ClientMu sync.RWMutex

	DeviceSensors map[int]map[uint16]float64
	DataMu        sync.RWMutex

	RelayStates map[int]map[int]bool
	RelayMu     sync.Mutex
}

func (s *ServerState) GetClient(deviceID int) *client.SafeClient {
	s.ClientMu.RLock()
	defer s.ClientMu.RUnlock()
	return s.Clients[deviceID]
}

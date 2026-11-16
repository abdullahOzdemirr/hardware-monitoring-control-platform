package simulator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

// Config Yapıları
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

type SensorHandler struct {
	Mu         sync.RWMutex
	DeviceName string
	Sensors    map[uint16]uint16 // Adres -> Değer
	Coils      map[uint16]bool   // Coil Adresi -> Durum (Açık/Kapalı)
	FaultMode  string            // "normal", "timeout", "error"
}

func NewSensorHandler(deviceName string, sensors []SensorConfig, relays []RelayConfig) *SensorHandler {
	sensorMap := make(map[uint16]uint16)

	// Başlangıç değerleri
	for _, s := range sensors {
		switch s.Address {
		case 0:
			sensorMap[s.Address] = 2450
		case 1:
			sensorMap[s.Address] = 5500
		default:
			sensorMap[s.Address] = 1000
		}
	}

	coilMap := make(map[uint16]bool)
	for _, r := range relays {
		coilMap[r.CoilAddress] = false
	}

	return &SensorHandler{
		DeviceName: deviceName,
		Sensors:    sensorMap,
		Coils:      coilMap,
		FaultMode:  "normal",
	}
}

func (h *SensorHandler) UpdateRandomValues() {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	for addr := range h.Sensors {
		val := int16(h.Sensors[addr])
		val += int16(rand.Intn(41) - 20)
		if val < 0 {
			val = 0
		}
		h.Sensors[addr] = uint16(val)
	}
}

func (h *SensorHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) ([]uint16, error) {
	h.Mu.RLock()
	mode := h.FaultMode
	h.Mu.RUnlock()

	if mode == "timeout" {
		fmt.Printf("[%s] İstek kasıtlı olarak geciktiriliyor...\n", h.DeviceName)
		time.Sleep(3 * time.Second)
	}

	if mode == "error" {
		fmt.Printf("[%s] Kasıtlı hata yanıtı döndürülüyor.\n", h.DeviceName)
		return nil, modbus.ErrIllegalDataAddress
	}

	h.Mu.RLock()
	defer h.Mu.RUnlock()

	var res []uint16
	for i := uint16(0); i < req.Quantity; i++ {
		addr := req.Addr + i
		val, exists := h.Sensors[addr]
		if !exists {
			return nil, modbus.ErrIllegalDataAddress
		}
		res = append(res, val)
	}
	return res, nil
}

func (h *SensorHandler) HandleCoils(req *modbus.CoilsRequest) ([]bool, error) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if req.IsWrite {
		for i := uint16(0); i < req.Quantity; i++ {
			addr := req.Addr + i
			_, exists := h.Coils[addr]
			if !exists {
				return nil, modbus.ErrIllegalDataAddress
			}
			targetVal := req.Args[0]

			// Donanımda arıza çıkma simülasyonu (%30 ihtimalle adres 1 takılır)
			if addr == 1 && rand.Float32() < 0.3 {
				fmt.Printf("[%s - Donanım Arızası] Röle adresi #%d mekanik olarak sıkıştı, komut uygulanamadı!\n", h.DeviceName, addr)
				continue
			}

			h.Coils[addr] = targetVal
			stateStr := "KAPATILDI (False)"
			if targetVal {
				stateStr = "AÇILDI (True)"
			}
			fmt.Printf("[%s - Fiziksel Donanım] Röle Adresi #%d -> %s\n", h.DeviceName, addr, stateStr)
		}
	}

	var res []bool
	for i := uint16(0); i < req.Quantity; i++ {
		addr := req.Addr + i
		val, exists := h.Coils[addr]
		if !exists {
			return nil, modbus.ErrIllegalDataAddress
		}
		res = append(res, val)
	}
	return res, nil
}

func (h *SensorHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) ([]bool, error) {
	return nil, modbus.ErrIllegalFunction
}

func (h *SensorHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) ([]uint16, error) {
	return nil, modbus.ErrIllegalFunction
}

package simulator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

type SensorHandler struct {
	Mu          sync.RWMutex
	Temperature uint16
	Humidity    uint16
	Coils       [4]bool
	FaultMode   string // "normal", "timeout", "error"
}

func NewSensorHandler() *SensorHandler {
	return &SensorHandler{
		Temperature: 2450,
		Humidity:    5500,
		FaultMode:   "normal",
	}
}

func (h *SensorHandler) HandleInputRegisters(req *modbus.InputRegistersRequest) ([]uint16, error) {
	h.Mu.RLock()
	mode := h.FaultMode
	h.Mu.RUnlock()

	if mode == "timeout" {
		fmt.Println("[Simülatör] İstek kasıtlı olarak geciktiriliyor...")
		time.Sleep(3 * time.Second)
	}

	if mode == "error" {
		fmt.Println("[Simülatör] Kasıtlı hata yanıtı döndürülüyor.")
		return nil, modbus.ErrIllegalDataAddress
	}

	h.Mu.RLock()
	defer h.Mu.RUnlock()

	var res []uint16
	for i := uint16(0); i < req.Quantity; i++ {
		addr := req.Addr + i
		switch addr {
		case 0:
			res = append(res, h.Temperature)
		case 1:
			res = append(res, h.Humidity)
		default:
			return nil, modbus.ErrIllegalDataAddress
		}
	}
	return res, nil
}

func (h *SensorHandler) HandleCoils(req *modbus.CoilsRequest) ([]bool, error) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if req.IsWrite {
		for i := uint16(0); i < req.Quantity; i++ {
			addr := req.Addr + i
			if addr >= 4 {
				return nil, modbus.ErrIllegalDataAddress
			}
			targetVal := req.Args[0]
			if addr == 1 && rand.Float32() < 0.3 {
				fmt.Printf("[Simülatör Donanım Arızası] Isıtıcı rölesi mekanik olarak sıkıştı, komut uygulanamadı!\n")
				continue
			}
			h.Coils[addr] = targetVal
			stateStr := "KAPATILDI (False)"
			if targetVal {
				stateStr = "AÇILDI (True)"
			}
			fmt.Printf("[Simülatör Fiziksel Donanım] Röle Kanalı #%d -> %s\n", addr, stateStr)
		}
	}

	var res []bool
	for i := uint16(0); i < req.Quantity; i++ {
		addr := req.Addr + i
		if addr >= 4 {
			return nil, modbus.ErrIllegalDataAddress
		}
		res = append(res, h.Coils[addr])
	}
	return res, nil
}

func (h *SensorHandler) HandleDiscreteInputs(req *modbus.DiscreteInputsRequest) ([]bool, error) {
	return nil, modbus.ErrIllegalFunction
}

func (h *SensorHandler) HandleHoldingRegisters(req *modbus.HoldingRegistersRequest) ([]uint16, error) {
	return nil, modbus.ErrIllegalFunction
}

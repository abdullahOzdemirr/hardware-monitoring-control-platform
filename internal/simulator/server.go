package simulator

import (
	"time"

	"github.com/simonvetter/modbus"
)

func NewModbusServer(url string, handler *SensorHandler) (*modbus.ModbusServer, error) {
	return modbus.NewServer(&modbus.ServerConfiguration{
		URL:        url,
		Timeout:    30 * time.Second,
		MaxClients: 1,
	}, handler)
}

package client

import (
	"sync"

	"github.com/simonvetter/modbus"
)

type SafeClient struct {
	Mu     sync.Mutex
	Client *modbus.ModbusClient
}

func (sc *SafeClient) SetClient(c *modbus.ModbusClient) {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()
	sc.Client = c
}

func (sc *SafeClient) Get() *modbus.ModbusClient {
	sc.Mu.Lock()
	defer sc.Mu.Unlock()
	return sc.Client
}

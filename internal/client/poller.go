package client

import (
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

func ConnectWithRetry(url string) (*modbus.ModbusClient, error) {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     url,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	if err := client.Open(); err != nil {
		return nil, err
	}
	return client, nil
}

func ReadWithRetry(client *modbus.ModbusClient, quantity uint16, maxRetries int) ([]uint16, error) {
	backoff := 500 * time.Millisecond
	for attempt := 1; attempt <= maxRetries; attempt++ {
		regs, err := client.ReadRegisters(0, quantity, modbus.INPUT_REGISTER)
		if err == nil {
			return regs, nil
		}
		fmt.Printf("\n[OKUMA HATASI] Deneme %d/%d başarısız: %v\n", attempt, maxRetries, err)
		if attempt < maxRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return nil, fmt.Errorf("max retry limitine ulaşıldı")
}

package poller

import (
	"fmt"
	"hardware-monitoring-control-platform/internal/client"
	"hardware-monitoring-control-platform/internal/core"
	"hardware-monitoring-control-platform/internal/pkg"
	"time"
)

func StartPoller(state *core.ServerState, done <-chan struct{}) {
	for _, dev := range state.AppConfig.Devices {
		go pollDevice(dev, state, done)
	}
}

func pollDevice(dev core.DeviceConfig, state *core.ServerState, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
		}

		c, err := client.ConnectWithRetry(dev.URL)
		if err != nil {
			fmt.Printf("[POLLER - %s] Bağlantı kurulamadı, 3 saniye sonra tekrar denenecek...\n", dev.Name)
			pkg.LogAction("", "", "Bağlantı kurulamadı")
			time.Sleep(3 * time.Second)
			continue
		}

		safeClient := state.GetClient(dev.ID)
		safeClient.SetClient(c)
		fmt.Printf("[POLLER - %s] Bağlantı aktif, veriler okunuyor...\n", dev.Name)

		ticker := time.NewTicker(2 * time.Second)
		connected := true

		for connected {
			select {
			case <-done:
				ticker.Stop()
				c.Close()
				return
			case <-ticker.C:
				ac := safeClient.Get()
				if ac == nil {
					connected = false
					break
				}

				sensorCount := uint16(len(dev.Sensors))
				regs, err := client.ReadWithRetry(ac, sensorCount, 3)
				if err != nil {
					fmt.Printf("\n[HATA - %s] Okuma başarısız, soket kapatılıyor...\n", dev.Name)
					ac.Close()
					safeClient.SetClient(nil)
					connected = false
					break
				}

				state.DataMu.Lock()
				if state.DeviceSensors[dev.ID] == nil {
					state.DeviceSensors[dev.ID] = make(map[uint16]float64)
				}

				for i, sensor := range dev.Sensors {
					if i < len(regs) {
						state.DeviceSensors[dev.ID][sensor.Address] = float64(regs[i]) / 100.0
					}
				}
				state.DataMu.Unlock()
			}
		}
		ticker.Stop()
		time.Sleep(3 * time.Second)
	}
}

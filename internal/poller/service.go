package poller

import (
	"fmt"
	"time"

	"hardware-monitoring-control-platform/internal/client"
	"hardware-monitoring-control-platform/internal/core"
)

// arka planda periyodik olarak Modbus cihazından sensör verilerini okur.
func StartPoller(state *core.ServerState, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
		}

		c, err := client.ConnectWithRetry(state.AppConfig.DeviceURL)
		if err != nil {
			fmt.Printf("[POLAR] Bağlantı kurulamadı, 3 saniye sonra tekrar denenecek...\n")
			time.Sleep(3 * time.Second)
			continue
		}

		state.SafeClient.SetClient(c)
		fmt.Println("[POLAR] Bağlantı aktif, sensör verileri okunmaya başlanıyor...")

		ticker := time.NewTicker(2 * time.Second)
		connected := true

		for connected {
			select {
			case <-done:
				ticker.Stop()
				c.Close()
				return
			case <-ticker.C:
				ac := state.SafeClient.Get()
				if ac == nil {
					connected = false
					break
				}

				regs, err := client.ReadWithRetry(ac, 3)
				if err != nil {
					fmt.Printf("\n[HATA] Okuma başarısız, soket kapatılıyor...\n")
					ac.Close()
					state.SafeClient.SetClient(nil)
					connected = false
					break
				}

				temp := float64(regs[0]) / 100.0
				hum := float64(regs[1]) / 100.0

				state.DataMu.Lock()
				state.LatestTemp = temp
				state.LatestHum = hum
				state.DataMu.Unlock()

				fmt.Printf("\r[%s] [POLLED] Sıcaklık: %.2f °C | Nem: %.2f %%\n> ", time.Now().Format("15:04:05"), temp, hum)
			}
		}

		ticker.Stop()
		time.Sleep(3 * time.Second)
	}
}

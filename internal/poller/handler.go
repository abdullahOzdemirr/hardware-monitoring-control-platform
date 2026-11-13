package poller

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"hardware-monitoring-control-platform/internal/core"
)

// terminal girişlerini dinleyerek anlık röle kontrol komutlarını işler.
func StartCLI(state *core.ServerState, done chan<- struct{}) {
	time.Sleep(500 * time.Millisecond) // Logların çakışmaması için kısa gecikme
	fmt.Print("\n[CLI Aktif] Röle komutu için örnek: '0 1' (Aç) veya '0 0' (Kapat). Çıkış için 'exit'\n> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "exit" || text == "quit" {
			close(done)
			return
		}

		var relayID int
		var stateInt int
		_, err := fmt.Sscanf(text, "%d %d", &relayID, &stateInt)
		if err != nil {
			fmt.Println("Geçersiz format! Örnek: '0 1' (Aç) veya '0 0' (Kapat)")
			fmt.Print("> ")
			continue
		}

		var targetRelay *core.RelayConfig
		for _, r := range state.AppConfig.Relays {
			if r.ID == relayID {
				targetRelay = &r
				break
			}
		}

		if targetRelay == nil {
			fmt.Printf("Hata: ID'si %d olan röle bulunamadı!\n> ", relayID)
			continue
		}

		targetState := (stateInt == 1)

		if !state.CanOperate() {
			fmt.Println("Hata: Şu anda aktif bir Modbus bağlantısı yok!\n> ")
			continue
		}

		ac := state.SafeClient.Get()

		state.SafeClient.Mu.Lock()
		err = ac.WriteCoil(targetRelay.CoilAddress, targetState)
		if err != nil {
			fmt.Printf("[YAZMA HATASI] Komut iletilemedi: %v\n> ", err)
			state.SafeClient.Mu.Unlock()
			continue
		}

		actualState, err := ac.ReadCoils(targetRelay.CoilAddress, 1)
		state.SafeClient.Mu.Unlock()

		if err != nil {
			fmt.Printf("[READ-BACK HATASI]: %v\n> ", err)
		} else if actualState[0] == targetState {
			state.RelayMu.Lock()
			state.RelayStates[relayID] = targetState
			state.RelayMu.Unlock()
			fmt.Printf("[ANLIK KONTROL BAŞARILI] '%s' -> %v (Fiziksel olarak doğrulandı)\n> ", targetRelay.Name, targetState)
		} else {
			fmt.Printf("[DONANIM ALARMI] '%s' durum değiştirilemedi!\n> ", targetRelay.Name)
		}

		fmt.Print("> ")
	}
}

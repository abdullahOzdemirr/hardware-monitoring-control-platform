package poller

import (
	"bufio"
	"fmt"
	"hardware-monitoring-control-platform/internal/core"
	"os"
	"strings"
	"time"
)

func StartCLI(state *core.ServerState, done chan<- struct{}) {
	time.Sleep(500 * time.Millisecond)
	fmt.Print("\n[CLI Aktif] Format: '<CihazID> <RoleID> <Deger>' (Örn: '1 1 1' -> Cihaz 1, Röle 1, Aç). Çıkış için 'exit'\n> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "exit" || text == "quit" {
			close(done)
			return
		}

		var deviceID, relayID, stateInt int
		_, err := fmt.Sscanf(text, "%d %d %d", &deviceID, &relayID, &stateInt)
		if err != nil {
			fmt.Println("Geçersiz format! Örnek: '1 1 1' (Cihaz 1, Röle 1, Aç)")
			fmt.Print("> ")
			continue
		}

		var targetDevice *core.DeviceConfig
		for _, d := range state.AppConfig.Devices {
			if d.ID == deviceID {
				targetDevice = &d
				break
			}
		}

		if targetDevice == nil {
			fmt.Printf("Hata: ID'si %d olan cihaz bulunamadı!\n> ", deviceID)
			continue
		}

		var targetRelay *core.RelayConfig
		for _, r := range targetDevice.Relays {
			if r.ID == relayID {
				targetRelay = &r
				break
			}
		}

		if targetRelay == nil {
			fmt.Printf("Hata: %s cihazında ID'si %d olan röle bulunamadı!\n> ", targetDevice.Name, relayID)
			continue
		}

		targetState := (stateInt == 1)
		safeClient := state.GetClient(deviceID)
		if safeClient == nil || safeClient.Get() == nil {
			fmt.Println("Hata: Bu cihaz için aktif bir Modbus bağlantısı yok!\n> ")
			continue
		}

		ac := safeClient.Get()
		safeClient.Mu.Lock()
		err = ac.WriteCoil(targetRelay.CoilAddress, targetState)
		if err != nil {
			fmt.Printf("[YAZMA HATASI] Komut iletilemedi: %v\n> ", err)
			safeClient.Mu.Unlock()
			continue
		}

		actualState, err := ac.ReadCoils(targetRelay.CoilAddress, 1)
		safeClient.Mu.Unlock()

		if err != nil {
			fmt.Printf("[READ-BACK HATASI]: %v\n> ", err)
		} else if actualState[0] == targetState {
			state.RelayMu.Lock()
			if state.RelayStates[deviceID] == nil {
				state.RelayStates[deviceID] = make(map[int]bool)
			}
			state.RelayStates[deviceID][relayID] = targetState
			state.RelayMu.Unlock()
			fmt.Printf("[BAŞARILI] %s -> '%s' -> %v\n> ", targetDevice.Name, targetRelay.Name, targetState)
		} else {
			fmt.Printf("[DONANIM ALARMI] '%s' durum değiştirilemedi!\n> ", targetRelay.Name)
		}
		fmt.Print("> ")
	}
}

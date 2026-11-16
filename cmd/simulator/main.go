package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"hardware-monitoring-control-platform/internal/simulator"
)

func main() {
	file, err := os.ReadFile("configs/config.json")
	if err != nil {
		log.Fatalf("Konfigürasyon dosyası okunamadı: %v", err)
	}

	var cfg simulator.Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		log.Fatalf("JSON parse hatası: %v", err)
	}

	fmt.Println("--- Çoklu Modbus Simülatör Merkezi Başlatılıyor ---")
	fmt.Println("---------------------------------------------------------")
	fmt.Println("HATA ENJEKSİYON KOMUTLARI (Tüm cihazlara uygulanır):")
	fmt.Println("   [n] Normal mod (Veriler normal akar)")
	fmt.Println("   [t] Timeout modu (Yanıt geciktirilir)")
	fmt.Println("   [e] Hata modu (Modbus hata kodu döndürülür)")
	fmt.Println("   [q] Çıkış")
	fmt.Println("---------------------------------------------------------")

	var handlers []*simulator.SensorHandler

	for _, dev := range cfg.Devices {
		handler := simulator.NewSensorHandler(dev.Name, dev.Sensors, dev.Relays)
		handlers = append(handlers, handler)

		listenURL := getListenURL(dev.URL)

		server, err := simulator.NewModbusServer(listenURL, handler)
		if err != nil {
			log.Fatalf("[%s] Modbus sunucusu oluşturulamadı: %v", dev.Name, err)
		}

		go func(deviceName, targetURL string) {
			if err := server.Start(); err != nil {
				log.Fatalf("[%s] Sunucu başlatılamadı (%s): %v", deviceName, targetURL, err)
			}
		}(dev.Name, listenURL)

		fmt.Printf("✔ %s Aktif -> Dinleniyor: %s (Tanımlı URL: %s)\n", dev.Name, listenURL, dev.URL)

		go func(h *simulator.SensorHandler) {
			for {
				time.Sleep(2 * time.Second)
				h.UpdateRandomValues()
			}
		}(handler)
	}

	// Konsoldan durumu değiştirme
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := strings.TrimSpace(strings.ToLower(scanner.Text()))
			for _, h := range handlers {
				h.Mu.Lock()
				switch cmd {
				case "n":
					h.FaultMode = "normal"
					fmt.Printf("\n[%s] Mod: NORMAL\n", h.DeviceName)
				case "t":
					h.FaultMode = "timeout"
					fmt.Printf("\n[%s] Mod: TIMEOUT\n", h.DeviceName)
				case "e":
					h.FaultMode = "error"
					fmt.Printf("\n[%s] Mod: HATA\n", h.DeviceName)
				case "q":
					fmt.Println("Çıkış yapılıyor...")
					os.Exit(0)
				}
				h.Mu.Unlock()
			}
		}
	}()

	// CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println("\nTüm simülatörler kapatıldı.")
}

func getListenURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "tcp://0.0.0.0:5020"
	}
	parts := strings.Split(u.Host, ":")
	port := "5020"
	if len(parts) > 1 {
		port = parts[1]
	}
	return fmt.Sprintf("tcp://0.0.0.0:%s", port)
}

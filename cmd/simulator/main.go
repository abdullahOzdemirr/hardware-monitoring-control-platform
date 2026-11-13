package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"hardware-monitoring-control-platform/internal/simulator"
)

func main() {
	handler := simulator.NewSensorHandler()

	server, err := simulator.NewModbusServer("tcp://0.0.0.0:5020", handler)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	defer server.Stop()

	fmt.Println("Sanal Sensör ve Röle Kartı Çalışıyor -> tcp://0.0.0.0:5020")
	fmt.Println("---------------------------------------------------------")
	fmt.Println("HATA ENJEKSİYON KOMUTLARI:")
	fmt.Println("   [n] Normal mod (Veriler normal akar)")
	fmt.Println("   [t] Timeout modu (Yanıt geciktirilir)")
	fmt.Println("   [e] Hata modu (Modbus hata kodu döndürülür)")
	fmt.Println("   [q] Çıkış")
	fmt.Println("---------------------------------------------------------")

	go func() {
		for {
			time.Sleep(2 * time.Second)
			handler.Mu.Lock()
			handler.Temperature = uint16(int16(handler.Temperature) + int16(rand.Intn(41)-20))
			handler.Humidity = uint16(int16(handler.Humidity) + int16(rand.Intn(61)-30))
			handler.Mu.Unlock()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := strings.TrimSpace(strings.ToLower(scanner.Text()))
			handler.Mu.Lock()
			switch cmd {
			case "n":
				handler.FaultMode = "normal"
				fmt.Println("\nMod: NORMAL (Veriler normal akıyor)")
			case "t":
				handler.FaultMode = "timeout"
				fmt.Println("\nMod: TIMEOUT (İstekler geciktiriliyor)")
			case "e":
				handler.FaultMode = "error"
				fmt.Println("\nMod: HATA (Geçersiz yanıtlar üretiliyor)")
			case "q":
				fmt.Println("Çıkış yapılıyor...")
				os.Exit(0)
			}
			handler.Mu.Unlock()
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println("\nSimülatör kapatıldı.")
}

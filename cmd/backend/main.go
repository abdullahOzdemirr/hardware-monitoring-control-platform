package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hardware-monitoring-control-platform/internal/api"
	"hardware-monitoring-control-platform/internal/client"
	"hardware-monitoring-control-platform/internal/core"
	"hardware-monitoring-control-platform/internal/poller"
)

func main() {
	file, err := os.ReadFile("configs/config.json")
	if err != nil {
		log.Fatalf("Konfigürasyon dosyası okunamadı: %v", err)
	}

	var cfg core.Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		log.Fatalf("JSON parse hatası: %v", err)
	}

	state := &core.ServerState{
		AppConfig:   cfg,
		SafeClient:  &client.SafeClient{},
		RelayStates: make(map[int]bool),
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Backend servisleri başlatılıyor...")
	done := make(chan struct{})
	var wg sync.WaitGroup

	// 1. HTTP API Sunucusunu Başlat
	wg.Add(1)
	go func() {
		defer wg.Done()
		router := api.NewRouter(state)
		fmt.Println("HTTP API Sunucusu aktif -> http://localhost:8080")
		if err := http.ListenAndServe(":8080", router); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP Sunucu Hatası: %v", err)
		}
	}()

	// 2. Modbus Polling Servisini Başlat
	wg.Add(1)
	go func() {
		defer wg.Done()
		poller.StartPoller(state, done)
	}()

	// 3. Terminal CLI Servisini Başlat (React var ama bura test amaçlı)
	wg.Add(1)
	go func() {
		defer wg.Done()
		poller.StartCLI(state, done)
	}()

	// 4. Kapanış Sinyalini Bekle
	<-sigChan
	fmt.Println("\nSistem kapatılıyor...")
	close(done)
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Güvenli çıkış yapıldı.")
}

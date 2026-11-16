package pkg

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex

func LogAction(username string, actionType string, details string) {
	mu.Lock()
	defer mu.Unlock()

	if username == "" {
		username = "Sistem / Anonim"
	}

	if actionType == "" {
		actionType = "Sistem"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] | KULLANICI: %s | İŞLEM: %s | DETAY: %s\n", timestamp, username, actionType, details)

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Log dosyası açılamadı: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Log dosyasına yazılamadı: %v", err)
	}
}

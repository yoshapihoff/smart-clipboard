package main

import (
	"log"
	"time"

	"github.com/yoshapihoff/smart-clipboard/internal/clipboard"
	"github.com/yoshapihoff/smart-clipboard/internal/config"
	"github.com/yoshapihoff/smart-clipboard/internal/storage"
	"github.com/yoshapihoff/smart-clipboard/internal/tray"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Ошибка загрузки конфигурации: %v", err)
		cfg = config.DefaultConfig()
	}

	store, err := storage.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("Ошибка инициализации хранилища: %v", err)
	}

	history, err := store.LoadHistory()
	if err != nil {
		log.Printf("Ошибка загрузки истории: %v", err)
	}

	clipboardManager := clipboard.NewManager(history, cfg.MaxItems)

	go monitorClipboard(clipboardManager, store, cfg.CheckInterval)
	tray.RunTray(clipboardManager, store, cfg)
}

func monitorClipboard(manager *clipboard.Manager, store *storage.Storage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		content, err := clipboard.GetClipboard()
		if err != nil {
			log.Printf("Ошибка чтения буфера обмена: %v", err)
			continue
		}

		if content != "" {
			manager.AddToHistory(content)
			storeErr := store.SaveHistory(manager.GetHistory())
			if storeErr != nil {
				log.Printf("Ошибка сохранения истории: %v", storeErr)
			}
		}
	}
}

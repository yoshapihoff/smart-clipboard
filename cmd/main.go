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
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Ошибка загрузки конфигурации: %v", err)
		cfg = config.DefaultConfig()
	}

	// Инициализация хранилища
	store, err := storage.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("Ошибка инициализации хранилища: %v", err)
	}

	// Загрузка истории
	history, err := store.LoadHistory()
	if err != nil {
		log.Printf("Ошибка загрузки истории: %v", err)
	}

	// Создание менеджера буфера обмена
	clipboardManager := clipboard.NewManager(history, cfg.MaxHistorySize)

	// Запуск системного трея
	tray.RunTray(clipboardManager, store, cfg)

	// Мониторинг буфера обмена
	go monitorClipboard(clipboardManager, cfg.CheckInterval)

	// Ожидание завершения
	select {}
}

func monitorClipboard(manager *clipboard.Manager, interval time.Duration) {
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
		}
	}
}

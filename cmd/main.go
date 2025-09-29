package main

import (
	"log"
	"time"

	"github.com/yoshapihoff/smart-clipboard/internal/clipboard"
	"github.com/yoshapihoff/smart-clipboard/internal/config"
	"github.com/yoshapihoff/smart-clipboard/internal/storage"
	"github.com/yoshapihoff/smart-clipboard/internal/sync"
	"github.com/yoshapihoff/smart-clipboard/internal/tray"
	"github.com/yoshapihoff/smart-clipboard/internal/types"
)

func main() {
	// Загружаем базовую конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Ошибка загрузки конфигурации: %v", err)
		cfg = config.DefaultConfig()
	}

	// Теперь создаем хранилище и загружаем локальные данные
	store, err := storage.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("Ошибка инициализации хранилища: %v", err)
	}

	// Загружаем локальную историю
	localHistory, err := store.LoadHistory()
	if err != nil {
		log.Printf("Ошибка загрузки локальной истории: %v", err)
	}

	// Настраиваем синхронизацию
	historyChan := make(chan []types.ClipboardItem, 10)
	syncManager, err := sync.NewSyncManager(historyChan)
	if err != nil {
		log.Printf("Ошибка инициализации синхронизации: %v", err)
	}

	// Создаем менеджер буфера обмена с финальной историей
	clipboardManager := clipboard.NewManager(localHistory, cfg.MaxItems, syncManager)

	// Устанавливаем callback для получения текущей истории
	syncManager.SetHistoryCallback(func() []types.ClipboardItem {
		return clipboardManager.GetHistory()
	})

	go handleSyncMessages(clipboardManager, store, historyChan)

	go monitorClipboard(clipboardManager, store, cfg.CheckInterval)
	tray.RunTray(clipboardManager, store, cfg)
}

func monitorClipboard(manager *clipboard.Manager, store *storage.Storage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Сначала читаем текущее содержимое буфера обмена и устанавливаем его как последнее
	initialContent, err := clipboard.GetClipboard()
	if err != nil {
		log.Printf("Ошибка начального чтения буфера обмена: %v", err)
	} else if initialContent != "" {
		manager.SetLastContent(initialContent)
		log.Printf("Initial clipboard content set: %s", initialContent[:min(20, len(initialContent))])
	}

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

func handleSyncMessages(manager *clipboard.Manager, store *storage.Storage, historyChan <-chan []types.ClipboardItem) {
	for history := range historyChan {
		log.Printf("Received %d history items via sync", len(history))
		manager.ReplaceHistory(history)
		err := store.SaveHistory(history)
		if err != nil {
			log.Printf("Error saving synced history: %v", err)
		}
	}
}

//go:build cgo
// +build cgo

package tray

import (
	"encoding/base64"
	"fmt"
	"log"

	"fyne.io/systray"
	"github.com/gen2brain/beeep"
	"github.com/yoshapihoff/smart-clipboard/internal/clipboard"
	"github.com/yoshapihoff/smart-clipboard/internal/config"
	"github.com/yoshapihoff/smart-clipboard/internal/storage"
)

var trayIcon []byte
var menuItemPool []*systray.MenuItem
var menuCancelChannels []chan struct{}

func initMenuItemPool(size int) {
	menuItemPool = make([]*systray.MenuItem, size)
	menuCancelChannels = make([]chan struct{}, size)
	for i := 0; i < size; i++ {
		menuItem := systray.AddMenuItem("", "")
		menuItem.Hide()
		menuItemPool[i] = menuItem
		menuCancelChannels[i] = make(chan struct{})
	}
}

func RunTray(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	systray.Run(onReady(manager, store, cfg), onExit(store))
}

func onReady(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) func() {
	return func() {
		systray.SetIcon(getIcon())
		// Формируем подсказку с учетом платформы
		tooltip := "Smart clipboard"
		systray.SetTooltip(tooltip)

		// Инициализируем пул элементов меню для истории
		initMenuItemPool(cfg.MaxItems)

		systray.AddSeparator()

		settingsMenu := systray.AddMenuItem("Settings", "Open settings")
		clearMenu := systray.AddMenuItem("Clear history", "Clear all history")
		systray.AddSeparator()
		quitMenu := systray.AddMenuItem("Quit", "Quit program")

		systray.AddSeparator()

		go func() {
			for range systray.TrayOpenedCh {
				rebuildHistoryMenu(manager, store, cfg)
			}
		}()

		go func() {
			for {
				select {
				case <-settingsMenu.ClickedCh:
					showSettingsWindow(cfg, store)
				case <-clearMenu.ClickedCh:
					manager.ClearClipboard()
					manager.ClearHistory()
					beeep.Notify("Smart clipboard", "History cleared", "")
				case <-quitMenu.ClickedCh:
					stopMenuHandlers() // Останавливаем все горутины перед выходом
					store.SaveHistory(manager.GetHistory())
					log.Println("Завершение работы приложения...")
					systray.Quit()
					return
				}
			}
		}()
	}
}

func onExit(store *storage.Storage) func() {
	return func() {
		stopMenuHandlers() // Останавливаем все горутины при выходе
		log.Println("Выход из приложения")
	}
}

func stopMenuHandlers() {
	for _, cancelChan := range menuCancelChannels {
		if cancelChan != nil {
			close(cancelChan)
		}
	}
}

func rebuildHistoryMenu(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	history := manager.GetHistory()
	// Останавливаем все старые горутины обработки кликов
	stopMenuHandlers()

	// Создаем новые каналы отмены
	for i := range menuCancelChannels {
		menuCancelChannels[i] = make(chan struct{})
	}

	// Скрываем все элементы из пула
	for _, menuItem := range menuItemPool {
		menuItem.Hide()
	}

	if len(history) == 0 {
		// Если история пуста, используем первый элемент из пула для сообщения
		if len(menuItemPool) > 0 {
			menuItemPool[0].SetTitle("History is empty")
			menuItemPool[0].SetTooltip("No clipboard history available")
			menuItemPool[0].Disable()
			menuItemPool[0].Show()
		}
		return
	}

	// Показываем элементы истории, используя пул
	itemsToShow := len(history)
	if itemsToShow > len(menuItemPool) {
		itemsToShow = len(menuItemPool)
	}

	for i := 0; i < itemsToShow; i++ {
		item := history[i]
		menuItem := menuItemPool[i]
		cancelChan := menuCancelChannels[i]

		var title string
		if cfg.DebugMode {
			title = fmt.Sprintf("[%d] %s", item.ClickCount, item.Preview)
		} else {
			title = item.Preview
		}

		menuItem.SetTitle(title)
		menuItem.SetTooltip(item.Timestamp.Format("2006-01-02 15:04:05"))
		menuItem.Enable()
		menuItem.Show()

		// Запускаем горутину для обработки кликов по элементу меню
		go func(menuItem *systray.MenuItem, clipboardItem clipboard.ClipboardItem, cancelChan chan struct{}) {
			for {
				select {
				case <-menuItem.ClickedCh:
					manager.CopyToClipboard(clipboardItem.Content)
					manager.IncrementClickCount(clipboardItem.Content)
					// После копирования пересобираем меню для обновления порядка
					rebuildHistoryMenu(manager, store, cfg)
					return // Выходим из горутины после обработки клика
				case <-cancelChan:
					return // Выходим из горутины при получении сигнала отмены
				}
			}
		}(menuItem, item, cancelChan)
	}
}

func showSettingsWindow(cfg *config.Config, store *storage.Storage) {

}

func getIcon() []byte {
	if len(trayIcon) > 0 {
		return trayIcon
	}

	var iconBase64 string
	isDark, err := isDarkTheme()
	if err != nil {
		log.Printf("tray: failed to check dark theme: %v", err)
		iconBase64 = iconBase64White
	} else if isDark {
		iconBase64 = iconBase64White
	} else {
		iconBase64 = iconBase64Dark
	}
	data, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil {
		log.Printf("tray: failed to decode base64 icon: %v", err)
	} else {
		trayIcon = data
		return trayIcon
	}
	return nil
}

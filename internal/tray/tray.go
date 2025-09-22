//go:build cgo
// +build cgo

package tray

import (
	"encoding/base64"
	"log"

	"github.com/yoshapihoff/smart-clipboard/internal/clipboard"
	"github.com/yoshapihoff/smart-clipboard/internal/config"
	"github.com/yoshapihoff/smart-clipboard/internal/storage"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

var trayIcon []byte

var historyItems []*systray.MenuItem

func RunTray(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	systray.Run(onReady(manager, store, cfg), onExit(store))
}

func onReady(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) func() {
	return func() {
		// Иконка в трее
		systray.SetIcon(getIcon())
		systray.SetTooltip("Smart clipboard")

		// Меню истории
		historyMenu := systray.AddMenuItem("History", "Show history")
		updateHistoryMenu(manager, historyMenu)

		// Разделитель
		systray.AddSeparator()

		// Настройки
		settingsMenu := systray.AddMenuItem("Settings", "Open settings")

		// Очистка истории
		clearMenu := systray.AddMenuItem("Clear history", "Clear all history")

		// Разделитель
		systray.AddSeparator()

		// Выход
		quitMenu := systray.AddMenuItem("Quit", "Quit program")

		// Обновление истории
		go func() {
			for {
				select {
				case <-historyMenu.ClickedCh:
					showHistoryWindow(manager)
				case <-settingsMenu.ClickedCh:
					showSettingsWindow(cfg, store)
				case <-clearMenu.ClickedCh:
					manager.ClearHistory()
					store.SaveHistory(manager.GetHistory())
					updateHistoryMenu(manager, historyMenu)
					beeep.Notify("Clipboard History", "History cleared", "")
				case <-quitMenu.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}
}

func onExit(store *storage.Storage) func() {
	return func() {
		// Сохранение данных при выходе
		log.Println("Выход из приложения")
	}
}

func updateHistoryMenu(manager *clipboard.Manager, parent *systray.MenuItem) {
	// Hide previously created submenu items, if any, to avoid having duplicates.
	for _, item := range historyItems {
		item.Hide()
	}
	// Reset the slice so we start tracking the fresh set of items.
	historyItems = historyItems[:0]

	history := manager.GetHistory()
	if len(history) == 0 {
		empty := parent.AddSubMenuItem("History is empty", "")
		empty.Disable()
		// Track the placeholder item so it can be cleared later.
		historyItems = append(historyItems, empty)
		return
	}

	for i, hItem := range history {
		menuItem := parent.AddSubMenuItem(hItem.Preview, hItem.Timestamp.Format("2006-01-02 15:04:05"))
		historyItems = append(historyItems, menuItem)

		go func(content string) {
			for range menuItem.ClickedCh {
				manager.CopyToClipboard(content)
				beeep.Notify("Smart Clipboard", "Text copied to clipboard", "")
			}
		}(hItem.Content)

		// Limit the number of elements displayed in the submenu.
		if i >= 9 {
			break
		}
	}
}

func showHistoryWindow(manager *clipboard.Manager) {
	// Здесь будет код окна с историей
	// Можно использовать fyne, walk или другие GUI библиотеки
	beeep.Alert("Smart Clipboard", "History window will be implemented in the next version", "")
}

func showSettingsWindow(cfg *config.Config, store *storage.Storage) {
	// Окно настроек
	beeep.Alert("Smart Clipboard", "Settings window will be implemented in the next version", "")
}

func getIcon() []byte {
	if len(trayIcon) > 0 {
		return trayIcon
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

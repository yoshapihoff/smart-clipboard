//go:build cgo
// +build cgo

package tray

import (
	"encoding/base64"
	"log"

	"fyne.io/systray"
	"github.com/gen2brain/beeep"
	"github.com/yoshapihoff/smart-clipboard/internal/clipboard"
	"github.com/yoshapihoff/smart-clipboard/internal/config"
	"github.com/yoshapihoff/smart-clipboard/internal/storage"
)

var trayIcon []byte

var historyItems []*systray.MenuItem

func RunTray(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	systray.Run(onReady(manager, store, cfg), onExit(store))
}

func RunTrayWithHotkeys(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	systray.Run(onReady(manager, store, cfg), onExit(store))
}

func onReady(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) func() {
	return func() {
		systray.SetIcon(getIcon())

		// Формируем подсказку с учетом платформы
		tooltip := "Smart clipboard\nLeft click: History\nRight click: Menu"
		systray.SetTooltip(tooltip)

		systray.AddSeparator()

		settingsMenu := systray.AddMenuItem("Settings", "Open settings")
		clearMenu := systray.AddMenuItem("Clear history", "Clear all history")
		systray.AddSeparator()
		quitMenu := systray.AddMenuItem("Quit", "Quit program")

		systray.AddSeparator()

		go func() {
			for range systray.TrayOpenedCh {
				rebuildMenu(manager, store, cfg)
			}
		}()

		go func() {
			for {
				select {
				case <-settingsMenu.ClickedCh:
					showSettingsWindow(cfg, store)
				case <-clearMenu.ClickedCh:
					manager.ClearHistory()
					rebuildMenu(manager, store, cfg)
					beeep.Notify("Smart clipboard", "History cleared", "")
				case <-quitMenu.ClickedCh:
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
		log.Println("Выход из приложения")
	}
}

func rebuildMenu(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	for _, item := range historyItems {
		item.Remove()
	}
	historyItems = historyItems[:0]

	history := manager.GetHistory()
	if len(history) == 0 {
		empty := systray.AddMenuItem("History is empty", "")
		empty.Disable()
		historyItems = append(historyItems, empty)
	} else {
		maxItems := cfg.MaxDisplayItems
		if maxItems <= 0 {
			maxItems = 10 // fallback to default if config is invalid
		}

		for i, hItem := range history {
			menuItem := systray.AddMenuItem(hItem.Preview, hItem.Timestamp.Format("2006-01-02 15:04:05"))
			historyItems = append(historyItems, menuItem)

			go func(content string) {
				for range menuItem.ClickedCh {
					manager.CopyToClipboard(content)
					manager.IncrementClickCount(content)
					beeep.Notify("Smart Clipboard", "Text copied to clipboard", "")
				}
			}(hItem.Content)

			if i >= maxItems-1 {
				break
			}
		}
	}
}

func showSettingsWindow(cfg *config.Config, store *storage.Storage) {
	beeep.Alert("Smart Clipboard", "Settings window will be implemented in the next version", "")
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

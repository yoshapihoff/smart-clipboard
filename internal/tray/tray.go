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

var historyMenuItems map[*systray.MenuItem]clipboard.ClipboardItem

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
		tooltip := "Smart clipboard"
		systray.SetTooltip(tooltip)

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
					manager.ClearHistory()
					rebuildHistoryMenu(manager, store, cfg)
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

func rebuildHistoryMenu(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	history := manager.GetHistory()

	if len(history) == 0 {
		for menuItem := range historyMenuItems {
			menuItem.Remove()
		}
		historyMenuItems = make(map[*systray.MenuItem]clipboard.ClipboardItem)

		empty := systray.AddMenuItem("History is empty", "")
		empty.Disable()
		historyMenuItems[empty] = clipboard.ClipboardItem{}
		return
	}

	var maxReusedMenuItemIndex int
	for menuItem, clipboardItem := range historyMenuItems {
		var title string
		if cfg.DebugMode {
			title = fmt.Sprintf("[%d] %s", clipboardItem.ClickCount, clipboardItem.Preview)
		} else {
			title = clipboardItem.Preview
		}

		menuItem.SetTitle(title)
		menuItem.SetTooltip(clipboardItem.Timestamp.Format("2006-01-02 15:04:05"))
		if menuItem.Disabled() {
			menuItem.Enable()
		}

		maxReusedMenuItemIndex += 1
	}

	for i := maxReusedMenuItemIndex; i < len(history); i++ {
		var title string
		if cfg.DebugMode {
			title = fmt.Sprintf("[%d] %s", history[i].ClickCount, history[i].Preview)
		} else {
			title = history[i].Preview
		}

		menuItem := systray.AddMenuItem(title, history[i].Timestamp.Format("2006-01-02 15:04:05"))
		menuItem.SetTitle(title)
		menuItem.SetTooltip(history[i].Timestamp.Format("2006-01-02 15:04:05"))

		historyMenuItems[menuItem] = history[i]
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

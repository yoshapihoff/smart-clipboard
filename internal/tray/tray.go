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
	"github.com/yoshapihoff/smart-clipboard/internal/types"
)

var trayIcon []byte
var menuItemPool *GenericSlice[*systray.MenuItem]
var menuCancelChannels []chan struct{}

func initMenuItemPool(size int) {
	stopMenuHandlers()
	menuItemPool = NewGenericSliceWithCapacity[*systray.MenuItem](size)
	menuCancelChannels = make([]chan struct{}, size)
	for i := 0; i < size; i++ {
		menuItem := systray.AddMenuItem("", "")
		menuItem.Hide()
		menuItemPool.Add(menuItem)
		menuCancelChannels[i] = make(chan struct{})
	}
}

func RunTray(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	systray.Run(onReady(manager, store, cfg), onExit(store))
}

func onReady(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) func() {
	return func() {
		systray.SetIcon(getIcon())
		tooltip := "Smart clipboard"
		systray.SetTooltip(tooltip)

		settingsMenu := systray.AddMenuItem("Settings", "Open settings")
		clearMenu := systray.AddMenuItem("Clear history", "Clear all history")
		systray.AddSeparator()
		quitMenu := systray.AddMenuItem("Quit", "Quit program")

		maxItemsMenu := settingsMenu.AddSubMenuItem(fmt.Sprintf("Max items: %d", cfg.MaxItems), "Max items")
		maxItemsMenu.Disable()
		incMaxItemsMenu := settingsMenu.AddSubMenuItem("+5 items", "Increase max items")
		decMaxItemsMenu := settingsMenu.AddSubMenuItem("-5 items", "Decrease max items")
		settingsMenu.AddSeparator()
		debugModeMenu := settingsMenu.AddSubMenuItem(fmt.Sprintf("Debug mode: %t", cfg.DebugMode), "Debug mode")

		// Sync is always enabled

		initMenuItemPool(cfg.MaxItems)

		go func() {
			for range systray.TrayOpenedCh {
				rebuildHistoryMenu(manager, store, cfg)
			}
		}()

		go func() {
			for {
				select {
				case <-incMaxItemsMenu.ClickedCh:
					cfg.MaxItems += 5
					initMenuItemPool(cfg.MaxItems)
					maxItemsMenu.SetTitle(fmt.Sprintf("Max items: %d", cfg.MaxItems))
					config.SaveConfig(cfg)
				case <-decMaxItemsMenu.ClickedCh:
					if cfg.MaxItems > 5 {
						cfg.MaxItems -= 5
						initMenuItemPool(cfg.MaxItems)
						maxItemsMenu.SetTitle(fmt.Sprintf("Max items: %d", cfg.MaxItems))
					}
				case <-debugModeMenu.ClickedCh:
					cfg.DebugMode = !cfg.DebugMode
					debugModeMenu.SetTitle(fmt.Sprintf("Debug mode: %t", cfg.DebugMode))
					config.SaveConfig(cfg)
				case <-clearMenu.ClickedCh:
					manager.ClearClipboard()
					manager.ClearHistory()
					store.SaveHistory(manager.GetHistory())
					initMenuItemPool(cfg.MaxItems)
					beeep.Notify("Smart clipboard", "History cleared", "")
				case <-quitMenu.ClickedCh:
					stopMenuHandlers()
					store.SaveHistory(manager.GetHistory())
					systray.Quit()
					return
				}
			}
		}()
	}
}

func onExit(store *storage.Storage) func() {
	return func() {
		stopMenuHandlers()
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
	systray.AddSeparator()

	history := manager.GetHistory()
	stopMenuHandlers()

	for i := range menuCancelChannels {
		menuCancelChannels[i] = make(chan struct{})
	}

	if len(history) == 0 {
		if menuItemPool.Length() > 0 {
			if menuItem, ok := menuItemPool.Get(0); ok {
				menuItem.SetTitle("History is empty")
				menuItem.SetTooltip("No clipboard history available")
				menuItem.Disable()
				menuItem.Show()
			}
		}
		return
	}

	itemsToShow := len(history)
	for i := 0; i < itemsToShow; i++ {
		item := history[i]
		menuItem, _ := menuItemPool.Get(i)
		cancelChan := menuCancelChannels[i]

		var title string
		if cfg.DebugMode {
			title = fmt.Sprintf("[%d] %s", item.ClickCount, item.Preview)
		} else {
			title = item.Preview
		}

		menuItem.SetTitle(title)
		menuItem.SetTooltip(item.Timestamp.Format("2006-01-02 15:04:05"))
		if menuItem.Disabled() {
			menuItem.Enable()
		}
		menuItem.Show()

		go func(menuItem *systray.MenuItem, clipboardItem types.ClipboardItem, cancelChan chan struct{}) {
			for {
				select {
				case <-menuItem.ClickedCh:
					manager.CopyToClipboard(clipboardItem.Content)
					manager.IncrementClickCount(clipboardItem.Content)
					return
				case <-cancelChan:
					return
				}
			}
		}(menuItem, item, cancelChan)
	}
}

func getIcon() []byte {
	if len(trayIcon) > 0 {
		return trayIcon
	}

	trayIcon, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil {
		log.Printf("tray: failed to decode base64 icon: %v", err)
		return nil
	}
	return trayIcon
}

//go:build !cgo
// +build !cgo

// The stub implementation of the tray package for environments where CGO is
// disabled or a system tray is not available (for example, many minimal Linux
// containers). It provides a no-op RunTray so the rest of the application can
// still compile and run in headless mode.
package tray

import (
    "log"

    "github.com/yoshapihoff/smart-clipboard/internal/clipboard"
    "github.com/yoshapihoff/smart-clipboard/internal/config"
    "github.com/yoshapihoff/smart-clipboard/internal/storage"
)

// RunTray is a noop when CGO is disabled. We simply log a message so the user
// knows the system-tray UI is not available in the current build.
func RunTray(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
    log.Println("[smart-clipboard] CGO disabled – system tray UI is not available. Running in headless mode.")
    // Persist any existing history to avoid data loss.
    if err := store.SaveHistory(manager.GetHistory()); err != nil {
        log.Printf("[smart-clipboard] failed to persist history: %v", err)
    }
}

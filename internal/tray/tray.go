package tray

import (
	"clipboard-history/pkg/clipboard"
	"clipboard-history/pkg/config"
	"clipboard-history/pkg/storage"
	"log"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

func RunTray(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) {
	systray.Run(onReady(manager, store, cfg), onExit(store))
}

func onReady(manager *clipboard.Manager, store *storage.Storage, cfg *config.Config) func() {
	return func() {
		// Иконка в трее
		systray.SetIcon(getIcon())
		systray.SetTitle("Clipboard History")
		systray.SetTooltip("Clipboard History Manager")

		// Меню истории
		historyMenu := systray.AddMenuItem("История", "Показать историю")
		updateHistoryMenu(manager, historyMenu)

		// Разделитель
		systray.AddSeparator()

		// Настройки
		settingsMenu := systray.AddMenuItem("Настройки", "Открыть настройки")

		// Очистка истории
		clearMenu := systray.AddMenuItem("Очистить историю", "Удалить всю историю")

		// Разделитель
		systray.AddSeparator()

		// Выход
		quitMenu := systray.AddMenuItem("Выход", "Завершить программу")

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
					beeep.Notify("Clipboard History", "История очищена", "assets/icon.png")
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
	// Очищаем старое меню
	for _, item := range parent.Items() {
		item.Hide()
	}

	history := manager.GetHistory()
	if len(history) == 0 {
		empty := parent.AddSubMenuItem("История пуста", "")
		empty.Disable()
		return
	}

	for i, item := range history {
		menuItem := parent.AddSubMenuItem(item.Preview, item.Timestamp.Format("2006-01-02 15:04:05"))
		go func(content string) {
			for range menuItem.ClickedCh {
				manager.CopyToClipboard(content)
				beeep.Notify("Clipboard History", "Текст скопирован в буфер", "assets/icon.png")
			}
		}(item.Content)

		// Ограничиваем количество элементов в меню
		if i >= 9 {
			break
		}
	}
}

func showHistoryWindow(manager *clipboard.Manager) {
	// Здесь будет код окна с историей
	// Можно использовать fyne, walk или другие GUI библиотеки
	beeep.Alert("Clipboard History", "Окно истории будет реализовано в следующей версии", "assets/icon.png")
}

func showSettingsWindow(cfg *config.Config, store *storage.Storage) {
	// Окно настроек
	beeep.Alert("Настройки", "Окно настроек будет реализовано в следующей версии", "assets/icon.png")
}

func getIcon() []byte {
	// Загрузка иконки из assets
	// В реальном приложении нужно загружать из файла
	return []byte{} // Заглушка
}

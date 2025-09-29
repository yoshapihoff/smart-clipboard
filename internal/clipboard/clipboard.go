package clipboard

import (
	"time"

	"github.com/yoshapihoff/smart-clipboard/internal/sync"
	"github.com/yoshapihoff/smart-clipboard/internal/types"
)

type Manager struct {
	history        []types.ClipboardItem
	maxHistorySize int
	syncManager    *sync.SyncManager
	lastContent    string // Отслеживаем последнее содержимое буфера обмена
}

func NewManager(initialHistory []types.ClipboardItem, maxSize int, syncManager *sync.SyncManager) *Manager {
	return &Manager{
		history:        initialHistory,
		maxHistorySize: maxSize,
		syncManager:    syncManager,
	}
}

// AddToHistory добавляет содержимое в историю и отправляет по сети только если содержимое изменилось
func (m *Manager) AddToHistory(content string) {
	if content == "" {
		return
	}

	// Проверяем, изменилось ли содержимое буфера обмена
	if content == m.lastContent {
		return // Содержимое не изменилось, ничего не делаем
	}

	// Обновляем последнее содержимое
	m.lastContent = content

	var existingClickCount int
	found := false

	// Проверяем, есть ли уже такой элемент в истории
	for _, item := range m.history {
		if item.Content == content {
			// Если элемент найден, сохраняем счётчик кликов и удаляем старый
			existingClickCount = item.ClickCount
			found = true
			m.removeFromHistory(content)
			break
		}
	}

	item := types.ClipboardItem{
		Content:    content,
		Timestamp:  time.Now(),
		Preview:    getPreview(content),
		ClickCount: existingClickCount,
	}

	// Если элемент не был найден, добавляем его в историю
	if !found {
		m.history = append([]types.ClipboardItem{item}, m.history...)
	} else {
		// Если элемент был найден, вставляем его в начало с сохранённым счётчиком
		m.history = append([]types.ClipboardItem{item}, m.history...)
	}

	// Сортируем историю: сначала по количеству кликов (по убыванию), затем по времени (по убыванию)
	m.sortHistory()

	// Ограничение размера истории
	if len(m.history) > m.maxHistorySize {
		m.history = m.history[:m.maxHistorySize]
	}

	// Отправляем историю по сети
	if m.syncManager != nil {
		go m.syncManager.SendHistory(m.history)
	}
}

// SetLastContent устанавливает последнее известное содержимое буфера обмена
func (m *Manager) SetLastContent(content string) {
	m.lastContent = content
}

// GetLastContent возвращает последнее известное содержимое буфера обмена
func (m *Manager) GetLastContent() string {
	return m.lastContent
}

// removeFromHistory удаляет элемент из истории по содержимому
func (m *Manager) removeFromHistory(content string) {
	for i, item := range m.history {
		if item.Content == content {
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}
}

// sortHistory сортирует историю: сначала по количеству кликов (по убыванию), затем по времени (по убыванию)
func (m *Manager) sortHistory() {
	// Используем встроенную сортировку с кастомной функцией сравнения
	for i := 0; i < len(m.history)-1; i++ {
		for j := i + 1; j < len(m.history); j++ {
			// Сравниваем элементы
			if m.shouldSwap(m.history[i], m.history[j]) {
				m.history[i], m.history[j] = m.history[j], m.history[i]
			}
		}
	}
}

// shouldSwap определяет, нужно ли поменять местами два элемента
func (m *Manager) shouldSwap(a, b types.ClipboardItem) bool {
	// Если количество кликов разное, элемент с большим количеством кликов должен быть выше
	if a.ClickCount != b.ClickCount {
		return a.ClickCount < b.ClickCount
	}

	// Если количество кликов одинаковое, более новый элемент должен быть выше
	return a.Timestamp.Before(b.Timestamp)
}

func (m *Manager) GetHistory() []types.ClipboardItem {
	return m.history
}

func (m *Manager) ClearHistory() {
	m.history = []types.ClipboardItem{}
}

func (m *Manager) CopyToClipboard(content string) error {
	return SetClipboard(content)
}

// ClearClipboard очищает содержимое системного буфера обмена
func (m *Manager) ClearClipboard() error {
	return SetClipboard("")
}

// IncrementClickCount увеличивает счётчик кликов для элемента и пересортировывает историю
func (m *Manager) IncrementClickCount(content string) {
	for i, item := range m.history {
		if item.Content == content {
			m.history[i].ClickCount++
			// Пересортировываем историю после изменения счётчика
			m.sortHistory()
			break
		}
	}
}

func (m *Manager) ReplaceHistory(history []types.ClipboardItem) {
	m.history = history
	m.sortHistory()

	// Ограничение размера истории
	if len(m.history) > m.maxHistorySize {
		m.history = m.history[:m.maxHistorySize]
	}
}

func getPreview(content string) string {
	if len(content) <= 100 {
		return content
	}
	return content[:100] + "..."
}

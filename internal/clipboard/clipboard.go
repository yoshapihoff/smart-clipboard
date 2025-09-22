package clipboard

import (
	"time"
)

type ClipboardItem struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Preview   string    `json:"preview"`
}

type Manager struct {
	history        []ClipboardItem
	maxHistorySize int
	lastContent    string
}

func NewManager(initialHistory []ClipboardItem, maxSize int) *Manager {
	return &Manager{
		history:        initialHistory,
		maxHistorySize: maxSize,
	}
}

func (m *Manager) AddToHistory(content string) {
	if content == "" || content == m.lastContent {
		return
	}

	item := ClipboardItem{
		Content:   content,
		Timestamp: time.Now(),
		Preview:   getPreview(content),
	}

	m.history = append([]ClipboardItem{item}, m.history...)
	m.lastContent = content

	// Ограничение размера истории
	if len(m.history) > m.maxHistorySize {
		m.history = m.history[:m.maxHistorySize]
	}
}

func (m *Manager) GetHistory() []ClipboardItem {
	return m.history
}

func (m *Manager) ClearHistory() {
	m.history = []ClipboardItem{}
	m.lastContent = ""
}

func (m *Manager) CopyToClipboard(content string) error {
	return SetClipboard(content)
}

func getPreview(content string) string {
	if len(content) <= 100 {
		return content
	}
	return content[:100] + "..."
}

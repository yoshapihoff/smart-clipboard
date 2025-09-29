package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/yoshapihoff/smart-clipboard/internal/types"
)

type Storage struct {
	filePath string
}

func NewStorage(filePath string) (*Storage, error) {
	// Создаем директорию если не существует
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Storage{filePath: filePath}, nil
}

func (s *Storage) SaveHistory(history []types.ClipboardItem) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) LoadHistory() ([]types.ClipboardItem, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []types.ClipboardItem{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var history []types.ClipboardItem
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (s *Storage) CleanOldEntries(maxAge time.Duration) error {
	history, err := s.LoadHistory()
	if err != nil {
		return err
	}

	var filtered []types.ClipboardItem
	cutoff := time.Now().Add(-maxAge)

	for _, item := range history {
		if item.Timestamp.After(cutoff) {
			filtered = append(filtered, item)
		}
	}

	return s.SaveHistory(filtered)
}

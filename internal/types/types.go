package types

import (
	"time"
)

type ClipboardItem struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Preview   string    `json:"preview"`
	ClickCount int      `json:"click_count"`
}

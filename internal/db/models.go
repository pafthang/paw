package db

import "time"

type ChatSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"index" json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []ChatMessage
}

type ChatMessage struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ChatSessionID uint      `gorm:"index" json:"chat_session_id"`
	Role          string    `gorm:"index" json:"role"`
	Content       string    `json:"content"`
	Model         string    `json:"model,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type MemoryItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"index" json:"type"`
	Content   string    `json:"content"`
	Metadata  string    `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

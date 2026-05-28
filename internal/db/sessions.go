package db

import (
	"strings"

	"gorm.io/gorm"
)

const defaultSessionLimit = 50

func CreateChatSession(database *gorm.DB, title string) (*ChatSession, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "New chat"
	}
	if len(title) > 80 {
		title = title[:80]
	}
	session := &ChatSession{Title: title}
	if err := database.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func GetChatSession(database *gorm.DB, id uint) (*ChatSession, error) {
	var session ChatSession
	if err := database.Preload("Messages", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("created_at asc, id asc")
	}).First(&session, id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func ListChatSessions(database *gorm.DB, limit int) ([]ChatSession, error) {
	if limit <= 0 {
		limit = defaultSessionLimit
	}
	var sessions []ChatSession
	if err := database.Order("updated_at desc, id desc").Limit(limit).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func AddChatMessage(database *gorm.DB, sessionID uint, role, content, model string) (*ChatMessage, error) {
	message := &ChatMessage{
		ChatSessionID: sessionID,
		Role:          role,
		Content:       content,
		Model:         model,
	}
	if err := database.Create(message).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&ChatSession{}).Where("id = ?", sessionID).Update("updated_at", message.CreatedAt).Error; err != nil {
		return nil, err
	}
	return message, nil
}

func DeleteChatSession(database *gorm.DB, id uint) error {
	return database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("chat_session_id = ?", id).Delete(&ChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Delete(&ChatSession{}, id).Error
	})
}

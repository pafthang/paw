package db

import (
	"strings"

	"gorm.io/gorm"
)

const defaultSessionLimit = 50

const DefaultHistoryLimit = 20
const MaxHistoryLimit = 100

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

func ListRecentChatMessages(database *gorm.DB, sessionID uint, limit int) ([]ChatMessage, error) {
	if limit < 0 {
		limit = 0
	}
	if limit == 0 {
		return []ChatMessage{}, nil
	}
	if limit > MaxHistoryLimit {
		limit = MaxHistoryLimit
	}

	var newestFirst []ChatMessage
	if err := database.Where("chat_session_id = ?", sessionID).
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&newestFirst).Error; err != nil {
		return nil, err
	}

	messages := make([]ChatMessage, len(newestFirst))
	for i := range newestFirst {
		messages[len(newestFirst)-1-i] = newestFirst[i]
	}
	return messages, nil
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

func RenameChatSession(database *gorm.DB, id uint, title string) (*ChatSession, error) {
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, gorm.ErrInvalidValue
	}
	if len(title) > 80 {
		title = title[:80]
	}
	if err := database.Model(&ChatSession{}).Where("id = ?", id).Update("title", title).Error; err != nil {
		return nil, err
	}
	return GetChatSession(database, id)
}

func SearchChatSessions(database *gorm.DB, query string, limit int) ([]ChatSession, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, gorm.ErrInvalidValue
	}
	if limit <= 0 {
		limit = defaultSessionLimit
	}
	like := "%" + query + "%"
	var sessions []ChatSession
	if err := database.
		Model(&ChatSession{}).
		Distinct("chat_sessions.*").
		Joins("LEFT JOIN chat_messages ON chat_messages.chat_session_id = chat_sessions.id").
		Where("chat_sessions.title LIKE ? OR chat_messages.content LIKE ?", like, like).
		Order("chat_sessions.updated_at desc, chat_sessions.id desc").
		Limit(limit).
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

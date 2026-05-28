package db

import (
	"github.com/pafthang/paw/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open() (*gorm.DB, error) {
	if err := config.EnsureDir(); err != nil {
		return nil, err
	}
	path, err := config.DBPath()
	if err != nil {
		return nil, err
	}
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(&ChatSession{}, &ChatMessage{}, &MemoryItem{}, &FileRecord{}, &AuditEvent{}); err != nil {
		return nil, err
	}
	return database, nil
}

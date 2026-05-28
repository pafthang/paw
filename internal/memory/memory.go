package memory

import (
	"errors"
	"strings"

	"github.com/pafthang/paw/internal/db"
	"gorm.io/gorm"
)

func Add(database *gorm.DB, itemType, content string, metadata string) (*db.MemoryItem, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	itemType = strings.TrimSpace(itemType)
	content = strings.TrimSpace(content)
	if itemType == "" {
		return nil, errors.New("type is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}
	item := db.MemoryItem{Type: itemType, Content: content, Metadata: strings.TrimSpace(metadata)}
	if err := database.Create(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func Get(database *gorm.DB, id uint) (*db.MemoryItem, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	if id == 0 {
		return nil, errors.New("id is required")
	}
	var item db.MemoryItem
	if err := database.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func Delete(database *gorm.DB, id uint) error {
	if database == nil {
		return errors.New("database is required")
	}
	if id == 0 {
		return errors.New("id is required")
	}
	return database.Delete(&db.MemoryItem{}, id).Error
}

func List(database *gorm.DB, limit int) ([]db.MemoryItem, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	if limit <= 0 {
		limit = 50
	}
	var items []db.MemoryItem
	if err := database.Order("created_at desc, id desc").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func Search(database *gorm.DB, query string, limit int) ([]db.MemoryItem, error) {
	if database == nil {
		return nil, errors.New("database is required")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = 50
	}
	like := "%" + query + "%"
	var items []db.MemoryItem
	if err := database.
		Where("content LIKE ? OR metadata LIKE ? OR type LIKE ?", like, like, like).
		Order("created_at desc, id desc").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&ChatSession{}, &ChatMessage{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestRenameAndSearchSessions(t *testing.T) {
	database := openTestDB(t)
	s1, err := CreateChatSession(database, "hello world")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := AddChatMessage(database, s1.ID, "user", "some message", "m"); err != nil {
		t.Fatalf("add msg: %v", err)
	}
	s2, err := RenameChatSession(database, s1.ID, "new title")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if s2.Title != "new title" {
		t.Fatalf("title=%q", s2.Title)
	}
	found, err := SearchChatSessions(database, "message", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(found) != 1 || found[0].ID != s1.ID {
		t.Fatalf("found=%#v", found)
	}
}

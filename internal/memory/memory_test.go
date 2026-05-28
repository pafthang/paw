package memory

import (
	"testing"

	"github.com/pafthang/paw/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&db.MemoryItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestMemoryCRUDAndSearch(t *testing.T) {
	database := openTestDB(t)

	item, err := Add(database, "fact", "The project is called Paw.", `{"source":"test"}`)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	got, err := Get(database, item.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Content != item.Content || got.Type != item.Type {
		t.Fatalf("got=%#v item=%#v", got, item)
	}

	results, err := Search(database, "Paw", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].ID != item.ID {
		t.Fatalf("results=%#v", results)
	}

	if err := Delete(database, item.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	results, err = List(database, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty list, got=%#v", results)
	}
}

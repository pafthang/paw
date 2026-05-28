package filestore

import (
	"os"
	"path/filepath"
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
	if err := gdb.AutoMigrate(&db.FileRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestAddListSearchDelete(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "hello.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	database := openTestDB(t)
	rec, err := AddFromPath(database, src, `{"source":"test"}`)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if rec.ID == 0 || rec.Sha256 == "" || rec.Path == "" {
		t.Fatalf("rec=%#v", rec)
	}

	items, err := List(database, 10)
	if err != nil || len(items) != 1 {
		t.Fatalf("list err=%v items=%#v", err, items)
	}

	found, err := Search(database, "hello", 10)
	if err != nil || len(found) != 1 {
		t.Fatalf("search err=%v found=%#v", err, found)
	}

	if err := Delete(database, rec.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	items, _ = List(database, 10)
	if len(items) != 0 {
		t.Fatalf("expected empty, got=%#v", items)
	}
}

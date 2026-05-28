package search

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/filestore"
	"github.com/pafthang/paw/internal/memory"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&db.ChatSession{}, &db.ChatMessage{}, &db.MemoryItem{}, &db.FileRecord{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestRun_ReturnsTypedResults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	database := openTestDB(t)
	_, _ = memory.Add(database, "fact", "Paw project", "")
	s, _ := db.CreateChatSession(database, "Paw session")
	_, _ = db.AddChatMessage(database, s.ID, "user", "Paw message", "m")

	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "paw.txt")
	_ = os.WriteFile(src, []byte("paw"), 0o600)
	_, _ = filestore.AddFromPath(database, src, "")

	resp, err := Run(database, "Paw", 50)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.Query != "Paw" {
		t.Fatalf("query=%q", resp.Query)
	}
	if len(resp.Results) == 0 {
		t.Fatalf("expected results")
	}
	for _, r := range resp.Results {
		if r.Type == "" || r.ID == 0 {
			t.Fatalf("bad result: %#v", r)
		}
	}
}

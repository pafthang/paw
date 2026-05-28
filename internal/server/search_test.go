package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/filestore"
	"github.com/pafthang/paw/internal/memory"
)

func TestSearchAPI_ReturnsResults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, _ = memory.Add(database, "fact", "Paw project", "")
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "paw.txt")
	_ = os.WriteFile(src, []byte("paw"), 0o600)
	_, _ = filestore.AddFromPath(database, src, "")

	s := New(config.DefaultSettings())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=Paw", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

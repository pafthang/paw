package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/db"
)

func writeToken(t *testing.T, home string, token string) {
	t.Helper()
	dir := filepath.Join(home, ".pocketpaw")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "access_token"), []byte(token+"\n"), 0o600); err != nil {
		t.Fatalf("write token: %v", err)
	}
}

func TestSessionsSearchAndPatch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeToken(t, home, "token123")

	s := New(config.DefaultSettings())

	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	session, err := db.CreateChatSession(database, "hello")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := db.AddChatMessage(database, session.ID, "user", "hello message", "test"); err != nil {
		t.Fatalf("add msg: %v", err)
	}

	// patch title
	patchBody, _ := json.Marshal(map[string]any{"title": "renamed"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/sessions/"+strconv.FormatUint(uint64(session.ID), 10), bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch code=%d body=%s", rec.Code, rec.Body.String())
	}

	// search
	req = httptest.NewRequest(http.MethodGet, "/api/v1/sessions/search?q=renamed", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search code=%d body=%s", rec.Code, rec.Body.String())
	}
}

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
)

func writeAccessToken(t *testing.T, home string, token string) {
	t.Helper()
	dir := filepath.Join(home, ".pocketpaw")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "access_token"), []byte(token+"\n"), 0o600); err != nil {
		t.Fatalf("write token: %v", err)
	}
}

func TestMemoryAPI_RequiresAuth(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	s := New(defaultTestSettings())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/memory", nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMemoryAPI_CRUD(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	s := New(defaultTestSettings())

	// create
	body, _ := json.Marshal(map[string]any{"type": "fact", "content": "Paw", "metadata": map[string]any{"source": "test"}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create code=%d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ID == 0 {
		t.Fatalf("bad create response: %v body=%s", err, rec.Body.String())
	}

	// list
	req = httptest.NewRequest(http.MethodGet, "/api/v1/memory", nil)
	req.Header.Set("X-Paw-Access-Token", "token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list code=%d body=%s", rec.Code, rec.Body.String())
	}

	// search
	req = httptest.NewRequest(http.MethodGet, "/api/v1/memory/search?q=Paw", nil)
	req.Header.Set("X-Paw-Access-Token", "token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search code=%d body=%s", rec.Code, rec.Body.String())
	}

	// get
	req = httptest.NewRequest(http.MethodGet, "/api/v1/memory/"+itoa(created.ID), nil)
	req.Header.Set("X-Paw-Access-Token", "token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get code=%d body=%s", rec.Code, rec.Body.String())
	}

	// delete
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/memory/"+itoa(created.ID), nil)
	req.Header.Set("X-Paw-Access-Token", "token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete code=%d body=%s", rec.Code, rec.Body.String())
	}
}

func defaultTestSettings() config.Settings {
	return config.DefaultSettings()
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

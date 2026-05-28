package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pafthang/paw/internal/config"
)

func TestFilesAPI_CRUD(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "hello.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	s := New(config.DefaultSettings())

	// create
	body, _ := json.Marshal(map[string]any{"path": src, "metadata": map[string]any{"source": "test"}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files", bytes.NewReader(body))
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
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list code=%d body=%s", rec.Code, rec.Body.String())
	}

	// search
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/search?q=hello", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search code=%d body=%s", rec.Code, rec.Body.String())
	}

	// get
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/"+itoa(created.ID), nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get code=%d body=%s", rec.Code, rec.Body.String())
	}

	// delete
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/files/"+itoa(created.ID), nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete code=%d body=%s", rec.Code, rec.Body.String())
	}
}

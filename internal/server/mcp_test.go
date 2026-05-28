package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pafthang/paw/internal/config"
)

func TestMCPAPI_AddListRemove(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	s := New(config.DefaultSettings())

	body, _ := json.Marshal(map[string]any{"name": "demo", "command": "echo", "args": []string{"hi"}, "env": map[string]string{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("add code=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/mcp", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list code=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/mcp/demo", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete code=%d body=%s", rec.Code, rec.Body.String())
	}
}

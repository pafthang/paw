package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pafthang/paw/internal/config"
)

func TestChannelsAPI_ListStatus(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	s := New(config.DefaultSettings())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list code=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/channels/status", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status code=%d body=%s", rec.Code, rec.Body.String())
	}
}

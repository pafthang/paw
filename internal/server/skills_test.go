package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pafthang/paw/internal/config"
)

func TestSkillsAPI_ListAndShow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeAccessToken(t, home, "token123")

	skillsDir, _ := config.SkillsDir()
	skillPath := filepath.Join(skillsDir, "demo", "skill.yaml")
	_ = os.MkdirAll(filepath.Dir(skillPath), 0o700)
	_ = os.WriteFile(skillPath, []byte(`
name: demo
description: demo skill
version: 0.0.1
prompts:
  system: hi
`), 0o600)

	s := New(config.DefaultSettings())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list code=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/skills/demo", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("show code=%d body=%s", rec.Code, rec.Body.String())
	}
}

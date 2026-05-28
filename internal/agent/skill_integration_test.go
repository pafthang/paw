package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/tools"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type captureLLM struct {
	lastReq llm.ChatRequest
}

func (c *captureLLM) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	c.lastReq = req
	return llm.ChatResponse{Model: "fake", Content: "ok"}, nil
}

func openSkillTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&db.ChatSession{}, &db.ChatMessage{}, &db.MemoryItem{}, &db.FileRecord{}, &db.AuditEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestAgentSkill_InsertsSystemPrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	skillsDir, err := config.SkillsDir()
	if err != nil {
		t.Fatalf("skills dir: %v", err)
	}
	skillPath := filepath.Join(skillsDir, "demo", "skill.yaml")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(skillPath, []byte(`
name: demo
description: demo skill
version: 0.0.1
prompts:
  system: |
    DEMO SKILL SYSTEM
`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	client := &captureLLM{}
	runner := NewRunner(openSkillTestDB(t), tools.DefaultRegistry())
	_, err = runner.Chat(context.Background(), client, ChatRequest{
		Prompt:        "hello",
		Model:         "fake",
		Skill:         "demo",
		MaxIterations: 1,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(client.lastReq.Messages) == 0 {
		t.Fatalf("no messages")
	}
	if client.lastReq.Messages[0].Role != "system" {
		t.Fatalf("first role=%q", client.lastReq.Messages[0].Role)
	}
	if !strings.Contains(client.lastReq.Messages[0].Content, "DEMO SKILL SYSTEM") {
		t.Fatalf("system=%q", client.lastReq.Messages[0].Content)
	}
}

package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/tools"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeLLM struct {
	responses []llm.ChatResponse
	calls     int
}

func (f *fakeLLM) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	if f.calls >= len(f.responses) {
		return llm.ChatResponse{Model: "fake", Content: "done"}, nil
	}
	resp := f.responses[f.calls]
	f.calls++
	return resp, nil
}

type fakeTool struct {
	name string
	run  func(input json.RawMessage) (tools.ToolResult, error)
}

func (t fakeTool) Name() string        { return t.name }
func (t fakeTool) Description() string { return "fake" }
func (t fakeTool) Run(ctx context.Context, input json.RawMessage) (tools.ToolResult, error) {
	return t.run(input)
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&db.ChatSession{}, &db.ChatMessage{}, &db.MemoryItem{}, &db.AuditEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestRunnerChat_MultiStepRunsToolsTwice(t *testing.T) {
	workspace := t.TempDir()

	client := &fakeLLM{responses: []llm.ChatResponse{
		{Model: "fake", Content: `{"tool_calls":[{"name":"file.read","input":{"path":"README.md"}}]}`},
		{Model: "fake", Content: `{"tool_calls":[{"name":"file.read","input":{"path":"README.md"}}]}`},
		{Model: "fake", Content: "final answer"},
	}}

	reg := tools.NewRegistry(fakeTool{
		name: "file.read",
		run: func(input json.RawMessage) (tools.ToolResult, error) {
			return tools.ToolResult{Content: "ok"}, nil
		},
	})
	runner := NewRunner(openTestDB(t), reg)

	resp, err := runner.Chat(context.Background(), client, ChatRequest{
		Prompt:          "do it",
		Model:           "fake",
		MaxIterations:   4,
		Workspace:       workspace,
		MaxContextChars: 10000,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !resp.UsedTools {
		t.Fatalf("expected used_tools=true")
	}
	if resp.Iterations != 3 {
		t.Fatalf("iterations = %d", resp.Iterations)
	}
	if resp.FinalResponse.Content != "final answer" {
		t.Fatalf("final = %q", resp.FinalResponse.Content)
	}
	if len(resp.Steps) != 3 || len(resp.Steps[0].ToolCalls) != 1 || len(resp.Steps[1].ToolCalls) != 1 {
		t.Fatalf("steps = %#v", resp.Steps)
	}
}

package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type FileWriteTool struct{}

func (FileWriteTool) Name() string { return "file.write" }
func (FileWriteTool) Description() string { return "Write UTF-8 text content to disk, creating parent directories when needed." }

func (FileWriteTool) Run(ctx context.Context, input json.RawMessage) (ToolResult, error) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return ToolResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(req.Path), 0o755); err != nil {
		return ToolResult{}, err
	}
	if err := os.WriteFile(req.Path, []byte(req.Content), 0o644); err != nil {
		return ToolResult{}, err
	}
	return ToolResult{Content: "file written", Data: map[string]any{"path": req.Path, "bytes": len(req.Content)}}, nil
}

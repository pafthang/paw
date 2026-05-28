package tools

import (
	"context"
	"encoding/json"
	"os"
)

type FileReadTool struct{}

func (FileReadTool) Name() string { return "file.read" }
func (FileReadTool) Description() string { return "Read a UTF-8 text file from disk." }

func (FileReadTool) Run(ctx context.Context, input json.RawMessage) (ToolResult, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return ToolResult{}, err
	}
	data, err := os.ReadFile(req.Path)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{Content: string(data), Data: map[string]any{"path": req.Path, "bytes": len(data)}}, nil
}

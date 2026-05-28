package tools

import (
	"context"
	"encoding/json"
)

type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input json.RawMessage) (ToolResult, error)
}

type ToolResult struct {
	Content string         `json:"content"`
	Data    map[string]any `json:"data,omitempty"`
}

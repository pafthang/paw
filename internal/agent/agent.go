package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/tools"
	"gorm.io/gorm"
)

type Runner struct {
	DB       *gorm.DB
	Registry *tools.Registry
}

type ToolCall struct {
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type RunRequest struct {
	SessionID uint       `json:"session_id,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type RunResponse struct {
	Results []ToolRunResult `json:"results"`
}

type ToolRunResult struct {
	ToolName string           `json:"tool_name"`
	Result   tools.ToolResult `json:"result,omitempty"`
	Error    string           `json:"error,omitempty"`
}

func NewRunner(database *gorm.DB, registry *tools.Registry) *Runner {
	if registry == nil {
		registry = tools.DefaultRegistry()
	}
	return &Runner{DB: database, Registry: registry}
}

func (r *Runner) Run(ctx context.Context, req RunRequest) (RunResponse, error) {
	if r.DB == nil {
		return RunResponse{}, fmt.Errorf("agent runner requires database")
	}
	if len(req.ToolCalls) == 0 {
		return RunResponse{}, fmt.Errorf("tool_calls is required")
	}
	out := RunResponse{Results: make([]ToolRunResult, 0, len(req.ToolCalls))}
	for _, call := range req.ToolCalls {
		tool, err := r.Registry.Get(call.Name)
		if err != nil {
			result := ToolRunResult{ToolName: call.Name, Error: err.Error()}
			out.Results = append(out.Results, result)
			_, _ = db.CreateToolAuditEvent(r.DB, req.SessionID, call.Name, call, result, err)
			continue
		}
		result, err := tool.Run(ctx, call.Input)
		runResult := ToolRunResult{ToolName: call.Name, Result: result}
		if err != nil {
			runResult.Error = err.Error()
		}
		out.Results = append(out.Results, runResult)
		_, _ = db.CreateToolAuditEvent(r.DB, req.SessionID, call.Name, call, runResult, err)
	}
	return out, nil
}

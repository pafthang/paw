package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/policy"
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
	SessionID           uint       `json:"session_id,omitempty"`
	ToolCalls           []ToolCall `json:"tool_calls"`
	Workspace           string     `json:"workspace,omitempty"`
	AllowShell          bool       `json:"allow_shell,omitempty"`
	AllowShellDangerous bool       `json:"allow_shell_dangerous,omitempty"`
	Iteration           int        `json:"iteration,omitempty"`
	OnTool              ToolHook   `json:"-"`
}

type ToolHook func(event ToolEvent)

type ToolEvent struct {
	Type      string        `json:"type"`
	Index     int           `json:"index"`
	Iteration int           `json:"iteration,omitempty"`
	ToolName  string        `json:"tool_name"`
	Call      ToolCall      `json:"call,omitempty"`
	Result    ToolRunResult `json:"result,omitempty"`
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
	for i, call := range req.ToolCalls {
		emitToolEvent(req.OnTool, ToolEvent{Type: "tool.started", Index: i, Iteration: req.Iteration, ToolName: call.Name, Call: call})

		policyOpts := policy.Options{
			Workspace:           req.Workspace,
			AllowShell:          req.AllowShell,
			AllowShellDangerous: req.AllowShellDangerous,
		}
		if err := policy.CheckToolCall(call.Name, call.Input, policyOpts); err != nil {
			result := ToolRunResult{ToolName: call.Name, Error: err.Error()}
			out.Results = append(out.Results, result)
			_, _ = db.CreateToolAuditEventWithType(r.DB, req.SessionID, "tool.denied", call.Name, call, result, err)
			emitToolEvent(req.OnTool, ToolEvent{Type: "tool.denied", Index: i, Iteration: req.Iteration, ToolName: call.Name, Result: result})
			continue
		}
		if normalized, err := policy.NormalizeToolInput(call.Name, call.Input, policyOpts); err == nil {
			call.Input = normalized
		}

		tool, err := r.Registry.Get(call.Name)
		if err != nil {
			result := ToolRunResult{ToolName: call.Name, Error: err.Error()}
			out.Results = append(out.Results, result)
			_, _ = db.CreateToolAuditEvent(r.DB, req.SessionID, call.Name, call, result, err)
			emitToolEvent(req.OnTool, ToolEvent{Type: "tool.error", Index: i, Iteration: req.Iteration, ToolName: call.Name, Result: result})
			continue
		}
		result, err := tool.Run(ctx, call.Input)
		runResult := ToolRunResult{ToolName: call.Name, Result: result}
		if err != nil {
			runResult.Error = err.Error()
		}
		out.Results = append(out.Results, runResult)
		_, _ = db.CreateToolAuditEvent(r.DB, req.SessionID, call.Name, call, runResult, err)
		if err != nil {
			emitToolEvent(req.OnTool, ToolEvent{Type: "tool.error", Index: i, Iteration: req.Iteration, ToolName: call.Name, Result: runResult})
		} else {
			emitToolEvent(req.OnTool, ToolEvent{Type: "tool.result", Index: i, Iteration: req.Iteration, ToolName: call.Name, Result: runResult})
		}
	}
	return out, nil
}

func emitToolEvent(hook ToolHook, event ToolEvent) {
	if hook != nil {
		hook(event)
	}
}

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"
)

type ShellRunTool struct{}

func (ShellRunTool) Name() string { return "shell.run" }
func (ShellRunTool) Description() string { return "Run a shell command when explicitly allowed." }

func (ShellRunTool) Run(ctx context.Context, input json.RawMessage) (ToolResult, error) {
	var req struct {
		Command string `json:"command"`
		Dir     string `json:"dir,omitempty"`
		Allow   bool   `json:"allow"`
		Timeout int    `json:"timeout_seconds,omitempty"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return ToolResult{}, err
	}
	if !req.Allow {
		return ToolResult{}, errors.New("shell.run requires allow=true")
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", req.Command)
	if req.Dir != "" {
		cmd.Dir = req.Dir
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := map[string]any{"stdout": stdout.String(), "stderr": stderr.String(), "command": req.Command, "dir": req.Dir}
	if err != nil {
		out["error"] = err.Error()
		return ToolResult{Content: stdout.String() + stderr.String(), Data: out}, err
	}
	return ToolResult{Content: stdout.String(), Data: out}, nil
}

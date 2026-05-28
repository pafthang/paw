package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type Options struct {
	Workspace           string
	AllowShell          bool
	AllowShellDangerous bool
}

type Denial struct {
	Tool   string
	Reason string
}

func (d Denial) Error() string { return d.Reason }

func CheckToolCall(toolName string, input json.RawMessage, opts Options) error {
	switch toolName {
	case "file.read", "file.write":
		var req struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return err
		}
		if err := EnsurePathWithinWorkspace(opts.Workspace, req.Path); err != nil {
			return Denial{Tool: toolName, Reason: err.Error()}
		}
		return nil
	case "shell.run":
		if !opts.AllowShell {
			return Denial{Tool: toolName, Reason: "shell.run denied by policy (allow_shell=false)"}
		}
		var req struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return err
		}
		if reason := shellDangerReason(req.Command); reason != "" && !opts.AllowShellDangerous {
			return Denial{Tool: toolName, Reason: "shell.run denied by policy (dangerous command): " + reason}
		}
		return nil
	default:
		return nil
	}
}

func NormalizeToolInput(toolName string, input json.RawMessage, opts Options) (json.RawMessage, error) {
	switch toolName {
	case "file.read":
		var req struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return nil, err
		}
		ws, err := ResolveWorkspace(opts.Workspace)
		if err != nil {
			return nil, err
		}
		resolved, err := ResolveInWorkspace(ws, req.Path)
		if err != nil {
			return nil, err
		}
		req.Path = resolved
		b, err := json.Marshal(req)
		return json.RawMessage(b), err
	case "file.write":
		var req struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return nil, err
		}
		ws, err := ResolveWorkspace(opts.Workspace)
		if err != nil {
			return nil, err
		}
		resolved, err := ResolveInWorkspace(ws, req.Path)
		if err != nil {
			return nil, err
		}
		req.Path = resolved
		b, err := json.Marshal(req)
		return json.RawMessage(b), err
	case "shell.run":
		var req map[string]any
		if err := json.Unmarshal(input, &req); err != nil {
			return nil, err
		}
		if _, ok := req["dir"]; !ok || strings.TrimSpace(fmt.Sprint(req["dir"])) == "" {
			if ws, err := ResolveWorkspace(opts.Workspace); err == nil {
				req["dir"] = ws
			}
		}
		b, err := json.Marshal(req)
		return json.RawMessage(b), err
	default:
		return input, nil
	}
}

func shellDangerReason(command string) string {
	c := strings.ToLower(command)
	c = strings.TrimSpace(c)
	switch {
	case c == "":
		return "empty command"
	case strings.Contains(c, "rm -rf /") || strings.Contains(c, "rm -fr /"):
		return "rm -rf /"
	case strings.Contains(c, "mkfs"):
		return "mkfs"
	case strings.Contains(c, "shutdown") || strings.Contains(c, "reboot"):
		return "shutdown/reboot"
	case strings.Contains(c, "dd if=") && strings.Contains(c, " of=/dev/"):
		return "dd to block device"
	default:
		return ""
	}
}

func EnsurePathWithinWorkspace(workspace, path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path is required")
	}
	if strings.HasPrefix(path, "~") {
		return fmt.Errorf("path %q is outside workspace", path)
	}
	ws, err := ResolveWorkspace(workspace)
	if err != nil {
		return err
	}
	target, err := ResolveInWorkspace(ws, path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(ws, target)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return fmt.Errorf("path %q is outside workspace", path)
	}
	return nil
}

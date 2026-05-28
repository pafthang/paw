package presets

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pafthang/paw/internal/mcp"
)

type MCPPreset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ListMCPPresets() []MCPPreset {
	return []MCPPreset{
		{Name: "filesystem", Description: "MCP filesystem server (npx @modelcontextprotocol/server-filesystem) scoped to a workspace"},
	}
}

func BuildMCPServerConfig(preset string, workspace string) (mcp.ServerConfig, error) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "filesystem":
		ws := strings.TrimSpace(workspace)
		if ws == "" {
			return mcp.ServerConfig{}, errors.New("workspace is required for filesystem preset")
		}
		ws, err := filepath.Abs(ws)
		if err != nil {
			return mcp.ServerConfig{}, err
		}
		return mcp.ServerConfig{
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", ws},
			Env:     map[string]string{},
		}, nil
	default:
		return mcp.ServerConfig{}, fmt.Errorf("unknown preset %q", preset)
	}
}

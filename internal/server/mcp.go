package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/mcp"
)

var serverMCPManager = mcp.NewManager()

func (s *Server) handleMCPList(c echo.Context) error {
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, cfg)
}

func (s *Server) handleMCPPresets(c echo.Context) error {
	cfg, _ := mcp.LoadConfig()
	installed := map[string]bool{}
	for name := range cfg.Servers {
		installed[name] = true
	}
	return c.JSON(http.StatusOK, []map[string]any{
		{
			"id":          "filesystem",
			"name":        "Filesystem",
			"description": "Expose a local workspace through the official filesystem MCP server.",
			"icon":        "folder",
			"category":    "local",
			"package":     "@modelcontextprotocol/server-filesystem",
			"transport":   "stdio",
			"docs_url":    "https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem",
			"needs_args":  true,
			"oauth":       false,
			"installed":   installed["filesystem"],
			"env_keys":    []map[string]any{},
		},
		{
			"id":          "github",
			"name":        "GitHub",
			"description": "Connect to GitHub MCP tooling using a GitHub token.",
			"icon":        "github",
			"category":    "developer",
			"package":     "@modelcontextprotocol/server-github",
			"transport":   "stdio",
			"docs_url":    "https://github.com/modelcontextprotocol/servers/tree/main/src/github",
			"needs_args":  false,
			"oauth":       false,
			"installed":   installed["github"],
			"env_keys": []map[string]any{
				{"key": "GITHUB_PERSONAL_ACCESS_TOKEN", "label": "GitHub token", "required": true, "placeholder": "ghp_...", "secret": true},
			},
		},
		{
			"id":          "sqlite",
			"name":        "SQLite",
			"description": "Connect an SQLite database through MCP.",
			"icon":        "database",
			"category":    "database",
			"package":     "@modelcontextprotocol/server-sqlite",
			"transport":   "stdio",
			"docs_url":    "https://github.com/modelcontextprotocol/servers/tree/main/src/sqlite",
			"needs_args":  true,
			"oauth":       false,
			"installed":   installed["sqlite"],
			"env_keys":    []map[string]any{},
		},
	})
}

func (s *Server) handleMCPInstallPreset(c echo.Context) error {
	var req struct {
		PresetID  string            `json:"preset_id"`
		Env       map[string]string `json:"env,omitempty"`
		ExtraArgs []string          `json:"extra_args,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	if req.PresetID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "preset_id is required"})
	}
	server, ok := mcpPresetServer(req.PresetID, req.Env, req.ExtraArgs)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "unknown preset"})
	}
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	cfg.Servers[req.PresetID] = server
	if err := mcp.SaveConfig(cfg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"status": "installed", "connected": false})
}

func mcpPresetServer(id string, env map[string]string, extraArgs []string) (mcp.ServerConfig, bool) {
	switch id {
	case "filesystem":
		args := append([]string{"-y", "@modelcontextprotocol/server-filesystem"}, extraArgs...)
		return mcp.ServerConfig{Command: "npx", Args: args, Env: env}, true
	case "github":
		args := append([]string{"-y", "@modelcontextprotocol/server-github"}, extraArgs...)
		return mcp.ServerConfig{Command: "npx", Args: args, Env: env}, true
	case "sqlite":
		args := append([]string{"-y", "@modelcontextprotocol/server-sqlite"}, extraArgs...)
		return mcp.ServerConfig{Command: "npx", Args: args, Env: env}, true
	default:
		return mcp.ServerConfig{}, false
	}
}

func (s *Server) handleMCPShow(c echo.Context) error {
	name := c.Param("name")
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := mcp.ValidateName(name); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	item, ok := cfg.Servers[name]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	return c.JSON(http.StatusOK, map[string]any{"name": name, "config": item})
}

func (s *Server) handleMCPAdd(c echo.Context) error {
	var req struct {
		Name    string            `json:"name"`
		Command string            `json:"command"`
		Args    []string          `json:"args,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	if err := mcp.ValidateName(req.Name); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	cfg.Servers[req.Name] = mcp.ServerConfig{Command: req.Command, Args: req.Args, Env: req.Env}
	if err := mcp.SaveConfig(cfg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"saved": true, "name": req.Name})
}

func (s *Server) handleMCPRemove(c echo.Context) error {
	name := c.Param("name")
	if err := mcp.ValidateName(name); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	delete(cfg.Servers, name)
	if err := mcp.SaveConfig(cfg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"removed": true, "name": name})
}

func (s *Server) handleMCPStart(c echo.Context) error {
	name := c.Param("name")
	cfg, err := mcp.LoadConfig()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	item, ok := cfg.Servers[name]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}
	st, err := serverMCPManager.Start(c.Request().Context(), name, item)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, st)
}

func (s *Server) handleMCPStop(c echo.Context) error {
	name := c.Param("name")
	st, err := serverMCPManager.Stop(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, st)
}

func (s *Server) handleMCPStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, serverMCPManager.Status())
}

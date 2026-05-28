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

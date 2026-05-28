package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/agent"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/tools"
)

func (s *Server) handleListTools(c echo.Context) error {
	return c.JSON(http.StatusOK, tools.DefaultRegistry().List())
}

func (s *Server) handleAgentRun(c echo.Context) error {
	var req agent.RunRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	runner := agent.NewRunner(database, tools.DefaultRegistry())
	resp, err := runner.Run(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleListAudit(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	events, err := db.ListAuditEvents(database, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	decoded := make([]map[string]any, 0, len(events))
	for _, event := range events {
		item := map[string]any{
			"id":         event.ID,
			"session_id": event.SessionID,
			"type":       event.Type,
			"tool_name":  event.ToolName,
			"error":      event.Error,
			"created_at": event.CreatedAt,
		}
		var input any
		if event.InputJSON != "" && json.Unmarshal([]byte(event.InputJSON), &input) == nil {
			item["input"] = input
		} else {
			item["input_json"] = event.InputJSON
		}
		var output any
		if event.OutputJSON != "" && json.Unmarshal([]byte(event.OutputJSON), &output) == nil {
			item["output"] = output
		} else {
			item["output_json"] = event.OutputJSON
		}
		decoded = append(decoded, item)
	}
	return c.JSON(http.StatusOK, decoded)
}

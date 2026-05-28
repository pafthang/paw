package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/health"
)

func (s *Server) healthSummary(c echo.Context) map[string]any {
	report := health.Run(c.Request().Context(), s.settings)
	issues := make([]map[string]any, 0, len(report.Checks))
	for _, check := range report.Checks {
		status := "ok"
		severity := "warning"
		switch check.Status {
		case "fail", "critical":
			status = "critical"
			severity = "critical"
		case "warn", "warning":
			status = "warning"
			severity = "warning"
		}
		issues = append(issues, map[string]any{
			"check_id":  check.Name,
			"name":      check.Name,
			"category":  "connectivity",
			"status":    status,
			"severity":  severity,
			"message":   check.Message,
			"fix_hint":  "",
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
			"details":   []string{},
		})
	}
	return map[string]any{
		"status":      report.Status,
		"check_count": len(report.Checks),
		"issues":      issues,
		"last_check":  time.Now().UTC().Format(time.RFC3339Nano),
		"checks":      report.Checks,
	}
}

func (s *Server) handleHealthCompat(c echo.Context) error {
	return c.JSON(http.StatusOK, s.healthSummary(c))
}

func (s *Server) handleHealthErrors(c echo.Context) error {
	return c.JSON(http.StatusOK, []map[string]any{})
}

func (s *Server) handleClearHealthErrors(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

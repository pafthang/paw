package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/skills"
)

func (s *Server) handleListSkills(c echo.Context) error {
	report, err := skills.LoadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, report.Skills)
}

func (s *Server) handleSearchSkills(c echo.Context) error {
	q := strings.ToLower(strings.TrimSpace(c.QueryParam("q")))
	report, err := skills.LoadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if q == "" {
		return c.JSON(http.StatusOK, map[string]any{"results": report.Skills})
	}
	results := make([]skills.Skill, 0)
	for _, skill := range report.Skills {
		if strings.Contains(strings.ToLower(skill.Name), q) || strings.Contains(strings.ToLower(skill.Description), q) {
			results = append(results, skill)
		}
	}
	return c.JSON(http.StatusOK, map[string]any{"results": results})
}

func (s *Server) handleGetSkill(c echo.Context) error {
	name := c.Param("name")
	skill, err := skills.LoadByName(name)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, skill)
}

func (s *Server) handleReloadSkills(c echo.Context) error {
	report, err := skills.LoadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, report)
}

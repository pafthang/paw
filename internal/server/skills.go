package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/skills"
)

func (s *Server) handleListSkills(c echo.Context) error {
	report, err := skills.LoadAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, report)
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

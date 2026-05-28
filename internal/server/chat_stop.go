package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) handleChatStop(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"stopped": true})
}

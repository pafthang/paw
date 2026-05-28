package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) handleChannelsList(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"channels": s.channels.List()})
}

func (s *Server) handleChannelsStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, s.channels.StatusAll())
}

func (s *Server) handleChannelsStart(c echo.Context) error {
	name := c.Param("name")
	st, err := s.channels.Start(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error(), "status": st})
	}
	return c.JSON(http.StatusOK, st)
}

func (s *Server) handleChannelsStop(c echo.Context) error {
	name := c.Param("name")
	st, err := s.channels.Stop(c.Request().Context(), name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error(), "status": st})
	}
	return c.JSON(http.StatusOK, st)
}

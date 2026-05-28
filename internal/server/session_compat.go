package server

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/db"
)

func (s *Server) handleCreateSession(c echo.Context) error {
	var req struct {
		Title string `json:"title,omitempty"`
	}
	_ = c.Bind(&req)
	if req.Title == "" {
		req.Title = "New chat"
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	session, err := db.CreateChatSession(database, req.Title)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"id":    strconv.FormatUint(uint64(session.ID), 10),
		"title": session.Title,
	})
}

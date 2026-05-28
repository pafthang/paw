package server

import (
	"net/http"

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
		"id":    stringID(session.ID),
		"title": session.Title,
	})
}

func stringID(id uint) string {
	return fmt.Sprintf("%d", id)
}

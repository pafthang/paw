package server

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	pawauth "github.com/pafthang/paw/internal/auth"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) handleWS(c echo.Context) error {
	if !pawauth.Check(extractAccessToken(c)) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid access token"})
	}
	conn, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	welcome := map[string]any{
		"type":    "hello",
		"service": "paw",
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	_ = conn.WriteJSON(welcome)
	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			return nil
		}
		msg["type"] = "echo"
		msg["time"] = time.Now().UTC().Format(time.RFC3339)
		if err := conn.WriteJSON(msg); err != nil {
			return nil
		}
	}
}

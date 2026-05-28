package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	pawauth "github.com/pafthang/paw/internal/auth"
)

func accessTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		switch path {
		case "/", "/api/v1/health", "/api/v1/status", "/ws":
			return next(c)
		}
		if !strings.HasPrefix(path, "/api/") {
			return next(c)
		}
		if pawauth.Check(extractAccessToken(c)) {
			return next(c)
		}
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing or invalid access token"})
	}
}

func extractAccessToken(c echo.Context) string {
	if token := c.Request().Header.Get("X-Paw-Access-Token"); token != "" {
		return token
	}
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[len("Bearer "):])
	}
	if token := c.QueryParam("access_token"); token != "" {
		return token
	}
	return ""
}

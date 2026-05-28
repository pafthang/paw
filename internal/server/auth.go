package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	pawauth "github.com/pafthang/paw/internal/auth"
)

const sessionCookieName = "pocketpaw_session"

func accessTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		if c.Request().Method == http.MethodOptions {
			return next(c)
		}
		switch path {
		case "/", "/api/v1/health", "/api/v1/status", "/ws", "/api/v1/ws", "/api/v1/auth/login", "/api/v1/auth/logout":
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
	if token := c.QueryParam("token"); token != "" {
		return token
	}
	if cookie, err := c.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func (s *Server) handleAuthLogin(c echo.Context) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"detail": "Invalid JSON body"})
	}
	token := strings.TrimSpace(req.Token)
	if !pawauth.Check(token) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "Invalid access token"})
	}
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   c.IsTLS(),
	})
	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAuthLogout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   c.IsTLS(),
	})
	return c.JSON(http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAuthSession(c echo.Context) error {
	token := extractAccessToken(c)
	if !pawauth.Check(token) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"detail": "Invalid master token"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"session_token":    token,
		"expires_in_hours": 24,
	})
}

func (s *Server) handleTokenRegenerate(c echo.Context) error {
	token, err := pawauth.RotateToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"detail": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

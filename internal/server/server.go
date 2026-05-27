package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/health"
	"github.com/pafthang/paw/internal/llm"
)

type Server struct {
	settings config.Settings
	echo     *echo.Echo
	addr     string
}

func New(settings config.Settings) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(securityHeaders)

	s := &Server{
		settings: settings,
		echo:     e,
		addr:     net.JoinHostPort(settings.WebHost, fmt.Sprintf("%d", settings.WebPort)),
	}

	e.GET("/", s.handleIndex)
	e.GET("/api/v1/health", s.handleHealth)
	e.GET("/api/v1/status", s.handleStatus)
	e.GET("/api/v1/settings", s.handleSettings)
	e.POST("/api/v1/chat", s.handleChat)

	return s
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting paw Go API server", "addr", s.addr)
		errCh <- s.echo.Start(s.addr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.echo.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) handleIndex(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"name":    "paw",
		"mode":    "go-core-stage3",
		"message": "PocketPaw Go core stage 3 is running.",
		"stack": []string{
			"cobra",
			"echo",
			"gorm",
			"sqlite",
		},
		"routes": []string{
			"GET /api/v1/health",
			"GET /api/v1/status",
			"GET /api/v1/settings",
			"POST /api/v1/chat",
		},
	})
}

func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, health.Run(c.Request().Context(), s.settings))
}

func (s *Server) handleStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status":        "ok",
		"implementation": "go",
		"stage":         "core-stage3",
		"web_host":      s.settings.WebHost,
		"web_port":      s.settings.WebPort,
		"agent_backend": s.settings.AgentBackend,
		"model":         s.settings.Model,
		"stack":         []string{"cobra", "echo", "gorm", "sqlite"},
	})
}

func (s *Server) handleSettings(c echo.Context) error {
	masked := s.settings
	if masked.OpenAIAPIKey != "" {
		masked.OpenAIAPIKey = "***"
	}
	if masked.AnthropicAPIKey != "" {
		masked.AnthropicAPIKey = "***"
	}
	if masked.TelegramBotToken != "" {
		masked.TelegramBotToken = "***"
	}
	return c.JSON(http.StatusOK, masked)
}

func (s *Server) handleChat(c echo.Context) error {
	var req struct {
		Model    string        `json:"model,omitempty"`
		Prompt   string        `json:"prompt,omitempty"`
		Messages []llm.Message `json:"messages,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	model := req.Model
	if model == "" {
		model = llm.DefaultModel(s.settings)
	}
	messages := req.Messages
	if len(messages) == 0 && req.Prompt != "" {
		messages = []llm.Message{{Role: "user", Content: req.Prompt}}
	}
	if len(messages) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "prompt or messages is required"})
	}
	client, err := llm.NewClient(s.settings)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp, err := client.Chat(c.Request().Context(), llm.ChatRequest{Model: model, Messages: messages})
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func securityHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		res := c.Response()
		res.Header().Set("X-Content-Type-Options", "nosniff")
		res.Header().Set("X-Frame-Options", "DENY")
		res.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return next(c)
	}
}

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pafthang/paw/internal/channels"
	tg "github.com/pafthang/paw/internal/channels/telegram"
	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/contextpack"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
)

type Server struct {
	settings config.Settings
	echo     *echo.Echo
	addr     string
	channels *channels.Manager
}

func New(settings config.Settings) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(securityHeaders)
	if len(settings.CORSAllowedOrigins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     settings.CORSAllowedOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
			AllowHeaders:     []string{"Authorization", "Content-Type", "X-Paw-Access-Token"},
			AllowCredentials: true,
		}))
	}
	e.Use(accessTokenMiddleware)

	s := &Server{
		settings: settings,
		echo:     e,
		addr:     net.JoinHostPort(settings.WebHost, fmt.Sprintf("%d", settings.WebPort)),
		channels: channels.NewManager(),
	}
	s.channels.Register(tg.New(settings))

	e.GET("/", s.handleIndex)
	e.GET("/ws", s.handleWS)
	e.GET("/api/v1/ws", s.handleWS)
	e.POST("/api/v1/auth/login", s.handleAuthLogin)
	e.POST("/api/v1/auth/logout", s.handleAuthLogout)
	e.POST("/api/v1/auth/session", s.handleAuthSession)
	e.POST("/api/v1/token/regenerate", s.handleTokenRegenerate)
	e.GET("/api/v1/health", s.handleHealth)
	e.GET("/api/v1/status", s.handleStatus)
	e.GET("/api/v1/settings", s.handleSettings)
	e.POST("/api/v1/chat", s.handleChat)
	e.GET("/api/v1/sessions", s.handleListSessions)
	e.GET("/api/v1/sessions/:id", s.handleGetSession)
	e.DELETE("/api/v1/sessions/:id", s.handleDeleteSession)
	e.GET("/api/v1/sessions/search", s.handleSearchSessions)
	e.PATCH("/api/v1/sessions/:id", s.handlePatchSession)
	e.GET("/api/v1/tools", s.handleListTools)
	e.POST("/api/v1/agent/run", s.handleAgentRun)
	e.POST("/api/v1/agent/chat", s.handleAgentChat)
	e.GET("/api/v1/audit", s.handleListAudit)
	e.GET("/api/v1/memory", s.handleListMemory)
	e.POST("/api/v1/memory", s.handleCreateMemory)
	e.GET("/api/v1/memory/:id", s.handleGetMemory)
	e.DELETE("/api/v1/memory/:id", s.handleDeleteMemory)
	e.GET("/api/v1/memory/search", s.handleSearchMemory)
	e.GET("/api/v1/files", s.handleListFiles)
	e.POST("/api/v1/files", s.handleCreateFile)
	e.GET("/api/v1/files/:id", s.handleGetFile)
	e.DELETE("/api/v1/files/:id", s.handleDeleteFile)
	e.GET("/api/v1/files/search", s.handleSearchFiles)
	e.GET("/api/v1/search", s.handleSearch)
	e.GET("/api/v1/skills", s.handleListSkills)
	e.GET("/api/v1/skills/:name", s.handleGetSkill)
	e.POST("/api/v1/skills/reload", s.handleReloadSkills)
	e.GET("/api/v1/mcp", s.handleMCPList)
	e.GET("/api/v1/mcp/:name", s.handleMCPShow)
	e.POST("/api/v1/mcp", s.handleMCPAdd)
	e.DELETE("/api/v1/mcp/:name", s.handleMCPRemove)
	e.POST("/api/v1/mcp/:name/start", s.handleMCPStart)
	e.POST("/api/v1/mcp/:name/stop", s.handleMCPStop)
	e.GET("/api/v1/mcp/status", s.handleMCPStatus)
	e.GET("/api/v1/channels", s.handleChannelsList)
	e.GET("/api/v1/channels/status", s.handleChannelsStatus)
	e.POST("/api/v1/channels/:name/start", s.handleChannelsStart)
	e.POST("/api/v1/channels/:name/stop", s.handleChannelsStop)
	s.registerCompatRoutes()

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
		"mode":    "go-core-stage7-telegram",
		"message": "PocketPaw Go core API compatibility layer is running.",
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
			"GET /ws",
			"GET /api/v1/ws",
			"POST /api/v1/auth/login",
			"POST /api/v1/auth/logout",
			"POST /api/v1/auth/session",
			"POST /api/v1/token/regenerate",
			"POST /api/v1/chat",
			"GET /api/v1/sessions",
			"GET /api/v1/sessions/:id",
			"GET /api/v1/sessions/search?q=...",
			"PATCH /api/v1/sessions/:id",
			"DELETE /api/v1/sessions/:id",
			"GET /api/v1/tools",
			"POST /api/v1/agent/run",
			"POST /api/v1/agent/chat",
			"GET /api/v1/audit",
			"GET /api/v1/memory",
			"POST /api/v1/memory",
			"GET /api/v1/memory/:id",
			"DELETE /api/v1/memory/:id",
			"GET /api/v1/memory/search?q=...",
			"GET /api/v1/files",
			"POST /api/v1/files",
			"GET /api/v1/files/:id",
			"DELETE /api/v1/files/:id",
			"GET /api/v1/files/search?q=...",
			"GET /api/v1/search?q=...",
			"GET /api/v1/skills",
			"GET /api/v1/skills/:name",
			"POST /api/v1/skills/reload",
			"GET /api/v1/mcp",
			"GET /api/v1/mcp/:name",
			"POST /api/v1/mcp",
			"DELETE /api/v1/mcp/:name",
			"POST /api/v1/mcp/:name/start",
			"POST /api/v1/mcp/:name/stop",
			"GET /api/v1/mcp/status",
			"GET /api/v1/channels",
			"GET /api/v1/channels/status",
			"POST /api/v1/channels/:name/start",
			"POST /api/v1/channels/:name/stop",
			"GET /api/v1/identity",
			"PUT /api/v1/identity",
			"GET /api/v1/kits",
			"GET /api/v1/kits/catalog",
			"GET /api/mission-control/agents",
			"GET /api/mission-control/notifications",
			"GET /api/mission-control/tasks/running",
		},
	})
}

func (s *Server) handleHealth(c echo.Context) error {
	return s.handleHealthCompat(c)
}

func (s *Server) handleStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status":         "ok",
		"implementation": "go",
		"stage":          "stage7-telegram",
		"web_host":       s.settings.WebHost,
		"web_port":       s.settings.WebPort,
		"agent_backend":  s.settings.AgentBackend,
		"model":          s.settings.Model,
		"stack":          []string{"cobra", "echo", "gorm", "sqlite"},
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
		Model           string        `json:"model,omitempty"`
		Prompt          string        `json:"prompt,omitempty"`
		Messages        []llm.Message `json:"messages,omitempty"`
		SessionID       uint          `json:"session_id,omitempty"`
		HistoryLimit    int           `json:"history_limit,omitempty"`
		SystemPrompt    string        `json:"system_prompt,omitempty"`
		MaxContextChars int           `json:"max_context_chars,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
	}
	model := req.Model
	if model == "" {
		model = llm.DefaultModel(s.settings)
	}
	incomingMessages := req.Messages
	if len(incomingMessages) == 0 && req.Prompt != "" {
		incomingMessages = []llm.Message{{Role: "user", Content: req.Prompt}}
	}
	if len(incomingMessages) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "prompt or messages is required"})
	}

	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	var session *db.ChatSession
	history := make([]llm.Message, 0, db.DefaultHistoryLimit)
	if req.SessionID > 0 {
		session, err = db.GetChatSession(database, req.SessionID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		historyLimit := req.HistoryLimit
		if historyLimit == 0 {
			historyLimit = db.DefaultHistoryLimit
		}
		recent, err := db.ListRecentChatMessages(database, session.ID, historyLimit)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		history = append(history, toLLMMessages(recent)...)
	} else {
		session, err = db.CreateChatSession(database, firstUserContent(incomingMessages))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	messages := contextpack.Pack(req.SystemPrompt, history, incomingMessages, req.MaxContextChars)

	client, err := llm.NewClient(s.settings)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp, err := client.Chat(c.Request().Context(), llm.ChatRequest{Model: model, Messages: messages})
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	for _, message := range incomingMessages {
		if _, err := db.AddChatMessage(database, session.ID, message.Role, message.Content, model); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	if _, err := db.AddChatMessage(database, session.ID, "assistant", resp.Content, resp.Model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"session_id":       session.ID,
		"history_messages": len(messages) - len(incomingMessages) - 1,
		"context":          contextpack.Stats(messages),
		"response":         resp,
	})
}

func (s *Server) handleListSessions(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	sessions, err := db.ListChatSessions(database, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sessions)
}

func (s *Server) handleGetSession(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid session id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	session, err := db.GetChatSession(database, uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, session)
}

func (s *Server) handleDeleteSession(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid session id"})
	}
	database, err := db.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if err := db.DeleteChatSession(database, uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func parseUintParam(c echo.Context, name string) (uint, error) {
	parsed, err := strconv.ParseUint(c.Param(name), 10, 64)
	return uint(parsed), err
}

func firstUserContent(messages []llm.Message) string {
	for _, message := range messages {
		if message.Role == "user" && message.Content != "" {
			return message.Content
		}
	}
	if len(messages) > 0 {
		return messages[0].Content
	}
	return "New chat"
}

func toLLMMessages(messages []db.ChatMessage) []llm.Message {
	out := make([]llm.Message, 0, len(messages))
	for _, message := range messages {
		if message.Role == "" || message.Content == "" {
			continue
		}
		out = append(out, llm.Message{Role: message.Role, Content: message.Content})
	}
	return out
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

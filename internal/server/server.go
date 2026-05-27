package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/pafthang/paw/internal/config"
	"github.com/pafthang/paw/internal/health"
	"github.com/pafthang/paw/internal/llm"
)

type Server struct {
	settings config.Settings
	http     *http.Server
}

func New(settings config.Settings) *Server {
	mux := http.NewServeMux()
	s := &Server{settings: settings}
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/status", s.handleStatus)
	mux.HandleFunc("GET /api/v1/settings", s.handleSettings)
	mux.HandleFunc("POST /api/v1/chat", s.handleChat)
	mux.HandleFunc("GET /", s.handleIndex)
	s.http = &http.Server{
		Addr:              net.JoinHostPort(settings.WebHost, fmt.Sprintf("%d", settings.WebPort)),
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting paw Go API server", "addr", s.http.Addr)
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "paw",
		"mode":    "go-core-stage2",
		"message": "PocketPaw Go core stage 2 is running.",
		"routes": []string{
			"GET /api/v1/health",
			"GET /api/v1/status",
			"GET /api/v1/settings",
			"POST /api/v1/chat",
		},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, health.Run(r.Context(), s.settings))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ok",
		"implementation": "go",
		"stage":         "core-stage2",
		"web_host":      s.settings.WebHost,
		"web_port":      s.settings.WebPort,
		"agent_backend": s.settings.AgentBackend,
		"model":         s.settings.Model,
	})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, masked)
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model    string        `json:"model,omitempty"`
		Prompt   string        `json:"prompt,omitempty"`
		Messages []llm.Message `json:"messages,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
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
		writeError(w, http.StatusBadRequest, "prompt or messages is required")
		return
	}
	client, err := llm.NewClient(s.settings)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := client.Chat(r.Context(), llm.ChatRequest{Model: model, Messages: messages})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

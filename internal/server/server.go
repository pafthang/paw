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
		"mode":    "go-core-stage1",
		"message": "PocketPaw Go core stage 1 is running.",
		"routes": []string{
			"GET /api/v1/health",
			"GET /api/v1/status",
			"GET /api/v1/settings",
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
		"stage":         "core-stage1",
		"web_host":      s.settings.WebHost,
		"web_port":      s.settings.WebPort,
		"agent_backend": s.settings.AgentBackend,
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

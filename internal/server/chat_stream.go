package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/contextpack"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
)

func (s *Server) handleChatStream(c echo.Context) error {
	var req struct {
		Content         string `json:"content"`
		SessionID       any    `json:"session_id,omitempty"`
		SystemPrompt    string `json:"system_prompt,omitempty"`
		MaxContextChars int    `json:"max_context_chars,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return writeSSEError(c, http.StatusBadRequest, "invalid JSON: "+err.Error())
	}
	if req.Content == "" {
		return writeSSEError(c, http.StatusBadRequest, "content is required")
	}

	res := c.Response()
	res.Header().Set(echo.HeaderContentType, "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")
	res.Header().Set("X-Accel-Buffering", "no")
	res.WriteHeader(http.StatusOK)

	stream := func(event string, data any) error {
		payload, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(res, "event: %s\n", event); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(res, "data: %s\n\n", payload); err != nil {
			return err
		}
		res.Flush()
		return nil
	}

	database, err := db.Open()
	if err != nil {
		_ = stream("error", map[string]string{"detail": err.Error()})
		return nil
	}

	model := llm.DefaultModel(s.settings)
	incoming := []llm.Message{{Role: "user", Content: req.Content}}
	var session *db.ChatSession
	history := make([]llm.Message, 0, db.DefaultHistoryLimit)

	if id := parseStreamSessionID(req.SessionID); id > 0 {
		existing, err := db.GetChatSession(database, id)
		if err == nil {
			session = existing
			recent, err := db.ListRecentChatMessages(database, session.ID, db.DefaultHistoryLimit)
			if err == nil {
				history = append(history, toLLMMessages(recent)...)
			}
		}
	}
	if session == nil {
		session, err = db.CreateChatSession(database, req.Content)
		if err != nil {
			_ = stream("error", map[string]string{"detail": err.Error()})
			return nil
		}
	}

	messages := contextpack.Pack(req.SystemPrompt, history, incoming, req.MaxContextChars)
	client, err := llm.NewClient(s.settings)
	if err != nil {
		_ = stream("error", map[string]string{"detail": err.Error()})
		return nil
	}
	resp, err := client.Chat(c.Request().Context(), llm.ChatRequest{Model: model, Messages: messages})
	if err != nil {
		_ = stream("error", map[string]string{"detail": err.Error()})
		return nil
	}

	if _, err := db.AddChatMessage(database, session.ID, "user", req.Content, model); err != nil {
		_ = stream("error", map[string]string{"detail": err.Error()})
		return nil
	}
	if _, err := db.AddChatMessage(database, session.ID, "assistant", resp.Content, resp.Model); err != nil {
		_ = stream("error", map[string]string{"detail": err.Error()})
		return nil
	}

	if err := stream("chunk", map[string]string{"content": resp.Content, "type": "text"}); err != nil {
		return nil
	}
	_ = stream("stream_end", map[string]any{
		"session_id": strconv.FormatUint(uint64(session.ID), 10),
		"usage": map[string]any{
			"input_tokens":         0,
			"output_tokens":        0,
			"cached_input_tokens":  0,
			"total_tokens":         0,
		},
	})
	return nil
}

func writeSSEError(c echo.Context, status int, detail string) error {
	return c.JSON(status, map[string]string{"detail": detail})
}

func parseStreamSessionID(raw any) uint {
	switch v := raw.(type) {
	case float64:
		if v > 0 {
			return uint(v)
		}
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			return uint(id)
		}
	}
	return 0
}

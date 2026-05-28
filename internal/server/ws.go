package server

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pafthang/paw/internal/agent"
	pawauth "github.com/pafthang/paw/internal/auth"
	"github.com/pafthang/paw/internal/contextpack"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsRequest struct {
	Type            string        `json:"type"`
	ID              string        `json:"id,omitempty"`
	Prompt          string        `json:"prompt,omitempty"`
	Model           string        `json:"model,omitempty"`
	Messages        []llm.Message `json:"messages,omitempty"`
	SessionID       uint          `json:"session_id,omitempty"`
	HistoryLimit    int           `json:"history_limit,omitempty"`
	SystemPrompt    string        `json:"system_prompt,omitempty"`
	MaxContextChars int           `json:"max_context_chars,omitempty"`
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

	_ = conn.WriteJSON(wsEvent("hello", "", map[string]any{
		"service": "paw",
		"time":    time.Now().UTC().Format(time.RFC3339),
		"commands": []string{
			"chat",
			"agent.chat",
			"ping",
			"echo",
		},
	}))
	for {
		var req wsRequest
		if err := conn.ReadJSON(&req); err != nil {
			return nil
		}
		switch req.Type {
		case "ping":
			_ = conn.WriteJSON(wsEvent("pong", req.ID, map[string]any{"time": time.Now().UTC().Format(time.RFC3339)}))
		case "chat":
			s.handleWSChat(conn, req)
		case "agent.chat":
			s.handleWSAgentChat(conn, req)
		default:
			_ = conn.WriteJSON(wsEvent("echo", req.ID, map[string]any{"request": req, "time": time.Now().UTC().Format(time.RFC3339)}))
		}
	}
}

func (s *Server) handleWSChat(conn *websocket.Conn, req wsRequest) {
	_ = conn.WriteJSON(wsEvent("chat.started", req.ID, map[string]any{"session_id": req.SessionID}))
	client, err := llm.NewClient(s.settings)
	if err != nil {
		_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
		return
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
		_ = conn.WriteJSON(wsEvent("chat.error", req.ID, map[string]any{"error": "prompt or messages is required"}))
		return
	}
	database, err := db.Open()
	if err != nil {
		_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
		return
	}
	var session *db.ChatSession
	history := make([]llm.Message, 0, db.DefaultHistoryLimit)
	if req.SessionID > 0 {
		session, err = db.GetChatSession(database, req.SessionID)
		if err != nil {
			_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
			return
		}
		historyLimit := req.HistoryLimit
		if historyLimit == 0 {
			historyLimit = db.DefaultHistoryLimit
		}
		recent, err := db.ListRecentChatMessages(database, session.ID, historyLimit)
		if err != nil {
			_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
			return
		}
		history = append(history, toLLMMessages(recent)...)
	} else {
		session, err = db.CreateChatSession(database, firstUserContent(incomingMessages))
		if err != nil {
			_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
			return
		}
	}
	messages := contextpack.Pack(req.SystemPrompt, history, incomingMessages, req.MaxContextChars)
	resp, err := client.Chat(nilContext(), llm.ChatRequest{Model: model, Messages: messages})
	if err != nil {
		_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
		return
	}
	for _, message := range incomingMessages {
		if _, err := db.AddChatMessage(database, session.ID, message.Role, message.Content, model); err != nil {
			_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
			return
		}
	}
	if _, err := db.AddChatMessage(database, session.ID, "assistant", resp.Content, resp.Model); err != nil {
		_ = conn.WriteJSON(wsError("chat.error", req.ID, err))
		return
	}
	_ = conn.WriteJSON(wsEvent("chat.result", req.ID, map[string]any{
		"session_id":       session.ID,
		"history_messages": len(messages) - len(incomingMessages) - 1,
		"context":          contextpack.Stats(messages),
		"response":         resp,
	}))
}

func (s *Server) handleWSAgentChat(conn *websocket.Conn, req wsRequest) {
	_ = conn.WriteJSON(wsEvent("agent.started", req.ID, map[string]any{"session_id": req.SessionID}))
	client, err := llm.NewClient(s.settings)
	if err != nil {
		_ = conn.WriteJSON(wsError("agent.error", req.ID, err))
		return
	}
	database, err := db.Open()
	if err != nil {
		_ = conn.WriteJSON(wsError("agent.error", req.ID, err))
		return
	}
	model := req.Model
	if model == "" {
		model = llm.DefaultModel(s.settings)
	}
	runner := agent.NewDefaultRunner(database)
	resp, err := runner.Chat(nilContext(), client, agent.ChatRequest{
		SessionID:       req.SessionID,
		Prompt:          req.Prompt,
		Model:           model,
		HistoryLimit:    req.HistoryLimit,
		MaxContextChars: req.MaxContextChars,
		SystemPrompt:    req.SystemPrompt,
	})
	if err != nil {
		_ = conn.WriteJSON(wsError("agent.error", req.ID, err))
		return
	}
	if resp.UsedTools {
		_ = conn.WriteJSON(wsEvent("agent.tools", req.ID, map[string]any{"tool_calls": resp.ToolCalls, "tool_run_response": resp.ToolRunResponse}))
	}
	_ = conn.WriteJSON(wsEvent("agent.result", req.ID, map[string]any{"response": resp}))
}

func wsEvent(eventType string, id string, payload map[string]any) map[string]any {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["type"] = eventType
	payload["id"] = id
	payload["time"] = time.Now().UTC().Format(time.RFC3339)
	return payload
}

func wsError(eventType string, id string, err error) map[string]any {
	return wsEvent(eventType, id, map[string]any{"error": err.Error()})
}

func nilContext() contextShim { return contextShim{} }

type contextShim struct{}

func (contextShim) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }
func (contextShim) Done() <-chan struct{} { return nil }
func (contextShim) Err() error { return nil }
func (contextShim) Value(key any) any { return nil }

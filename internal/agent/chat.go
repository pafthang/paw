package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pafthang/paw/internal/contextpack"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/tools"
	"gorm.io/gorm"
)

const ToolCallSystemPrompt = `You are Paw, a local coding assistant with access to tools.

When you need tools, respond with strict JSON only:
{
  "tool_calls": [
    {"name":"file.read","input":{"path":"README.md"}}
  ]
}

Available tools:
- file.read: read a UTF-8 text file from disk. input: {"path":"..."}
- file.write: write UTF-8 text to disk. input: {"path":"...","content":"..."}
- shell.run: run a shell command only when explicitly allowed. input: {"command":"...","allow":true,"timeout_seconds":30}

If you do not need tools, answer normally in plain text.`

const FinalAnswerSystemPrompt = `You are Paw, a local coding assistant.
You have just received tool results. Use them to answer the user's original request.
Do not emit more tool_calls JSON in this final answer. Be concise and practical.`

type ChatRequest struct {
	SessionID       uint     `json:"session_id,omitempty"`
	Prompt          string   `json:"prompt"`
	Model           string   `json:"model,omitempty"`
	HistoryLimit    int      `json:"history_limit,omitempty"`
	MaxContextChars int      `json:"max_context_chars,omitempty"`
	SystemPrompt    string   `json:"system_prompt,omitempty"`
	OnTool          ToolHook `json:"-"`
}

type ChatResponse struct {
	SessionID       uint             `json:"session_id"`
	ModelResponse   llm.ChatResponse `json:"model_response"`
	FinalResponse   llm.ChatResponse `json:"final_response,omitempty"`
	ToolCalls       []ToolCall       `json:"tool_calls,omitempty"`
	ToolRunResponse RunResponse      `json:"tool_run_response,omitempty"`
	UsedTools       bool             `json:"used_tools"`
}

func (r *Runner) Chat(ctx context.Context, client llm.Client, req ChatRequest) (ChatResponse, error) {
	if r.DB == nil {
		return ChatResponse{}, fmt.Errorf("agent runner requires database")
	}
	if client == nil {
		return ChatResponse{}, fmt.Errorf("agent chat requires llm client")
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return ChatResponse{}, fmt.Errorf("prompt is required")
	}

	session, history, err := r.prepareSession(req)
	if err != nil {
		return ChatResponse{}, err
	}
	incoming := []llm.Message{{Role: "user", Content: prompt}}
	systemPrompt := req.SystemPrompt
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = ToolCallSystemPrompt
	}
	messages := contextpack.Pack(systemPrompt, history, incoming, req.MaxContextChars)
	modelResp, err := client.Chat(ctx, llm.ChatRequest{Model: req.Model, Messages: messages})
	if err != nil {
		return ChatResponse{}, err
	}

	calls := ParseToolCalls(modelResp.Content)
	var runResp RunResponse
	finalResp := modelResp
	if len(calls) > 0 {
		runResp, err = r.Run(ctx, RunRequest{SessionID: session.ID, ToolCalls: calls, OnTool: req.OnTool})
		if err != nil {
			return ChatResponse{}, err
		}
		finalResp, err = r.finalizeWithToolResults(ctx, client, req, history, prompt, modelResp, runResp)
		if err != nil {
			return ChatResponse{}, err
		}
	}
	if _, err := db.AddChatMessage(r.DB, session.ID, "user", prompt, req.Model); err != nil {
		return ChatResponse{}, err
	}
	if len(calls) > 0 {
		if _, err := db.AddChatMessage(r.DB, session.ID, "assistant", finalResp.Content, finalResp.Model); err != nil {
			return ChatResponse{}, err
		}
	} else {
		if _, err := db.AddChatMessage(r.DB, session.ID, "assistant", modelResp.Content, modelResp.Model); err != nil {
			return ChatResponse{}, err
		}
	}
	return ChatResponse{SessionID: session.ID, ModelResponse: modelResp, FinalResponse: finalResp, ToolCalls: calls, ToolRunResponse: runResp, UsedTools: len(calls) > 0}, nil
}

func (r *Runner) finalizeWithToolResults(ctx context.Context, client llm.Client, req ChatRequest, history []llm.Message, prompt string, modelResp llm.ChatResponse, runResp RunResponse) (llm.ChatResponse, error) {
	toolJSON, err := json.MarshalIndent(runResp, "", "  ")
	if err != nil {
		return llm.ChatResponse{}, err
	}
	incoming := []llm.Message{
		{Role: "user", Content: prompt},
		{Role: "assistant", Content: modelResp.Content},
		{Role: "user", Content: "Tool results:\n" + string(toolJSON) + "\n\nNow answer my original request using these results."},
	}
	messages := contextpack.Pack(FinalAnswerSystemPrompt, history, incoming, req.MaxContextChars)
	return client.Chat(ctx, llm.ChatRequest{Model: req.Model, Messages: messages})
}

func (r *Runner) prepareSession(req ChatRequest) (*db.ChatSession, []llm.Message, error) {
	if req.SessionID == 0 {
		session, err := db.CreateChatSession(r.DB, req.Prompt)
		return session, nil, err
	}
	session, err := db.GetChatSession(r.DB, req.SessionID)
	if err != nil {
		return nil, nil, err
	}
	historyLimit := req.HistoryLimit
	if historyLimit == 0 {
		historyLimit = db.DefaultHistoryLimit
	}
	recent, err := db.ListRecentChatMessages(r.DB, session.ID, historyLimit)
	if err != nil {
		return nil, nil, err
	}
	history := make([]llm.Message, 0, len(recent))
	for _, message := range recent {
		if message.Role == "" || message.Content == "" {
			continue
		}
		history = append(history, llm.Message{Role: message.Role, Content: message.Content})
	}
	return session, history, nil
}

func ParseToolCalls(content string) []ToolCall {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	var envelope struct {
		ToolCalls []ToolCall `json:"tool_calls"`
	}
	if err := json.Unmarshal([]byte(content), &envelope); err != nil {
		return nil
	}
	return envelope.ToolCalls
}

func NewDefaultRunner(database *gorm.DB) *Runner {
	return NewRunner(database, tools.DefaultRegistry())
}

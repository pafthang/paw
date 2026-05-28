package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pafthang/paw/internal/contextpack"
	"github.com/pafthang/paw/internal/db"
	"github.com/pafthang/paw/internal/llm"
	"github.com/pafthang/paw/internal/skills"
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
	SessionID           uint      `json:"session_id,omitempty"`
	Prompt              string    `json:"prompt"`
	Model               string    `json:"model,omitempty"`
	HistoryLimit        int       `json:"history_limit,omitempty"`
	MaxContextChars     int       `json:"max_context_chars,omitempty"`
	SystemPrompt        string    `json:"system_prompt,omitempty"`
	Skill               string    `json:"skill,omitempty"`
	MaxIterations       int       `json:"max_iterations,omitempty"`
	Workspace           string    `json:"workspace,omitempty"`
	AllowShell          bool      `json:"allow_shell,omitempty"`
	AllowShellDangerous bool      `json:"allow_shell_dangerous,omitempty"`
	OnTool              ToolHook  `json:"-"`
	OnAgent             AgentHook `json:"-"`
}

type ChatResponse struct {
	SessionID       uint             `json:"session_id"`
	ModelResponse   llm.ChatResponse `json:"model_response"`
	FinalResponse   llm.ChatResponse `json:"final_response,omitempty"`
	ToolCalls       []ToolCall       `json:"tool_calls,omitempty"`
	ToolRunResponse RunResponse      `json:"tool_run_response,omitempty"`
	UsedTools       bool             `json:"used_tools"`
	Iterations      int              `json:"iterations,omitempty"`
	Steps           []ChatStep       `json:"steps,omitempty"`
	Skill           string           `json:"skill,omitempty"`
}

type ChatStep struct {
	Iteration       int              `json:"iteration"`
	ModelResponse   llm.ChatResponse `json:"model_response"`
	ToolCalls       []ToolCall       `json:"tool_calls,omitempty"`
	ToolRunResponse RunResponse      `json:"tool_run_response,omitempty"`
	ParseError      string           `json:"parse_error,omitempty"`
}

type AgentHook func(event AgentEvent)

type AgentEvent struct {
	Type        string           `json:"type"`
	Iteration   int              `json:"iteration,omitempty"`
	ModelResult llm.ChatResponse `json:"model_response,omitempty"`
	ToolCalls   []ToolCall       `json:"tool_calls,omitempty"`
	RunResponse RunResponse      `json:"tool_run_response,omitempty"`
	Error       string           `json:"error,omitempty"`
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
	systemPrompt := req.SystemPrompt
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = ToolCallSystemPrompt
	}
	if strings.TrimSpace(req.Skill) != "" {
		skill, err := skills.LoadByName(req.Skill)
		if err != nil {
			return ChatResponse{}, err
		}
		_, _ = db.CreateAuditEvent(r.DB, db.AuditEvent{
			SessionID: session.ID,
			Type:      "agent.skill",
			ToolName:  strings.TrimSpace(req.Skill),
		})
		if strings.TrimSpace(skill.Prompts.System) != "" {
			systemPrompt = strings.TrimSpace(skill.Prompts.System) + "\n\n" + strings.TrimSpace(systemPrompt)
		}
	}
	maxIterations := req.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 4
	}

	conversation := []llm.Message{{Role: "user", Content: prompt}}
	steps := make([]ChatStep, 0, maxIterations)
	var firstModel llm.ChatResponse
	var finalModel llm.ChatResponse
	var firstCalls []ToolCall
	var firstRun RunResponse
	usedTools := false
	for iteration := 1; iteration <= maxIterations; iteration++ {
		emitAgentEvent(req.OnAgent, AgentEvent{Type: "iteration.started", Iteration: iteration})
		activeSystem := systemPrompt
		if usedTools {
			activeSystem = FinalAnswerSystemPrompt
		}
		packed := contextpack.Pack(activeSystem, history, conversation, req.MaxContextChars)
		modelResp, err := client.Chat(ctx, llm.ChatRequest{Model: req.Model, Messages: packed})
		if err != nil {
			emitAgentEvent(req.OnAgent, AgentEvent{Type: "iteration.error", Iteration: iteration, Error: err.Error()})
			return ChatResponse{}, err
		}
		emitAgentEvent(req.OnAgent, AgentEvent{Type: "iteration.model_result", Iteration: iteration, ModelResult: modelResp})
		if iteration == 1 {
			firstModel = modelResp
		}
		calls, parseErr := ParseToolCallsEnvelope(modelResp.Content)
		step := ChatStep{Iteration: iteration, ModelResponse: modelResp}
		if parseErr != nil {
			step.ParseError = parseErr.Error()
		}
		if len(calls) == 0 {
			finalModel = modelResp
			steps = append(steps, step)
			emitAgentEvent(req.OnAgent, AgentEvent{Type: "iteration.finished", Iteration: iteration})
			break
		}
		usedTools = true
		step.ToolCalls = calls
		emitAgentEvent(req.OnAgent, AgentEvent{Type: "tools", Iteration: iteration, ToolCalls: calls})
		runResp, err := r.Run(ctx, RunRequest{
			SessionID:           session.ID,
			ToolCalls:           calls,
			Workspace:           req.Workspace,
			AllowShell:          req.AllowShell,
			AllowShellDangerous: req.AllowShellDangerous,
			Iteration:           iteration,
			OnTool:              req.OnTool,
		})
		if err != nil {
			emitAgentEvent(req.OnAgent, AgentEvent{Type: "iteration.error", Iteration: iteration, Error: err.Error()})
			return ChatResponse{}, err
		}
		step.ToolRunResponse = runResp
		emitAgentEvent(req.OnAgent, AgentEvent{Type: "tools.result", Iteration: iteration, RunResponse: runResp})
		steps = append(steps, step)
		if iteration == 1 {
			firstCalls = calls
			firstRun = runResp
		}
		toolJSON, err := json.MarshalIndent(runResp, "", "  ")
		if err != nil {
			return ChatResponse{}, err
		}
		conversation = append(conversation,
			llm.Message{Role: "assistant", Content: modelResp.Content},
			llm.Message{Role: "user", Content: "Tool results:\n" + string(toolJSON) + "\n\nContinue solving my request."},
		)
		if iteration == maxIterations {
			finalModel = llm.ChatResponse{Model: modelResp.Model, Content: fmt.Sprintf("agent stopped: max iterations (%d) reached while tools still requested", maxIterations)}
			emitAgentEvent(req.OnAgent, AgentEvent{Type: "iteration.finished", Iteration: iteration})
		}
	}

	if _, err := db.AddChatMessage(r.DB, session.ID, "user", prompt, req.Model); err != nil {
		return ChatResponse{}, err
	}
	if usedTools {
		if _, err := db.AddChatMessage(r.DB, session.ID, "assistant", finalModel.Content, finalModel.Model); err != nil {
			return ChatResponse{}, err
		}
	} else {
		if _, err := db.AddChatMessage(r.DB, session.ID, "assistant", firstModel.Content, firstModel.Model); err != nil {
			return ChatResponse{}, err
		}
	}
	return ChatResponse{
		SessionID:       session.ID,
		ModelResponse:   firstModel,
		FinalResponse:   finalModel,
		ToolCalls:       firstCalls,
		ToolRunResponse: firstRun,
		UsedTools:       usedTools,
		Iterations:      len(steps),
		Steps:           steps,
		Skill:           strings.TrimSpace(req.Skill),
	}, nil
}

func emitAgentEvent(hook AgentHook, event AgentEvent) {
	if hook != nil {
		hook(event)
	}
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
	calls, _ := ParseToolCallsEnvelope(content)
	return calls
}

func ParseToolCallsEnvelope(content string) ([]ToolCall, error) {
	blob, ok := extractFirstJSONObject(content)
	if !ok {
		return nil, nil
	}
	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(blob, &asMap); err != nil {
		return nil, err
	}
	var callsRaw json.RawMessage
	if v, ok := asMap["tool_calls"]; ok {
		callsRaw = v
	} else if v, ok := asMap["tools"]; ok {
		callsRaw = v
	} else {
		return nil, nil
	}
	var calls []ToolCall
	if err := json.Unmarshal(callsRaw, &calls); err != nil {
		return nil, err
	}
	if len(calls) == 0 {
		return nil, nil
	}
	out := make([]ToolCall, 0, len(calls))
	for _, call := range calls {
		call.Name = strings.TrimSpace(call.Name)
		if call.Name == "" || len(call.Input) == 0 || !json.Valid(call.Input) {
			continue
		}
		out = append(out, call)
	}
	return out, nil
}

func extractFirstJSONObject(content string) ([]byte, bool) {
	s := strings.TrimSpace(content)
	if s == "" {
		return nil, false
	}
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	start := strings.IndexByte(s, '{')
	if start < 0 {
		return nil, false
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if ch == '\\' {
				escape = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return []byte(s[start : i+1]), true
			}
		}
	}
	return nil, false
}

func NewDefaultRunner(database *gorm.DB) *Runner {
	return NewRunner(database, tools.DefaultRegistry())
}

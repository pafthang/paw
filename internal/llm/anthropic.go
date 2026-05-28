package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const anthropicDefaultBaseURL = "https://api.anthropic.com"
const anthropicVersionHeader = "2023-06-01"

type AnthropicClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		baseURL: anthropicDefaultBaseURL,
		apiKey:  strings.TrimSpace(apiKey),
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *AnthropicClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return ChatResponse{}, errors.New("anthropic api key is missing (set anthropic_api_key)")
	}

	system, messages := anthropicConvertMessages(req.Messages)
	if len(messages) == 0 {
		return ChatResponse{}, fmt.Errorf("anthropic chat requires at least one non-system message")
	}

	payload := struct {
		Model     string                `json:"model"`
		MaxTokens int                   `json:"max_tokens"`
		System    string                `json:"system,omitempty"`
		Messages  []anthropicReqMessage `json:"messages"`
	}{
		Model:     req.Model,
		MaxTokens: 1024,
		System:    system,
		Messages:  messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.baseURL, "/")+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersionHeader)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return ChatResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		msg := strings.TrimSpace(string(b))
		if msg != "" {
			return ChatResponse{}, fmt.Errorf("anthropic chat failed: %s: %s", resp.Status, msg)
		}
		return ChatResponse{}, fmt.Errorf("anthropic chat failed: %s", resp.Status)
	}

	var out struct {
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ChatResponse{}, err
	}
	var sb strings.Builder
	for _, block := range out.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	content := sb.String()
	if strings.TrimSpace(content) == "" {
		return ChatResponse{}, fmt.Errorf("anthropic chat returned empty content")
	}
	return ChatResponse{Model: out.Model, Content: content}, nil
}

type anthropicReqMessage struct {
	Role    string                 `json:"role"`
	Content []anthropicTextContent `json:"content"`
}

type anthropicTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func anthropicConvertMessages(in []Message) (system string, out []anthropicReqMessage) {
	systemParts := make([]string, 0, 2)
	out = make([]anthropicReqMessage, 0, len(in))
	for _, msg := range in {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		content := msg.Content
		if strings.TrimSpace(content) == "" {
			continue
		}
		switch role {
		case "system":
			systemParts = append(systemParts, content)
			continue
		case "assistant", "user":
		default:
			role = "user"
		}
		out = append(out, anthropicReqMessage{
			Role: role,
			Content: []anthropicTextContent{
				{Type: "text", Text: content},
			},
		})
	}
	return strings.Join(systemParts, "\n\n"), out
}

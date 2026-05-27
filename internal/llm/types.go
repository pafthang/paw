package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model,omitempty"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Model   string `json:"model,omitempty"`
	Content string `json:"content"`
}

type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

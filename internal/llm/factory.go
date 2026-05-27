package llm

import (
	"fmt"
	"strings"

	"github.com/pafthang/paw/internal/config"
)

func NewClient(settings config.Settings) (Client, error) {
	backend := strings.TrimSpace(strings.ToLower(settings.AgentBackend))
	switch backend {
	case "", "ollama":
		return NewOllamaClient(settings.OllamaHost), nil
	case "openai_compatible", "openai-compatible", "openai":
		baseURL := settings.OpenAICompatibleBaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return NewOpenAICompatibleClient(baseURL, settings.OpenAIAPIKey), nil
	default:
		return nil, fmt.Errorf("unsupported agent_backend %q; supported: ollama, openai_compatible", settings.AgentBackend)
	}
}

func DefaultModel(settings config.Settings) string {
	if strings.TrimSpace(settings.Model) != "" {
		return strings.TrimSpace(settings.Model)
	}
	switch strings.ToLower(strings.TrimSpace(settings.AgentBackend)) {
	case "openai", "openai_compatible", "openai-compatible":
		return "gpt-4o-mini"
	default:
		return "qwen2.5:7b"
	}
}

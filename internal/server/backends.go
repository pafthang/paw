package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) handleListBackends(c echo.Context) error {
	backends := []map[string]any{
		{
			"name":               "ollama",
			"displayName":        "Ollama",
			"available":          s.settings.OllamaHost != "",
			"capabilities":       []string{"chat", "local"},
			"builtinTools":       []string{},
			"requiredKeys":       []string{},
			"supportedProviders": []string{"ollama"},
			"installHint": map[string]any{
				"external_cmd": "ollama",
				"docs_url":     "https://ollama.com/download",
			},
			"beta": false,
		},
		{
			"name":               "openai-compatible",
			"displayName":        "OpenAI Compatible",
			"available":          s.settings.OpenAICompatibleBaseURL != "",
			"capabilities":       []string{"chat"},
			"builtinTools":       []string{},
			"requiredKeys":       []string{"openai_api_key"},
			"supportedProviders": []string{"openai-compatible"},
			"installHint": map[string]any{
				"docs_url": "https://platform.openai.com/docs/api-reference",
			},
			"beta": false,
		},
		{
			"name":               "openai",
			"displayName":        "OpenAI",
			"available":          s.settings.OpenAIAPIKey != "",
			"capabilities":       []string{"chat"},
			"builtinTools":       []string{},
			"requiredKeys":       []string{"openai_api_key"},
			"supportedProviders": []string{"openai"},
			"installHint": map[string]any{
				"docs_url": "https://platform.openai.com/docs",
			},
			"beta": false,
		},
		{
			"name":               "anthropic",
			"displayName":        "Anthropic",
			"available":          s.settings.AnthropicAPIKey != "",
			"capabilities":       []string{"chat"},
			"builtinTools":       []string{},
			"requiredKeys":       []string{"anthropic_api_key"},
			"supportedProviders": []string{"anthropic"},
			"installHint": map[string]any{
				"docs_url": "https://docs.anthropic.com/claude/reference/messages_post",
			},
			"beta": false,
		},
	}
	return c.JSON(http.StatusOK, backends)
}

func (s *Server) handleFetchOllamaModels(c echo.Context) error {
	// Minimal compatible endpoint for the settings UI. The Go core does not
	// keep an Ollama model catalog cache yet, so return the configured model.
	model := s.settings.Model
	if model == "" {
		model = "qwen2.5:7b"
	}
	return c.JSON(http.StatusOK, []string{model})
}

func (s *Server) handleVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"version":       "go-core",
		"python":        "not-used",
		"agent_backend": s.settings.AgentBackend,
		"timestamp":     time.Now().UTC().Format(time.RFC3339Nano),
	})
}

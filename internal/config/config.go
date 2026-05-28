package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const appDirName = ".pocketpaw"

// Settings intentionally mirrors a small, stable subset of the existing
// Python config.json so the Go core can run beside the current implementation.
type Settings struct {
	WebHost                 string   `json:"web_host"`
	WebPort                 int      `json:"web_port"`
	AgentBackend            string   `json:"agent_backend"`
	Model                   string   `json:"model,omitempty"`
	OllamaHost              string   `json:"ollama_host,omitempty"`
	OpenAICompatibleBaseURL string   `json:"openai_compatible_base_url,omitempty"`
	OpenAIAPIKey            string   `json:"openai_api_key,omitempty"`
	AnthropicAPIKey         string   `json:"anthropic_api_key,omitempty"`
	TelegramBotToken        string   `json:"telegram_bot_token,omitempty"`
	AllowedUserID           int64    `json:"allowed_user_id,omitempty"`
	HealthCheckOnStartup    bool     `json:"health_check_on_startup"`
	CORSAllowedOrigins      []string `json:"api_cors_allowed_origins,omitempty"`
}

func DefaultSettings() Settings {
	return Settings{
		WebHost:                 "127.0.0.1",
		WebPort:                 8888,
		AgentBackend:            "ollama",
		Model:                   "qwen2.5:7b",
		OllamaHost:              "http://127.0.0.1:11434",
		OpenAICompatibleBaseURL: "https://api.openai.com/v1",
		HealthCheckOnStartup:    true,
	}
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, appDirName), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func DBPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "paw.db"), nil
}

func AccessTokenPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "access_token"), nil
}

func EnsureDir() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o700)
}

func Load() (Settings, error) {
	settings := DefaultSettings()
	path, err := Path()
	if err != nil {
		return settings, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return settings, err
	}
	if len(data) == 0 {
		return settings, nil
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, fmt.Errorf("read config %s: %w", path, err)
	}
	if settings.WebHost == "" {
		settings.WebHost = "127.0.0.1"
	}
	if settings.WebPort == 0 {
		settings.WebPort = 8888
	}
	if settings.AgentBackend == "" {
		settings.AgentBackend = "ollama"
	}
	if settings.Model == "" {
		settings.Model = "qwen2.5:7b"
	}
	if settings.OllamaHost == "" {
		settings.OllamaHost = "http://127.0.0.1:11434"
	}
	if settings.OpenAICompatibleBaseURL == "" {
		settings.OpenAICompatibleBaseURL = "https://api.openai.com/v1"
	}
	return settings, nil
}

func Save(settings Settings) error {
	if err := EnsureDir(); err != nil {
		return err
	}
	path, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

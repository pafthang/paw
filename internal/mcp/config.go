package mcp

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/pafthang/paw/internal/config"
)

type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type Config struct {
	Servers map[string]ServerConfig `json:"servers"`
}

func LoadConfig() (Config, error) {
	path, err := config.MCPPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{Servers: map[string]ServerConfig{}}, nil
	}
	if err != nil {
		return Config{}, err
	}
	if len(data) == 0 {
		return Config{Servers: map[string]ServerConfig{}}, nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Servers == nil {
		cfg.Servers = map[string]ServerConfig{}
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	if err := config.EnsureDir(); err != nil {
		return err
	}
	path, err := config.MCPPath()
	if err != nil {
		return err
	}
	if cfg.Servers == nil {
		cfg.Servers = map[string]ServerConfig{}
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func ListServerNames(cfg Config) []string {
	names := make([]string, 0, len(cfg.Servers))
	for name := range cfg.Servers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name is required")
	}
	if strings.Contains(name, " ") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return errors.New("invalid name")
	}
	return nil
}

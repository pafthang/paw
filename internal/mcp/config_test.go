package mcp

import (
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Servers == nil {
		t.Fatalf("servers nil")
	}
	cfg.Servers["demo"] = ServerConfig{Command: "echo", Args: []string{"hi"}}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	cfg2, err := LoadConfig()
	if err != nil {
		t.Fatalf("load2: %v", err)
	}
	if cfg2.Servers["demo"].Command != "echo" {
		t.Fatalf("cfg2=%#v", cfg2)
	}
}

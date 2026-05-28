package presets

import (
	"testing"
)

func TestBuildMCPServerConfig_FilesystemRequiresWorkspace(t *testing.T) {
	if _, err := BuildMCPServerConfig("filesystem", ""); err == nil {
		t.Fatalf("expected error")
	}
	cfg, err := BuildMCPServerConfig("filesystem", ".")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cfg.Command == "" || len(cfg.Args) == 0 {
		t.Fatalf("cfg=%#v", cfg)
	}
}

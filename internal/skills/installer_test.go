package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAndUninstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "skill.yaml"), []byte(`
name: demo
description: demo skill
version: 0.0.1
prompts:
  system: hi
`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	s, err := InstallFromDir(src, InstallOptions{})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if s.Name != "demo" {
		t.Fatalf("skill=%#v", s)
	}

	// re-install without force should fail
	if _, err := InstallFromDir(src, InstallOptions{}); err == nil {
		t.Fatalf("expected error")
	}

	// force overwrite
	if _, err := InstallFromDir(src, InstallOptions{Force: true}); err != nil {
		t.Fatalf("force install: %v", err)
	}

	// uninstall requires --yes
	if err := Uninstall("demo", UninstallOptions{Yes: false}); err == nil {
		t.Fatalf("expected error")
	}
	if err := Uninstall("demo", UninstallOptions{Yes: true}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
}

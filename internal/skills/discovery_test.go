package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAll_ReportsInvalidSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root, err := SkillsRoot()
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	badDir := filepath.Join(root, "bad")
	if err := os.MkdirAll(badDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(badDir, "skill.yaml"), []byte("name:"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	report, err := LoadAll()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(report.Errors) == 0 {
		t.Fatalf("expected errors, got %#v", report)
	}
}

package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_ValidSkill(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skill.yaml")
	if err := os.WriteFile(path, []byte(`
name: go-reviewer
description: Helps review Go code and run tests.
version: 0.1.0
prompts:
  system: |
    You are a Go code reviewer.
commands:
  - name: test
    description: Run Go tests
    tool: shell.run
    input:
      command: go test ./...
      timeout_seconds: 120
`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	s, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if s.Name != "go-reviewer" || s.Version != "0.1.0" {
		t.Fatalf("skill=%#v", s)
	}
	if len(s.Commands) != 1 || s.Commands[0].Tool != "shell.run" {
		t.Fatalf("commands=%#v", s.Commands)
	}
}

func TestLoadFromFile_InvalidSkillMissingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skill.yaml")
	_ = os.WriteFile(path, []byte(`description: x
version: 0.1.0
prompts:
  system: hi
`), 0o600)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Fatalf("expected error")
	}
}

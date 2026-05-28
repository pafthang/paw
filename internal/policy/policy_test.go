package policy

import "testing"

func TestEnsurePathWithinWorkspace_DeniesTraversal(t *testing.T) {
	ws := t.TempDir()
	if err := EnsurePathWithinWorkspace(ws, "../outside.txt"); err == nil {
		t.Fatalf("expected denial")
	}
	if err := EnsurePathWithinWorkspace(ws, "/etc/passwd"); err == nil {
		t.Fatalf("expected denial")
	}
	if err := EnsurePathWithinWorkspace(ws, "~/.ssh/id_rsa"); err == nil {
		t.Fatalf("expected denial")
	}
}

func TestEnsurePathWithinWorkspace_AllowsInside(t *testing.T) {
	ws := t.TempDir()
	if err := EnsurePathWithinWorkspace(ws, "README.md"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

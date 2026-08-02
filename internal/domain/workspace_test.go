package domain

import "testing"

func TestNewWorkspace(t *testing.T) {
	ws, err := NewWorkspace("truelogics")
	if err != nil {
		t.Fatalf("NewWorkspace: unexpected error: %v", err)
	}
	if ws.ID == "" {
		t.Fatal("NewWorkspace: expected a non-empty ID")
	}
	if ws.Name != "truelogics" {
		t.Fatalf("NewWorkspace: Name = %q, want %q", ws.Name, "truelogics")
	}
	if ws.CreatedAt.IsZero() {
		t.Fatal("NewWorkspace: expected CreatedAt to be set")
	}
}

func TestNewWorkspaceRejectsEmptyName(t *testing.T) {
	for _, name := range []string{"", "   "} {
		if _, err := NewWorkspace(name); err == nil {
			t.Fatalf("NewWorkspace(%q): expected error, got nil", name)
		}
	}
}

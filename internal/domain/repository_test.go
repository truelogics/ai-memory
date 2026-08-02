package domain

import "testing"

func TestNewRepository(t *testing.T) {
	repo, err := NewRepository("ws-1", "ai-memory", "/repos/ai-memory")
	if err != nil {
		t.Fatalf("NewRepository: unexpected error: %v", err)
	}
	if repo.ID == "" {
		t.Fatal("NewRepository: expected a non-empty ID")
	}
	if repo.WorkspaceID != "ws-1" || repo.Name != "ai-memory" || repo.LocalPath != "/repos/ai-memory" {
		t.Fatalf("NewRepository: unexpected fields: %+v", repo)
	}
}

func TestNewRepositoryValidation(t *testing.T) {
	cases := []struct {
		workspaceID, name, path string
	}{
		{"", "ai-memory", "/repos/ai-memory"},
		{"ws-1", "", "/repos/ai-memory"},
		{"ws-1", "ai-memory", ""},
	}
	for _, c := range cases {
		if _, err := NewRepository(c.workspaceID, c.name, c.path); err == nil {
			t.Fatalf("NewRepository(%q, %q, %q): expected error, got nil", c.workspaceID, c.name, c.path)
		}
	}
}

func TestRepositoryWithRemoteAndMarkIndexed(t *testing.T) {
	repo, err := NewRepository("ws-1", "ai-memory", "/repos/ai-memory")
	if err != nil {
		t.Fatalf("NewRepository: unexpected error: %v", err)
	}
	repo = repo.WithRemote("git@github.com:truelogics/ai-memory.git")
	if repo.RemoteURL != "git@github.com:truelogics/ai-memory.git" {
		t.Fatalf("WithRemote: RemoteURL = %q", repo.RemoteURL)
	}

	indexed := repo.MarkIndexed("abc123", repo.LastIndexedAt)
	if indexed.LastIndexedCommit != "abc123" {
		t.Fatalf("MarkIndexed: LastIndexedCommit = %q", indexed.LastIndexedCommit)
	}
	// Original value must be unmodified (value receiver, not pointer).
	if repo.LastIndexedCommit != "" {
		t.Fatalf("MarkIndexed mutated original: %+v", repo)
	}
}

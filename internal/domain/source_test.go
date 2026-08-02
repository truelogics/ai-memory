package domain

import "testing"

func TestNewSource(t *testing.T) {
	src, err := NewSource("repo-1", SourceTypeFilesystem)
	if err != nil {
		t.Fatalf("NewSource: unexpected error: %v", err)
	}
	if src.ID == "" {
		t.Fatal("NewSource: expected a non-empty ID")
	}
	if src.RepositoryID != "repo-1" || src.Type != SourceTypeFilesystem {
		t.Fatalf("NewSource: unexpected fields: %+v", src)
	}
}

func TestNewSourceValidation(t *testing.T) {
	if _, err := NewSource("", SourceTypeFilesystem); err == nil {
		t.Fatal("NewSource: expected error for empty repository id")
	}
	if _, err := NewSource("repo-1", ""); err == nil {
		t.Fatal("NewSource: expected error for empty type")
	}
}

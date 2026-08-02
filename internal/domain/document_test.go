package domain

import "testing"

func TestNewRawDocument(t *testing.T) {
	raw, err := NewRawDocument("src-1", "README.md", []byte("# Hello"))
	if err != nil {
		t.Fatalf("NewRawDocument: unexpected error: %v", err)
	}
	if raw.Path != "README.md" || raw.SourceID != "src-1" || string(raw.Bytes) != "# Hello" {
		t.Fatalf("NewRawDocument: unexpected fields: %+v", raw)
	}
	if raw.FetchedAt.IsZero() {
		t.Fatal("NewRawDocument: expected FetchedAt to be set")
	}
}

func TestNewRawDocumentValidation(t *testing.T) {
	if _, err := NewRawDocument("", "README.md", nil); err == nil {
		t.Fatal("NewRawDocument: expected error for empty source id")
	}
	if _, err := NewRawDocument("src-1", "", nil); err == nil {
		t.Fatal("NewRawDocument: expected error for empty path")
	}
}

func TestNewCanonicalDocument(t *testing.T) {
	doc, err := NewCanonicalDocument("repo-1", "src-1", "README.md")
	if err != nil {
		t.Fatalf("NewCanonicalDocument: unexpected error: %v", err)
	}
	if doc.ID == "" {
		t.Fatal("NewCanonicalDocument: expected a non-empty ID")
	}
	if doc.Type != DocTypeUnknown {
		t.Fatalf("NewCanonicalDocument: Type = %q, want %q", doc.Type, DocTypeUnknown)
	}
	if doc.Metadata == nil {
		t.Fatal("NewCanonicalDocument: expected non-nil Metadata")
	}
}

func TestNewCanonicalDocumentValidation(t *testing.T) {
	if _, err := NewCanonicalDocument("", "src-1", "README.md"); err == nil {
		t.Fatal("NewCanonicalDocument: expected error for empty repository id")
	}
	if _, err := NewCanonicalDocument("repo-1", "src-1", ""); err == nil {
		t.Fatal("NewCanonicalDocument: expected error for empty path")
	}
}

func TestNewCanonicalDocumentIDStableByRepoAndPath(t *testing.T) {
	a, _ := NewCanonicalDocument("repo-1", "src-1", "README.md")
	b, _ := NewCanonicalDocument("repo-1", "src-2", "README.md")
	if a.ID != b.ID {
		t.Fatalf("document ID changed when only source id changed: %q != %q", a.ID, b.ID)
	}

	c, _ := NewCanonicalDocument("repo-1", "src-1", "ARCHITECTURE.md")
	if a.ID == c.ID {
		t.Fatalf("document ID collided for a different path: %q", a.ID)
	}
}

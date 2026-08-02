package domain

import "testing"

func TestNewChunk(t *testing.T) {
	c, err := NewChunk("doc-1", 0, "Introduction", "Some text.")
	if err != nil {
		t.Fatalf("NewChunk: unexpected error: %v", err)
	}
	if c.ID == "" || c.DocumentID != "doc-1" || c.Index != 0 || c.Heading != "Introduction" {
		t.Fatalf("NewChunk: unexpected fields: %+v", c)
	}
}

func TestNewChunkValidation(t *testing.T) {
	if _, err := NewChunk("", 0, "h", "c"); err == nil {
		t.Fatal("NewChunk: expected error for empty document id")
	}
	if _, err := NewChunk("doc-1", -1, "h", "c"); err == nil {
		t.Fatal("NewChunk: expected error for negative index")
	}
}

func TestNewChunkIDStableAcrossRechunk(t *testing.T) {
	a, _ := NewChunk("doc-1", 2, "h", "content v1")
	b, _ := NewChunk("doc-1", 2, "h", "content v2")
	if a.ID != b.ID {
		t.Fatalf("chunk ID changed when only content changed: %q != %q", a.ID, b.ID)
	}
}

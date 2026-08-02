package domain

import "testing"

func TestNewTag(t *testing.T) {
	tag, err := NewTag("doc-1", "severity", "error")
	if err != nil {
		t.Fatalf("NewTag: unexpected error: %v", err)
	}
	if tag.ID == "" || tag.DocumentID != "doc-1" || tag.Key != "severity" || tag.Value != "error" {
		t.Fatalf("NewTag: unexpected fields: %+v", tag)
	}
}

func TestNewTagValidation(t *testing.T) {
	if _, err := NewTag("", "severity", "error"); err == nil {
		t.Fatal("NewTag: expected error for empty document id")
	}
	if _, err := NewTag("doc-1", "", "error"); err == nil {
		t.Fatal("NewTag: expected error for empty key")
	}
}

func TestNewTagIDIsDeterministic(t *testing.T) {
	a, _ := NewTag("doc-1", "severity", "error")
	b, _ := NewTag("doc-1", "severity", "error")
	if a.ID != b.ID {
		t.Fatalf("NewTag ids not deterministic: %q != %q", a.ID, b.ID)
	}
}

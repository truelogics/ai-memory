package domain

import "testing"

func TestNewIDIsUniqueAndUUIDShaped(t *testing.T) {
	a := newID()
	b := newID()
	if a == b {
		t.Fatalf("newID produced duplicate ids: %q", a)
	}
	if len(a) != 36 {
		t.Fatalf("newID() = %q, want 36 chars (UUID-shaped), got %d", a, len(a))
	}
}

func TestContentIDIsDeterministic(t *testing.T) {
	a := contentID("repo-1", "README.md")
	b := contentID("repo-1", "README.md")
	if a != b {
		t.Fatalf("contentID not deterministic: %q != %q", a, b)
	}
}

func TestContentIDDiffersByInput(t *testing.T) {
	a := contentID("repo-1", "README.md")
	b := contentID("repo-1", "ARCHITECTURE.md")
	c := contentID("repo-2", "README.md")
	if a == b || a == c || b == c {
		t.Fatalf("contentID collided across distinct inputs: %q %q %q", a, b, c)
	}
}

package domain

import "testing"

func TestMetadataGetSet(t *testing.T) {
	m := NewMetadata()
	if _, ok := m.Get("status"); ok {
		t.Fatal("Get on empty Metadata: expected ok=false")
	}
	m.Set("status", "accepted")
	v, ok := m.Get("status")
	if !ok || v != "accepted" {
		t.Fatalf("Get(%q) = (%q, %v), want (%q, true)", "status", v, ok, "accepted")
	}
}

func TestMetadataCloneIsIndependent(t *testing.T) {
	m := NewMetadata()
	m.Set("status", "accepted")
	clone := m.Clone()
	clone.Set("status", "deprecated")
	if v, _ := m.Get("status"); v != "accepted" {
		t.Fatalf("original Metadata mutated via clone: got %q", v)
	}
	if v, _ := clone.Get("status"); v != "deprecated" {
		t.Fatalf("clone.Get(status) = %q, want %q", v, "deprecated")
	}
}

package domain

import "testing"

func TestNewRelationship(t *testing.T) {
	rel, err := NewRelationship("doc-1", "doc-2", RelationshipReferences, RelationshipExplicit)
	if err != nil {
		t.Fatalf("NewRelationship: unexpected error: %v", err)
	}
	if rel.ID == "" || rel.FromDocumentID != "doc-1" || rel.ToDocumentID != "doc-2" {
		t.Fatalf("NewRelationship: unexpected fields: %+v", rel)
	}
}

func TestNewRelationshipRejectsSelfReference(t *testing.T) {
	if _, err := NewRelationship("doc-1", "doc-1", RelationshipRelated, RelationshipInferred); err == nil {
		t.Fatal("NewRelationship: expected error for self-referencing relationship")
	}
}

func TestNewRelationshipRejectsUnknownEnums(t *testing.T) {
	if _, err := NewRelationship("doc-1", "doc-2", "made-up", RelationshipExplicit); err == nil {
		t.Fatal("NewRelationship: expected error for unknown type")
	}
	if _, err := NewRelationship("doc-1", "doc-2", RelationshipRelated, "made-up"); err == nil {
		t.Fatal("NewRelationship: expected error for unknown source")
	}
}

func TestNewRelationshipRequiresBothIDs(t *testing.T) {
	if _, err := NewRelationship("", "doc-2", RelationshipRelated, RelationshipExplicit); err == nil {
		t.Fatal("NewRelationship: expected error for empty from id")
	}
	if _, err := NewRelationship("doc-1", "", RelationshipRelated, RelationshipExplicit); err == nil {
		t.Fatal("NewRelationship: expected error for empty to id")
	}
}

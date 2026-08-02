package domain

import (
	"errors"
	"strings"
)

// RelationshipType is the (open, growing per KNOWLEDGE_MODEL.md §4)
// vocabulary of edge types. v1 implements the subset below.
type RelationshipType string

const (
	RelationshipRelated    RelationshipType = "related"
	RelationshipReferences RelationshipType = "references"
	RelationshipSupersedes RelationshipType = "supersedes"
)

// RelationshipSource records whether an edge was explicitly authored
// (e.g. an ADR's `supersedes:` front-matter) or inferred (e.g. shared
// tags). See DATABASE.md's `relationships.source` column.
type RelationshipSource string

const (
	RelationshipExplicit RelationshipSource = "explicit"
	RelationshipInferred RelationshipSource = "inferred"
)

// Relationship is a directed, typed edge between two Documents in v1 —
// later, potentially a Document and an Entity (see KNOWLEDGE_MODEL.md's
// open question on polymorphic relationship endpoints).
type Relationship struct {
	ID             string
	FromDocumentID string
	ToDocumentID   string
	Type           RelationshipType
	Source         RelationshipSource
}

// NewRelationship validates and constructs a Relationship.
func NewRelationship(fromDocumentID, toDocumentID string, relType RelationshipType, source RelationshipSource) (Relationship, error) {
	fromDocumentID = strings.TrimSpace(fromDocumentID)
	toDocumentID = strings.TrimSpace(toDocumentID)
	if fromDocumentID == "" || toDocumentID == "" {
		return Relationship{}, errors.New("domain: relationship requires both document ids")
	}
	if fromDocumentID == toDocumentID {
		return Relationship{}, errors.New("domain: relationship cannot reference itself")
	}
	switch relType {
	case RelationshipRelated, RelationshipReferences, RelationshipSupersedes:
	default:
		return Relationship{}, errors.New("domain: unknown relationship type")
	}
	switch source {
	case RelationshipExplicit, RelationshipInferred:
	default:
		return Relationship{}, errors.New("domain: unknown relationship source")
	}
	return Relationship{
		ID:             contentID(fromDocumentID, toDocumentID, string(relType)),
		FromDocumentID: fromDocumentID,
		ToDocumentID:   toDocumentID,
		Type:           relType,
		Source:         source,
	}, nil
}

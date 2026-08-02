// Package search implements kernel.Search: ranked full-text query plus
// related-document enrichment, on top of kernel.Storage's raw FTS5
// results. See ARCHITECTURE.md's Search component.
package search

import (
	"context"
	"fmt"

	"github.com/truelogics/ai-memory/internal/domain"
	"github.com/truelogics/ai-memory/internal/kernel"
)

// maxRelated caps how many related documents Search attaches per result,
// so one heavily-tagged document doesn't drown out the actual match.
const maxRelated = 5

// statsTagKeys are the structural Tags internal/parser/markdown attaches
// to nearly every document (heading_count, etc.) — excluded from the
// shared-tag "related" fallback because they'd make almost every document
// "related" to every other one, which isn't useful.
var statsTagKeys = map[string]bool{
	"heading_count":    true,
	"code_block_count": true,
	"link_count":       true,
	"table_count":      true,
}

// Search implements kernel.Search against a kernel.Storage.
type Search struct {
	Storage kernel.Storage
}

var _ kernel.Search = (*Search)(nil)

// New returns a Search backed by storage.
func New(storage kernel.Storage) *Search {
	return &Search{Storage: storage}
}

// Search implements kernel.Search.
func (s *Search) Search(ctx context.Context, query string, opts kernel.SearchOptions) ([]kernel.SearchResult, error) {
	matches, err := s.Storage.SearchChunks(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	results := make([]kernel.SearchResult, 0, len(matches))
	for _, m := range matches {
		related, err := s.relatedDocuments(ctx, m.Document)
		if err != nil {
			return nil, err
		}
		results = append(results, kernel.SearchResult{
			Document: m.Document,
			Chunk:    m.Chunk,
			Score:    m.Score,
			Snippet:  m.Snippet,
			Related:  related,
		})
	}
	return results, nil
}

// relatedDocuments finds other documents connected to doc: explicit
// Relationships first (an ADR's `supersedes:`, etc.); if there are none —
// the common case in v1, since nothing populates Relationships except a
// front-matter reference that resolves — falls back to documents sharing
// a non-structural Tag (e.g. `audience: agent`).
func (s *Search) relatedDocuments(ctx context.Context, doc domain.CanonicalDocument) ([]domain.CanonicalDocument, error) {
	rels, err := s.Storage.ListRelationships(ctx, doc.ID)
	if err != nil {
		return nil, fmt.Errorf("search: related documents for %s: %w", doc.ID, err)
	}

	seen := map[string]bool{doc.ID: true}
	var related []domain.CanonicalDocument
	for _, rel := range rels {
		otherID := rel.ToDocumentID
		if otherID == doc.ID {
			otherID = rel.FromDocumentID
		}
		if seen[otherID] {
			continue
		}
		seen[otherID] = true
		if other, err := s.Storage.GetDocument(ctx, otherID); err == nil {
			related = append(related, other)
		}
	}
	if len(related) > 0 {
		return related, nil
	}

	for _, tag := range doc.Tags {
		if statsTagKeys[tag.Key] {
			continue
		}
		others, err := s.Storage.FindDocumentsByTag(ctx, doc.RepositoryID, tag.Key, tag.Value, doc.ID)
		if err != nil {
			return nil, fmt.Errorf("search: related documents for %s: %w", doc.ID, err)
		}
		for _, other := range others {
			if seen[other.ID] {
				continue
			}
			seen[other.ID] = true
			related = append(related, other)
			if len(related) >= maxRelated {
				return related, nil
			}
		}
	}
	return related, nil
}

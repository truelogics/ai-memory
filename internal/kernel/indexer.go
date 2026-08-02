package kernel

import (
	"context"

	"github.com/truelogics/ai-memory/internal/domain"
)

// IndexResult summarizes one Index or Sync run — the numbers `eng
// index`/`eng sync` print.
type IndexResult struct {
	Scanned   int
	Added     int
	Updated   int
	Unchanged int
	Errors    int
	// Deleted is only ever non-zero from Sync — Index never removes rows
	// for files no longer on disk. See RFC-0003.
	Deleted int
}

// Indexer orchestrates one `eng index`/`eng sync` run for a Repository:
// Collector -> Parser -> Normalizer -> Chunker per item, then writes via
// Storage. Owns incremental-index decisions (skipping unchanged files via
// ContentHash). Must not execute persistence mechanics directly
// (Storage's job) or interpret file formats itself (Parser's job).
type Indexer interface {
	Index(ctx context.Context, repo domain.Repository) (IndexResult, error)
	// Sync incrementally re-indexes repo using git as the source of
	// truth for what changed — RFC-0003/GRAPH.md. Falls back to Index
	// when the underlying Collector doesn't support incremental
	// collection or repo has never been indexed.
	Sync(ctx context.Context, repo domain.Repository) (IndexResult, error)
}

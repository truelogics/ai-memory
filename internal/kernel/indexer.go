package kernel

import (
	"context"

	"github.com/truelogics/ai-memory/internal/domain"
)

// IndexResult summarizes one Index run — the numbers `eng index` prints.
type IndexResult struct {
	Scanned   int
	Added     int
	Updated   int
	Unchanged int
	Errors    int
}

// Indexer orchestrates one `eng index` run for a Repository: Collector ->
// Parser -> Normalizer -> Chunker per item, then writes via Storage. Owns
// incremental-index decisions (skipping unchanged files via ContentHash).
// Must not execute persistence mechanics directly (Storage's job) or
// interpret file formats itself (Parser's job).
type Indexer interface {
	Index(ctx context.Context, repo domain.Repository) (IndexResult, error)
}

package kernel

import (
	"context"

	"github.com/truelogics/ai-memory/internal/domain"
)

// Collector fetches raw bytes from a Repository. v1 has one
// implementation (internal/collector/filesystem) that both enumerates and
// reads local files; a future non-filesystem Collector (GitHub, Slack,
// Notion) implements the same interface without Indexer, Parser,
// Normalizer, Chunker, Storage, or Search changing. See INTERFACES.md.
type Collector interface {
	// Collect returns every RawDocument collectible from repo.
	Collect(ctx context.Context, repo domain.Repository) ([]domain.RawDocument, error)
}

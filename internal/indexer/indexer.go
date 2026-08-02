// Package indexer implements kernel.Indexer: orchestrates Collector ->
// Parser -> Normalizer -> graph.Extract -> Chunker per item, then writes
// through Storage. Owns incremental-index decisions; does not execute
// persistence mechanics itself or interpret file formats. See
// ARCHITECTURE.md and RFC-0003 (relationship extraction).
package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/truelogics/ai-memory/internal/domain"
	"github.com/truelogics/ai-memory/internal/graph"
	"github.com/truelogics/ai-memory/internal/kernel"
)

// Indexer implements kernel.Indexer.
type Indexer struct {
	Collector kernel.Collector
	// Parsers are tried in order via CanParse; the first match parses
	// the RawDocument. v1 has one (markdown), but Indexer doesn't know
	// or care how many there are — a new Parser is just a new entry.
	Parsers    []kernel.Parser
	Normalizer kernel.Normalizer
	Chunker    kernel.Chunker
	Storage    kernel.Storage
	// Resolver extracts Relationships from a document's front-matter
	// references (RFC-0003). Nil skips extraction entirely — kept
	// optional so existing callers/tests that don't care about
	// relationships aren't forced to wire one up.
	Resolver kernel.ReferenceResolver
	// Now is injectable for deterministic tests; defaults to time.Now.
	Now func() time.Time
}

var _ kernel.Indexer = (*Indexer)(nil)

// New wires the pipeline components into an Indexer. resolver may be nil
// to skip relationship extraction (see Resolver's doc comment).
func New(collector kernel.Collector, parsers []kernel.Parser, normalizer kernel.Normalizer, chunker kernel.Chunker, storage kernel.Storage, resolver kernel.ReferenceResolver) *Indexer {
	return &Indexer{
		Collector:  collector,
		Parsers:    parsers,
		Normalizer: normalizer,
		Chunker:    chunker,
		Storage:    storage,
		Resolver:   resolver,
		Now:        time.Now,
	}
}

// Index implements kernel.Indexer: `eng index`'s entire pipeline for one
// Repository.
func (idx *Indexer) Index(ctx context.Context, repo domain.Repository) (kernel.IndexResult, error) {
	var result kernel.IndexResult

	raws, err := idx.Collector.Collect(ctx, repo)
	if err != nil {
		return result, fmt.Errorf("indexer: collect %s: %w", repo.Name, err)
	}
	result.Scanned = len(raws)

	for _, raw := range raws {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		idx.indexOne(ctx, repo, raw, &result)
	}

	if err := idx.updateIndexState(ctx, repo, result); err != nil {
		return result, fmt.Errorf("indexer: update index state for %s: %w", repo.Name, err)
	}
	return result, nil
}

// indexOne runs one RawDocument through Parser -> Normalizer -> (skip if
// unchanged) -> graph.Extract -> Chunker -> Storage, updating result as
// it goes. Errors at any stage count against result.Errors and move on
// to the next file — one bad file must not abort the whole `eng index`
// run.
func (idx *Indexer) indexOne(ctx context.Context, repo domain.Repository, raw domain.RawDocument, result *kernel.IndexResult) {
	parser := idx.pickParser(raw)
	if parser == nil {
		result.Errors++
		return
	}

	doc, err := parser.Parse(ctx, raw)
	if err != nil {
		result.Errors++
		return
	}

	doc, err = idx.Normalizer.Normalize(ctx, doc)
	if err != nil {
		result.Errors++
		return
	}

	existing, found, err := idx.Storage.FindDocumentByPath(ctx, repo.ID, doc.Path)
	if err != nil {
		result.Errors++
		return
	}
	if found && doc.ContentHash != "" && existing.ContentHash == doc.ContentHash {
		result.Unchanged++
		return
	}

	if idx.Resolver != nil {
		rels, err := graph.Extract(ctx, doc, idx.Resolver)
		if err != nil {
			result.Errors++
			return
		}
		doc.Relationships = append(doc.Relationships, rels...)
	}

	chunks, err := idx.Chunker.Chunk(ctx, doc)
	if err != nil {
		result.Errors++
		return
	}

	if err := idx.Storage.PutDocument(ctx, doc); err != nil {
		result.Errors++
		return
	}
	if err := idx.Storage.PutChunks(ctx, doc.ID, chunks); err != nil {
		result.Errors++
		return
	}

	if found {
		result.Updated++
	} else {
		result.Added++
	}
}

func (idx *Indexer) pickParser(raw domain.RawDocument) kernel.Parser {
	for _, p := range idx.Parsers {
		if p.CanParse(raw) {
			return p
		}
	}
	return nil
}

func (idx *Indexer) updateIndexState(ctx context.Context, repo domain.Repository, result kernel.IndexResult) error {
	status := kernel.IndexStatusClean
	if result.Errors > 0 {
		status = kernel.IndexStatusError
	}
	state := kernel.IndexState{
		RepositoryID: repo.ID,
		// v1 doesn't detect deletions, so this is "documents
		// successfully seen this run," not necessarily every document
		// ever indexed for this repo if files were removed on disk.
		DocumentCount:   result.Added + result.Updated + result.Unchanged,
		LastFullIndexAt: idx.now(),
		Status:          status,
	}
	return idx.Storage.PutIndexState(ctx, state)
}

func (idx *Indexer) now() time.Time {
	if idx.Now != nil {
		return idx.Now()
	}
	return time.Now()
}

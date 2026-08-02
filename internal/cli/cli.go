// Package cli implements the four commands Step 7's Definition of Done
// requires — init, index, search, status — wiring the real kernel
// components together. See docs/cli/CLI.md.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/truelogics/ai-memory/internal/chunker"
	fscollector "github.com/truelogics/ai-memory/internal/collector/filesystem"
	"github.com/truelogics/ai-memory/internal/domain"
	"github.com/truelogics/ai-memory/internal/graph"
	"github.com/truelogics/ai-memory/internal/indexer"
	"github.com/truelogics/ai-memory/internal/kernel"
	"github.com/truelogics/ai-memory/internal/normalizer"
	mdparser "github.com/truelogics/ai-memory/internal/parser/markdown"
	"github.com/truelogics/ai-memory/internal/search"
	"github.com/truelogics/ai-memory/internal/storage/sqlite"
)

const (
	dbDirName  = ".eng"
	dbFileName = "memory.db"

	// defaultWorkspaceID stands in for a real `workspaces` table, which
	// v1's schema doesn't have (see DATABASE.md) — one SQLite file is
	// one Workspace in v1, so there's nothing to disambiguate yet.
	defaultWorkspaceID = "default"
)

func dbPath(dir string) string {
	return filepath.Join(dir, dbDirName, dbFileName)
}

// openStore opens the workspace database for dir, requiring it to
// already exist (i.e. `eng init` has run) unless create is true.
func openStore(dir string, create bool) (*sqlite.Store, error) {
	path := dbPath(dir)
	if !create {
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("no workspace at %s — run `eng init` first", dir)
		}
	} else if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", dbDirName, err)
	}
	return sqlite.Open(path)
}

// gitRemote best-effort reads dir's `origin` remote; empty on any error.
func gitRemote(dir string) string {
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// findOrCreateRepository registers dir as a Repository if it isn't
// already, returning whether it already existed.
func findOrCreateRepository(ctx context.Context, store kernel.Storage, dir string) (domain.Repository, bool, error) {
	existing, ok, err := store.FindRepositoryByPath(ctx, dir)
	if err != nil {
		return domain.Repository{}, false, err
	}
	if ok {
		return existing, true, nil
	}
	repo, err := domain.NewRepository(defaultWorkspaceID, filepath.Base(dir), dir)
	if err != nil {
		return domain.Repository{}, false, err
	}
	return repo, false, nil
}

func newIndexer(store kernel.Storage) *indexer.Indexer {
	return indexer.New(
		fscollector.New(),
		[]kernel.Parser{mdparser.New()},
		normalizer.New(),
		chunker.New(chunker.StrategyHeading),
		store,
		graph.NewResolver(store),
	)
}

// Init implements `eng init`: creates .eng/memory.db and registers dir's
// repository.
func Init(ctx context.Context, dir string, out io.Writer) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	store, err := openStore(absDir, true)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	defer store.Close()

	repo, existed, err := findOrCreateRepository(ctx, store, absDir)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	if remote := gitRemote(absDir); remote != "" {
		repo = repo.WithRemote(remote)
	}
	if err := store.PutRepository(ctx, repo); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	fmt.Fprintf(out, "Created workspace at %s\n", dbPath(absDir))
	verb := "Registered"
	if existed {
		verb = "Already registered"
	}
	fmt.Fprintf(out, "%s repository: %s (%s)\n", verb, repo.Name, absDir)
	return nil
}

// Index implements `eng index [path]`: runs the full Collector -> Parser
// -> Normalizer -> Chunker -> Storage pipeline for dir.
func Index(ctx context.Context, dir string, out io.Writer) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}

	store, err := openStore(absDir, false)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}
	defer store.Close()

	repo, _, err := findOrCreateRepository(ctx, store, absDir)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}
	if err := store.PutRepository(ctx, repo); err != nil {
		return fmt.Errorf("index: %w", err)
	}

	result, err := newIndexer(store).Index(ctx, repo)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}

	fmt.Fprintf(out, "%s: %d scanned, %d added, %d updated, %d unchanged, %d errors\n",
		repo.Name, result.Scanned, result.Added, result.Updated, result.Unchanged, result.Errors)
	return nil
}

// Search implements `eng search <query>`.
func Search(ctx context.Context, dir, query string, out io.Writer) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	store, err := openStore(absDir, false)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}
	defer store.Close()

	results, err := search.New(store).Search(ctx, query, kernel.SearchOptions{Limit: 10})
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(out, "No matches.")
		return nil
	}
	for i, r := range results {
		fmt.Fprintf(out, "%d. %-40s score %.2f\n", i+1, r.Document.Path, r.Score)
		if r.Snippet != "" {
			fmt.Fprintf(out, "   %q\n", r.Snippet)
		}
		if len(r.Related) > 0 {
			names := make([]string, len(r.Related))
			for j, rel := range r.Related {
				names[j] = rel.Path
			}
			fmt.Fprintf(out, "   related: %s\n", strings.Join(names, ", "))
		}
	}
	return nil
}

// Status implements `eng status [path]`.
func Status(ctx context.Context, dir string, out io.Writer) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}

	store, err := openStore(absDir, false)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	defer store.Close()

	repos, err := store.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if len(repos) == 0 {
		fmt.Fprintln(out, "No repositories registered. Run `eng index .` first.")
		return nil
	}

	fmt.Fprintf(out, "%-20s %6s  %-19s  %s\n", "REPOSITORY", "DOCS", "LAST INDEXED", "STATUS")
	for _, repo := range repos {
		state, ok, err := store.GetIndexState(ctx, repo.ID)
		if err != nil {
			return fmt.Errorf("status: %w", err)
		}
		status, docs, lastIndexed := "not indexed", 0, "-"
		if ok {
			status = string(state.Status)
			docs = state.DocumentCount
			if !state.LastFullIndexAt.IsZero() {
				lastIndexed = state.LastFullIndexAt.Format("2006-01-02 15:04:05")
			}
		}
		fmt.Fprintf(out, "%-20s %6d  %-19s  %s\n", repo.Name, docs, lastIndexed, status)
	}
	return nil
}

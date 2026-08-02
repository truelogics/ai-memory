package retriever

import (
	"context"
	"testing"

	"github.com/truelogics/ai-memory/internal/domain"
	"github.com/truelogics/ai-memory/internal/kernel"
	"github.com/truelogics/ai-memory/internal/search"
	"github.com/truelogics/ai-memory/internal/storage/sqlite"
)

func openTestStore(t *testing.T) kernel.Storage {
	t.Helper()
	store, err := sqlite.Open("file:" + t.Name() + "?mode=memory&cache=private")
	if err != nil {
		t.Fatalf("sqlite.Open: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func putDoc(t *testing.T, ctx context.Context, storage kernel.Storage, repoID, path, content string, docType domain.DocType) domain.CanonicalDocument {
	t.Helper()
	doc, err := domain.NewCanonicalDocument(repoID, repoID, path)
	if err != nil {
		t.Fatalf("NewCanonicalDocument: %v", err)
	}
	doc.Content = content
	doc.Type = docType
	if err := storage.PutDocument(ctx, doc); err != nil {
		t.Fatalf("PutDocument: %v", err)
	}
	chunk, err := domain.NewChunk(doc.ID, 0, "", content)
	if err != nil {
		t.Fatalf("NewChunk: %v", err)
	}
	if err := storage.PutChunks(ctx, doc.ID, []domain.Chunk{chunk}); err != nil {
		t.Fatalf("PutChunks: %v", err)
	}
	return doc
}

func TestRetrieveGroupsByDocType(t *testing.T) {
	ctx := context.Background()
	storage := openTestStore(t)
	repo, _ := domain.NewRepository("ws-1", "ai-memory", "/repos/ai-memory")
	_ = storage.PutRepository(ctx, repo)

	putDoc(t, ctx, storage, repo.ID, "ARCHITECTURE.md", "the authentication pipeline architecture", domain.DocTypeStandard)
	putDoc(t, ctx, storage, repo.ID, "adr-0003.md", "authentication decision: use JWT", domain.DocTypeADR)
	putDoc(t, ctx, storage, repo.ID, "rules/auth.md", "authentication must use JWT tokens", domain.DocTypeRule)

	r := New(search.New(storage))
	bundle, err := r.Retrieve(ctx, "authentication")
	if err != nil {
		t.Fatalf("Retrieve: unexpected error: %v", err)
	}
	if bundle.Task != "authentication" {
		t.Fatalf("bundle.Task = %q, want %q", bundle.Task, "authentication")
	}

	byLabel := map[string]kernel.RetrievalGroup{}
	for _, g := range bundle.Groups {
		byLabel[g.Label] = g
	}

	if len(byLabel["Architecture"].Results) != 1 {
		t.Errorf("Architecture group = %+v, want 1 result", byLabel["Architecture"])
	}
	if len(byLabel["Related ADRs"].Results) != 1 {
		t.Errorf("Related ADRs group = %+v, want 1 result", byLabel["Related ADRs"])
	}
	if len(byLabel["Rules"].Results) != 1 {
		t.Errorf("Rules group = %+v, want 1 result", byLabel["Rules"])
	}

	// RFC-0001 non-goals: no PR/Issue ingestion — these must be present,
	// but empty, not omitted.
	for _, label := range []string{"Related Issues", "Related PRs"} {
		g, ok := byLabel[label]
		if !ok {
			t.Errorf("expected a %q group to be present (even if empty)", label)
			continue
		}
		if len(g.Results) != 0 {
			t.Errorf("%q group = %+v, want empty (nothing ingests these yet)", label, g)
		}
	}
}

func TestRetrieveNoMatches(t *testing.T) {
	ctx := context.Background()
	storage := openTestStore(t)
	repo, _ := domain.NewRepository("ws-1", "ai-memory", "/repos/ai-memory")
	_ = storage.PutRepository(ctx, repo)
	putDoc(t, ctx, storage, repo.ID, "a.md", "totally unrelated content", domain.DocTypeReadme)

	r := New(search.New(storage))
	bundle, err := r.Retrieve(ctx, "zzzznonexistentterm")
	if err != nil {
		t.Fatalf("Retrieve: unexpected error: %v", err)
	}
	for _, g := range bundle.Groups {
		if len(g.Results) != 0 {
			t.Errorf("group %q = %+v, want empty for a query with no matches", g.Label, g)
		}
	}
}

func TestKeywordsDropsStopwordsAndUsesOr(t *testing.T) {
	got := keywords("Review the authentication PR")
	want := "authentication OR pr"
	if got != want {
		t.Fatalf("keywords(...) = %q, want %q", got, want)
	}
}

func TestKeywordsFallsBackToRawTaskWhenAllStopwords(t *testing.T) {
	got := keywords("the a an")
	if got != "the a an" {
		t.Fatalf("keywords(all stopwords) = %q, want the original task unchanged", got)
	}
}

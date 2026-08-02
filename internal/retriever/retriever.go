// Package retriever implements kernel.Retriever: gathers everything
// relevant for a task by searching and grouping results by Knowledge
// Type. See ARCHITECTURE.md's Retriever and RFC-0003/the Step 8
// Definition of Done ("Review authentication PR" -> Architecture, ADRs,
// Rules, ...).
package retriever

import (
	"context"
	"fmt"
	"strings"

	"github.com/truelogics/ai-memory/internal/domain"
	"github.com/truelogics/ai-memory/internal/kernel"
)

// groupOrder fixes the section order every RetrievalBundle uses,
// regardless of which types actually have hits — so "Related Issues" and
// "Related PRs" show up (empty) even though nothing ingests those yet,
// matching CLI.md's existing `eng ask` convention of showing what's
// *not* covered, not just omitting it silently.
var groupOrder = []struct {
	label   string
	docType domain.DocType
}{
	{"Architecture", domain.DocTypeStandard},
	{"Related ADRs", domain.DocTypeADR},
	{"Rules", domain.DocTypeRule},
	{"Related RFCs", domain.DocTypeRFC},
	{"Roadmap", domain.DocTypeRoadmap},
	{"Documentation", domain.DocTypeReadme},
}

// alwaysShownEmpty are sections with no producer yet (RFC-0001's
// non-goals: no PR/Issue ingestion) — shown empty rather than omitted.
var alwaysShownEmpty = []string{"Related Issues", "Related PRs"}

// stopwords are dropped when turning a free-text task ("Review the
// authentication PR") into a search query — plain keyword search, not
// natural-language understanding, so noise words just hurt recall.
var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "review": true, "please": true,
	"for": true, "and": true, "or": true, "in": true, "of": true,
	"to": true, "is": true, "on": true, "about": true, "this": true,
	"that": true, "our": true, "we": true, "it": true,
}

// Retriever implements kernel.Retriever against a kernel.Search.
type Retriever struct {
	Search kernel.Search
}

var _ kernel.Retriever = (*Retriever)(nil)

// New returns a Retriever backed by search.
func New(search kernel.Search) *Retriever {
	return &Retriever{Search: search}
}

// Retrieve implements kernel.Retriever.
func (r *Retriever) Retrieve(ctx context.Context, task string) (kernel.RetrievalBundle, error) {
	results, err := r.Search.Search(ctx, keywords(task), kernel.SearchOptions{Limit: 20})
	if err != nil {
		return kernel.RetrievalBundle{}, fmt.Errorf("retriever: %w", err)
	}

	byType := make(map[domain.DocType][]kernel.SearchResult, len(groupOrder))
	other := []kernel.SearchResult{}
	known := make(map[domain.DocType]bool, len(groupOrder))
	for _, g := range groupOrder {
		known[g.docType] = true
	}
	for _, res := range results {
		if known[res.Document.Type] {
			byType[res.Document.Type] = append(byType[res.Document.Type], res)
			continue
		}
		other = append(other, res)
	}

	groups := make([]kernel.RetrievalGroup, 0, len(groupOrder)+len(alwaysShownEmpty)+1)
	for _, g := range groupOrder {
		groups = append(groups, kernel.RetrievalGroup{Label: g.label, Results: byType[g.docType]})
	}
	for _, label := range alwaysShownEmpty {
		groups = append(groups, kernel.RetrievalGroup{Label: label})
	}
	if len(other) > 0 {
		groups = append(groups, kernel.RetrievalGroup{Label: "Other", Results: other})
	}

	return kernel.RetrievalBundle{Task: task, Groups: groups}, nil
}

// keywords turns a free-text task into an FTS5 query: lowercase, drop
// stopwords and duplicates, OR the rest together — a bareword multi-term
// FTS5 MATCH is an implicit AND, which is too strict for a natural
// sentence like "Review authentication PR" (nothing indexes "PR" at all
// yet, and requiring it would zero out an otherwise-good match on
// "authentication"). Falls back to the raw task if every word was a
// stopword.
func keywords(task string) string {
	words := strings.Fields(strings.ToLower(task))
	seen := make(map[string]bool, len(words))
	var kept []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?#()[]{}\"'")
		if w == "" || stopwords[w] || seen[w] {
			continue
		}
		seen[w] = true
		kept = append(kept, w)
	}
	if len(kept) == 0 {
		return task
	}
	return strings.Join(kept, " OR ")
}

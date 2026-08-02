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

// labelByDocType maps a Knowledge Type to the section label its
// documents get grouped under.
var labelByDocType = map[domain.DocType]string{
	domain.DocTypeStandard: "Architecture",
	domain.DocTypeADR:      "Related ADRs",
	domain.DocTypeRule:     "Rules",
	domain.DocTypeRFC:      "Related RFCs",
	domain.DocTypeRoadmap:  "Roadmap",
	domain.DocTypeReadme:   "Documentation",
}

// alwaysShownEmpty are sections with no producer yet (RFC-0001's
// non-goals: no PR/Issue ingestion) — shown empty rather than omitted.
var alwaysShownEmpty = []string{"Related Issues", "Related PRs"}

// defaultPriority is the section order used when Retriever.Priority is
// unset — matches Step 8's own "Priority example" (Architecture, ADRs,
// ..., README) as closely as this kernel's actual Knowledge Types allow.
// Milestone 7: this used to be the only order; now it's just the default.
var defaultPriority = []string{
	"Architecture", "Related ADRs", "Rules", "Related RFCs", "Roadmap",
	"Documentation", "Related Issues", "Related PRs",
}

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
	// Priority orders the returned groups (Milestone 7). Zero value uses
	// defaultPriority. A custom Priority only needs to name the sections
	// worth moving to the front — anything from defaultPriority it
	// doesn't mention still appears afterward; nothing is silently
	// dropped by a partial list.
	Priority []string
}

var _ kernel.Retriever = (*Retriever)(nil)

// New returns a Retriever backed by search, using defaultPriority.
func New(search kernel.Search) *Retriever {
	return &Retriever{Search: search}
}

// Retrieve implements kernel.Retriever.
func (r *Retriever) Retrieve(ctx context.Context, task string) (kernel.RetrievalBundle, error) {
	results, err := r.Search.Search(ctx, keywords(task), kernel.SearchOptions{Limit: 20})
	if err != nil {
		return kernel.RetrievalBundle{}, fmt.Errorf("retriever: %w", err)
	}

	byLabel := make(map[string][]kernel.SearchResult, len(labelByDocType)+len(alwaysShownEmpty))
	var other []kernel.SearchResult
	for _, res := range results {
		if label, ok := labelByDocType[res.Document.Type]; ok {
			byLabel[label] = append(byLabel[label], res)
			continue
		}
		other = append(other, res)
	}
	for _, label := range alwaysShownEmpty {
		if _, ok := byLabel[label]; !ok {
			byLabel[label] = nil
		}
	}

	order := r.Priority
	if len(order) == 0 {
		order = defaultPriority
	}

	emitted := make(map[string]bool, len(byLabel))
	groups := make([]kernel.RetrievalGroup, 0, len(byLabel)+1)
	emit := func(label string) {
		if emitted[label] {
			return
		}
		emitted[label] = true
		groups = append(groups, kernel.RetrievalGroup{Label: label, Results: byLabel[label]})
	}
	for _, label := range order {
		emit(label)
	}
	// A partial custom Priority reorders what it names; anything from
	// the default set it left out still appears, just afterward.
	for _, label := range defaultPriority {
		emit(label)
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

// Package memory is AI Memory's public Go SDK — the only supported way
// for code outside this module to use the kernel. See RFC-0004 and
// docs/sdk/GO_SDK.md. Deliberately narrow: five methods, four types.
// Everything under internal/ stays free to change without this package's
// shape having to change with it.
package memory

// Repository is a registered repository in a Workspace — a narrower,
// public view of internal/domain.Repository. No LastIndexedCommit: that's
// an implementation detail of incremental sync, not something a consumer
// reads or sets directly.
type Repository struct {
	ID        string
	Name      string
	LocalPath string
	RemoteURL string
}

// IndexResult summarizes one Index or Sync call.
type IndexResult struct {
	Scanned   int
	Added     int
	Updated   int
	Unchanged int
	Deleted   int
	Errors    int
}

// SearchOptions filters a Search call.
type SearchOptions struct {
	RepositoryID string
	Limit        int
}

// SearchResult is one ranked hit.
type SearchResult struct {
	Path    string
	Score   float64
	Snippet string
	Related []string // related file paths
}

// FileContext is one entry within a ContextPackage section.
type FileContext struct {
	Path    string
	Score   float64
	Snippet string
}

// ContextPackage is Memory.Context's return value — the structured
// contract between AI Memory and a programmatic consumer (RFC-0004),
// replacing internal/contextbuilder's flat text for anything other than
// a terminal.
//
// RelatedIssues and RelatedPRs are always present, always empty — no
// Issue/PR ingestion exists (RFC-0001's non-goals, still true). There is
// deliberately no SimilarCode or Risks field: this kernel has never
// indexed source code (only markdown), and risk assessment is a
// reasoning task for a consumer's LLM layer, not a retrieval output — see
// RFC-0004's Proposal for why faking either would be worse than omitting
// them.
type ContextPackage struct {
	Task          string
	RelevantFiles []FileContext
	ADRs          []FileContext
	Rules         []FileContext
	RelatedIssues []FileContext
	RelatedPRs    []FileContext
}

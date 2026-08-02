package kernel

import "context"

// AssembledContext is Context Builder's output — v1: formatted text for
// cmd/eng's stdout. Future: a token-budget-aware prompt for Milestone 3's
// LLM layer. Named AssembledContext, not Context, only to avoid reading
// oddly next to the stdlib "context" package import in this same file —
// INTERFACES.md calls this type "Context".
type AssembledContext struct {
	Task string
	Body string
}

// ContextBuilder packages a RetrievalBundle into whatever a consumer
// needs. Must not decide what's relevant (Retriever's job) or call a
// model itself — generation is a separate, later component that consumes
// ContextBuilder's output.
type ContextBuilder interface {
	Build(ctx context.Context, bundle RetrievalBundle) (AssembledContext, error)
}

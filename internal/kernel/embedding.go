package kernel

import "context"

// Embedding is a dense vector representation of text. Nothing in this
// kernel produces one yet — Milestone 5 (hybrid search) is the first
// consumer, once a concrete EmbeddingProvider exists.
type Embedding []float32

// EmbeddingProvider turns text into Embeddings. Pluggable specifically so
// this kernel never ties itself to one vendor — named candidates (OpenAI,
// Ollama, Voyage, Gemini, FastEmbed) are exactly that, names, not
// implementations. Milestone 4 is the abstraction only: no provider
// ships in this pass, on purpose, so the interface gets designed against
// what embeddings are used *for* (Milestone 5's hybrid search) rather
// than shaped around whichever provider happened to be implemented
// first.
type EmbeddingProvider interface {
	// Embed returns one Embedding per input text, in the same order.
	// Batched (not one-at-a-time) because every real provider bills and
	// rate-limits per request, not per text.
	Embed(ctx context.Context, texts []string) ([]Embedding, error)

	// Dimensions reports the vector length this provider produces, so
	// Storage can size a vector column/index appropriately once
	// Milestone 5 needs one. Different providers (and different models
	// within the same provider) produce different lengths — this is not
	// a kernel-wide constant.
	Dimensions() int
}

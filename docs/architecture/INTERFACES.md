---
doc: INTERFACES
audience: [human, agent]
status: draft
owner: ai-memory
last_reviewed: 2026-08-02
---

# Interfaces

> Companion to [KNOWLEDGE_MODEL.md](KNOWLEDGE_MODEL.md) (Lifecycle) and
> [ARCHITECTURE.md](ARCHITECTURE.md) (pipeline). Names the Go interface at
> each lifecycle stage — `Source → Collector → Parser → Normalizer →
> Chunker → Indexer → Storage → Search → Retriever` — so the seam between
> stages exists before any implementation does. **Design only: no methods,
> no implementations.** Method signatures and concrete types are Sprint 2.

## Why interfaces before implementation

Sprint 2 will write real code against these names. Deciding the boundary now
— what each interface owns, what it deliberately doesn't — means the first
concrete implementation slots into a shape already agreed on, instead of the
shape getting discovered (and re-argued) mid-implementation. It also means
v1 can ship a trivial implementation behind an interface (e.g. `Normalizer`
doing nothing to a markdown Document) without that being a hack — Milestone
2 swaps the implementation, not the callers.

## `Source`

```go
type Source interface{}
```

**Package:** `internal/source` (new)

**Responsibility:** Represents where Documents originate — v1: a registered
Repository (a git repo path on local disk). Owns identity (which Workspace/
Repository this is) and enumeration (what's collectible from it — e.g. which
file paths under this repo are worth looking at).

**Does not own:** Reading bytes (`Collector`) or interpreting content
(`Parser`). A `Source` can be enumerated without fetching anything.

**v1:** One concrete type — a local git repository. Real, not trivial:
`eng add`/`eng index` genuinely need repo enumeration in v1, unlike
`Normalizer` below.

## `Collector`

```go
type Collector interface{}
```

**Package:** `internal/collector` (new)

**Responsibility:** Given something a `Source` enumerated, fetches its raw
bytes. v1: `os.ReadFile` against a local path. Future: an HTTP call against
GitHub's/Slack's/Notion's API. This is the only place actual I/O for
fetching content happens.

**Does not own:** Deciding *what* to fetch (`Source`) or *interpreting* what
was fetched (`Parser`).

**v1:** Real, but thin — a local filesystem read. Still its own interface
so a future non-filesystem `Collector` doesn't require touching `Source` or
`Parser`.

## `Parser`

```go
type Parser interface{}
```

**Package:** `internal/parser` (exists — see ARCHITECTURE.md)

**Responsibility:** Turns raw bytes + path into a structured Document:
front-matter, body, and an inferred doc type. v1: markdown only.

**Does not own:** Chunking, storage, or normalization across source
formats — a `Parser` doesn't know a `Normalizer` exists downstream.

**v1:** Real — this is the first component Sprint 2 actually implements.

## `Normalizer`

```go
type Normalizer interface{}
```

**Package:** `internal/normalizer` (new)

**Responsibility:** Reconciles whatever a `Parser` produced into one
canonical Document shape, regardless of source-format quirks — so a future
Slack-thread `Parser`'s output and today's markdown `Parser`'s output both
come out the other side as the same shape for `Chunker` to consume.

**Does not own:** Format-specific interpretation (that's each `Parser`'s
job) or deciding chunk boundaries (`Chunker`'s job).

**v1:** Trivial — markdown's `Parser` output already *is* the canonical
shape, so v1's `Normalizer` is a pass-through. The interface exists so the
seam is there when a second, non-markdown `Parser` needs it — not because
v1 has real normalization work to do.

## `Chunker`

```go
type Chunker interface{}
```

**Package:** `internal/chunker` (new)

**Responsibility:** Splits a normalized Document into Chunks — the unit
`Search` actually indexes (see DATABASE.md's `document_chunks`). Owns the
chunking strategy: whole-file, per-heading, or fixed-window (an open
question in RFC-0001, still unresolved here — this interface is where that
decision gets implemented once made).

**Does not own:** Persisting chunks (`Indexer`/`Storage`) or ranking them
(`Search`).

**v1:** Real — chunking has to exist for `eng search` to return a snippet
instead of "the whole file matched."

## `Indexer`

```go
type Indexer interface{}
```

**Package:** `internal/indexer` (exists — see ARCHITECTURE.md)

**Responsibility:** Orchestrates one `eng index` run — walks a `Source`,
calls `Collector` → `Parser` → `Normalizer` → `Chunker` per item, and hands
the results to `Storage`. Owns incremental-index decisions (skipping
unchanged files via `content_hash`, per DATABASE.md).

**Does not own:** The actual persistence mechanics (`Storage`) or
interpreting file formats (`Parser`).

**v1:** Real.

## `Storage`

```go
type Storage interface{}
```

**Package:** `internal/storage` (exists — see ARCHITECTURE.md)

**Responsibility:** The only component that touches SQLite. Owns reads and
writes for every table in DATABASE.md (`repositories`, `documents`,
`document_chunks`, `tags`, `relationships`, `index_state`).

**Does not own:** Deciding what to store (`Indexer`) or how to rank what
comes back (`Search`).

**v1:** Real.

## `Search`

```go
type Search interface{}
```

**Package:** `internal/search` (exists — see ARCHITECTURE.md)

**Responsibility:** Ranked full-text query over `Storage` — takes a query
string, returns ranked results with matched snippets and related files.

**Does not own:** Turning a natural-language question into a query, or
assembling multiple results into a labeled bundle (`Retriever`).

**v1:** Real.

## `Retriever`

```go
type Retriever interface{}
```

**Package:** `internal/retriever` (exists — see ARCHITECTURE.md)

**Responsibility:** Takes a natural-language question (`eng ask`), extracts
search terms, calls `Search`, and groups/de-duplicates results into a
labeled bundle (Architecture docs / ADRs / Rules / Related PRs). No
generation, no prose — assembly only.

**Does not own:** Generating an answer. This is the seam Milestone 3's LLM
layer calls into for context, without `Retriever` itself changing.

**v1:** Real, heuristic — see CLI.md's `eng ask` for the current shape of
"real."

## What v1 actually builds real vs. trivial implementations of

| Interface | v1 implementation |
|---|---|
| Source | Real — local git repository |
| Collector | Real, thin — local filesystem read |
| Parser | Real — markdown |
| Normalizer | Trivial — pass-through |
| Chunker | Real |
| Indexer | Real |
| Storage | Real |
| Search | Real |
| Retriever | Real, heuristic |

Only `Normalizer` is trivial in v1. It's still a named interface, not
skipped entirely, because the moment a second `Source`/`Parser` pair exists
(Milestone 2), something has to reconcile their output with markdown's —
and that code belongs behind this seam, not scattered into whichever
`Parser` happens to be second.

## Open questions

- Do `Collector` and `Source` end up merged into one interface in practice,
  since v1's `Collector` is trivial enough (`os.ReadFile`) that splitting it
  from `Source`'s enumeration might be over-abstraction for a single-Source
  system? Left as two per KNOWLEDGE_MODEL.md's Lifecycle, to be revisited
  once method signatures are actually drafted.
- Where do cross-cutting concerns — error handling, logging, context
  cancellation — attach? Not decided; deferred until methods are designed.
- Does `Chunker`'s strategy need to be pluggable per doc type (an ADR chunks
  differently than a README), or is one strategy enough for v1?

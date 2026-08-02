---
doc: ARCHITECTURE
audience: [human, agent]
status: draft
owner: ai-memory
last_reviewed: 2026-08-02
---

# Architecture

> Companion to [RFC-0001](../../rfcs/0001-engineering-memory-kernel.md),
> [RFC-0002](../../rfcs/0002-knowledge-engine.md) (why this pipeline shape),
> [KNOWLEDGE_MODEL.md](KNOWLEDGE_MODEL.md), and
> [INTERFACES.md](INTERFACES.md). This describes the v1 kernel only — no AI,
> no embeddings, no agents.

## Pipeline

This is the one diagram everything else in this repo's docs points back
to — [KNOWLEDGE_MODEL.md](KNOWLEDGE_MODEL.md)'s Lifecycle and
[INTERFACES.md](INTERFACES.md)'s ten interfaces are both this same shape,
described at a different level (Lifecycle: the general, source-agnostic
version; Interfaces: the Go seam per stage). This is the concrete v1
instance of both — same stages, named for what they actually are today.

```
Filesystem
      │
      ▼
  Collector       reads a file's bytes off disk
      │
      ▼
   Parser         turns bytes + path into a structured Document
      │
      ▼
  Normalizer      reconciles Document into one canonical shape (pass-through in v1)
      │
      ▼
  Chunker         splits a Document into Chunks — the unit Search indexes
      │
      ▼
  Indexer         orchestrates the above per file, writes via Storage
      │
      ▼
  Storage         persists records (SQLite)
      │
      ▼
   Search         ranked full-text query over Storage
      │
      ▼
  Retriever       gathers everything relevant for a task into a bundle
      │
      ▼
Context Builder   packages the bundle for a consumer — v1: CLI text. Later: an LLM prompt
```

Every arrow is a local, synchronous function call. Nothing in this diagram
makes a network call, calls a model, or leaves the machine. `Filesystem` is
the pipeline's one open end, not a component with its own interface —
it's what v1's `Source` reads from. `Context Builder` is a real component
(see INTERFACES.md); what it produces (`Context`) is the pipeline's output,
consumed by `cmd/eng` today.

## Components

Full responsibilities for every component below live in
[INTERFACES.md](INTERFACES.md). This section is the shorter, narrative
version for the five that are more than a pass-through in v1.

### Collector (`internal/collector`)

Reads a file's raw bytes off disk, given a path `Indexer` found by walking a
`Repository`. Thin in v1 — a wrapped `os.ReadFile` — but its own seam so a
future non-filesystem `Collector` (an HTTP call) doesn't require touching
`Parser`.

### Parser (`internal/parser`)

Takes the bytes `Collector` fetched and produces a `Document` (see
[DOMAIN_MODEL.md](DOMAIN_MODEL.md)): path, front-matter, body, doc type,
content hash. v1 parses markdown only. Doc type (`adr`, `rule`, `standard`,
`roadmap`, `readme`, …) is inferred from front-matter `doc:` field and path
conventions, falling back to `unknown`.

Responsibility boundary: parsing has no opinion about storage or ranking. It
turns bytes into a typed struct and nothing else — it doesn't know SQLite
exists.

### Normalizer (`internal/normalizer`)

Reconciles a `Parser`'s output into one canonical `Document` shape. v1 is a
pass-through — markdown's `Parser` output already is that shape — kept as
its own step so a second `Parser` (Milestone 2+) has somewhere to reconcile
into, instead of `Chunker` needing to special-case every source format.

### Chunker (`internal/chunker`)

Splits a normalized `Document` into `document_chunks` (see
[DATABASE.md](DATABASE.md)) — the unit `Search` actually indexes, so a query
returns a matched section instead of "the whole file matched."

### Indexer (`internal/indexer`)

Walks a `Repository`, calling `Collector` → `Parser` → `Normalizer` →
`Chunker` per file, then writes the resulting records via `Storage`: the
document row, its chunks, and any tags extracted from front-matter. Also
records `git` metadata available cheaply (last commit, author) without doing
full history analysis in v1. Owns incremental-index decisions (skipping
unchanged files via `content_hash`).

Responsibility boundary: indexing decides *what* gets stored, not *how* it's
queried later. It writes once per `eng index` run; it never reads back its
own output to answer a query.

### Storage (`internal/storage`)

SQLite adapter. Owns the schema ([`DATABASE.md`](DATABASE.md)) and all reads/writes.
Every other component talks to Storage through a narrow interface (e.g.
`PutDocument`, `PutChunks`, `Query`) — no component reaches into SQLite
directly except this one, so swapping the backing store later doesn't
ripple outward.

### Search (`internal/search`)

Takes a query string, runs it against Storage's full-text index (SQLite
FTS5), returns a ranked list: file, score, matched snippet, related files (by
shared tags or same-ADR linkage). This is what `eng search` calls directly.

### Retriever (`internal/retriever`)

Gathers everything relevant for a task — not just a search query. v1's only
task shape is a natural-language question (`eng ask`): extract keywords,
call Search with them, group and de-duplicate the results into a labeled
bundle (e.g. "Architecture docs", "ADRs", "Related PRs" — PRs empty until
Milestone 2 ingests them). No generation, no synthesis of a prose answer —
the value is better *assembly* of what Search already found, not new
intelligence. A broader task shape (e.g. "review PR #123") is where this is
headed, not implemented in v1.

### Context Builder (`internal/contextbuilder`)

Packages a `Retriever` bundle into whatever a consumer needs. v1's only
consumer is a terminal, so this is thin: formatting the bundle as readable
CLI output. This is the seam where Milestone 3's LLM layer plugs in later —
it calls `Context Builder` for a packaged prompt instead of formatting
`Retriever`'s bundle itself, so `Retriever` never has to know or care what
its output gets turned into.

### CLI (`cmd/eng`)

Thin layer translating the seven commands in [`CLI.md`](../cli/CLI.md)
(`init`, `add`, `index`, `search`, `ask`, `status`, `doctor`) into calls
against the components above, plus formatting output for a terminal. No business logic lives here — if the CLI were deleted and
replaced with an HTTP handler tomorrow, none of the components above would
change.

## Non-goals reflected in this design

- No component here calls out to a model or an embedding API — Search,
  Retriever, and Context Builder are pure functions over Storage's results.
- No component assumes a single repo — `Workspace` (see DOMAIN_MODEL.md) can
  span `ai-memory`, `engineering`, `roadmap`, `vision` from day one, even
  though v1 may only be exercised against one repo at a time.
- No component is network-facing. `eng` is a local CLI against a local
  SQLite file.

## Where later milestones attach

This pipeline is deliberately built so Milestones 2–4 are additive, not
rewrites:

- **Milestone 2 (Intelligence):** a `relationships` table and ranking model
  sit between Storage and Search — Search's interface doesn't change, its
  results just get better.
- **Milestone 3 (AI layer):** a new generation component calls Context
  Builder for a packaged prompt and an LLM for generation, sitting *after*
  Context Builder in the pipeline — Retriever and Context Builder don't
  change.
- **Milestone 4 (Engineering OS):** PR/planning/bug intelligence are new
  Parsers (PR, Issue) feeding the same Indexer → Storage → Search path.

If a future milestone requires changing Parser, Indexer, or Storage's core
contract to fit, that's a signal the v1 design needs revisiting — not a
reason to bolt something on sideways.

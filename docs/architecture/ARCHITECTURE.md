---
doc: ARCHITECTURE
audience: [human, agent]
status: draft
owner: ai-memory
last_reviewed: 2026-08-01
---

# Architecture

> Companion to [RFC-0001](../../rfcs/0001-engineering-memory-kernel.md) and
> [KNOWLEDGE_MODEL.md](KNOWLEDGE_MODEL.md). This describes the v1 kernel
> only — no AI, no embeddings, no agents. KNOWLEDGE_MODEL.md's Lifecycle
> section maps its `Collect`/`Normalize` stages onto this pipeline — both are
> effectively no-ops in v1 since the only Source is a local git repo and the
> only format is markdown.

## Pipeline

```
Developer / CLI
      │
      ▼
   Parser         reads files, produces structured Documents
      │
      ▼
  Indexer         turns Documents into searchable records
      │
      ▼
  Storage         persists records (SQLite)
      │
      ▼
   Search         ranked full-text query over Storage
      │
      ▼
  Retriever       assembles a question into a bundle of Search results
      │
      ▼
  CLI output
```

Every arrow is a local, synchronous function call. Nothing in this diagram
makes a network call, calls a model, or leaves the machine.

## Components

### Parser (`internal/parser`)

Reads a file from disk and produces a `Document` (see
[DOMAIN_MODEL.md](DOMAIN_MODEL.md)): path, front-matter, body, doc type,
content hash. v1 parses markdown only. Doc type (`adr`, `rule`, `standard`,
`roadmap`, `readme`, …) is inferred from front-matter `doc:` field and path
conventions, falling back to `unknown`.

Responsibility boundary: parsing has no opinion about storage or ranking. It
turns bytes into a typed struct and nothing else — it doesn't know SQLite
exists.

### Indexer (`internal/indexer`)

Walks a `Repository`, invokes the Parser per file, and turns each `Document`
into the records Storage needs: the document row, its chunks (see Open
questions in RFC-0001 for what a "chunk" is), and any tags extracted from
front-matter. Also records `git` metadata available cheaply (last commit,
author) without doing full history analysis in v1.

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

Takes a natural-language question (`eng ask`), not a search query. In v1
this is deliberately simple: extract keywords, call Search with them, group
and de-duplicate the results into a labeled bundle (e.g. "Architecture docs",
"ADRs", "Related PRs" — PRs empty until Milestone 2 ingests them). No
generation, no synthesis of a prose answer — the value is better *assembly*
of what Search already found, not new intelligence.

This is the seam where Milestone 3's LLM layer plugs in later: it will call
Retriever for context and generate prose on top, without Retriever itself
changing.

### CLI (`cmd/eng`)

Thin layer translating five commands (`init`, `index`, `search`, `ask`,
`status`) into calls against the components above, plus formatting output for
a terminal. No business logic lives here — if the CLI were deleted and
replaced with an HTTP handler tomorrow, none of the components above would
change.

## Non-goals reflected in this design

- No component here calls out to a model or an embedding API — Search and
  Retriever are pure functions over Storage.
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
- **Milestone 3 (AI layer):** a new component calls Retriever for context and
  an LLM for generation, sitting *after* Retriever in the pipeline, not
  inside it.
- **Milestone 4 (Engineering OS):** PR/planning/bug intelligence are new
  Parsers (PR, Issue) feeding the same Indexer → Storage → Search path.

If a future milestone requires changing Parser, Indexer, or Storage's core
contract to fit, that's a signal the v1 design needs revisiting — not a
reason to bolt something on sideways.

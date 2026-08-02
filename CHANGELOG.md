# Changelog

All notable changes to this project are documented here. Format loosely
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **Kernel MVP (Step 7).** `eng init`, `eng index`, `eng search`, and `eng
  status` work end-to-end against real markdown repositories — see
  [`SPRINT_2_REVIEW.md`](SPRINT_2_REVIEW.md) for the full milestone
  breakdown, coverage numbers, and known gaps. New packages:
  - `internal/domain` — Workspace, Repository, Source, Document (Raw +
    Canonical), Chunk, Metadata, Tag, Relationship (100% test coverage).
  - `internal/kernel` — the ten pipeline interfaces with real method
    signatures (previously design-only).
  - `internal/collector/filesystem` — walks a repo for markdown.
  - `internal/parser/markdown` — goldmark-based (headings, code blocks,
    links, tables, front-matter), not a regex.
  - `internal/normalizer` — defaults, dedup, path cleaning.
  - `internal/chunker` — heading / paragraph / fixed-size strategies.
  - `internal/storage/sqlite` — schema, CRUD, FTS5 search, transactions
    (`modernc.org/sqlite`, pure Go).
  - `internal/search` — ranking plus related-document lookup (explicit
    Relationships, falling back to shared Tags).
  - `internal/indexer` — orchestrates the full pipeline, incremental via
    content hash.
  - `internal/cli` — testable command implementations `cmd/eng` calls into.
- RFC-0001: Engineering Memory Kernel — scopes v1 to full-text search over
  markdown/ADRs, no AI, embeddings, or agents.
- `docs/architecture/`: KNOWLEDGE_MODEL.md, ARCHITECTURE.md, DOMAIN_MODEL.md,
  DATABASE.md, INTERFACES.md.
- `docs/cli/CLI.md`: seven-command CLI design (`init`, `add`, `index`,
  `search`, `ask`, `status`, `doctor`) — four implemented, three designed
  only (`add`, `ask`, `doctor`).
- RFC-0002: Knowledge Engine — why this pipeline shape.
- LICENSE (Apache 2.0), CONTRIBUTING.md.

### Changed

- `ARCHITECTURE.md` and `INTERFACES.md` updated with real package
  locations and two corrections against what they originally predicted:
  `Normalizer` does real (if small) work, not nothing; `Retriever` and
  `Context Builder` were not implemented in Step 7 (scope stopped at
  `eng ask` not being required), despite originally being marked "real."
- Repository restructured: `RFC/` → `rfcs/`; root-level design docs moved
  under `docs/architecture/` and `docs/cli/`.
- `VISION.md` and `ROADMAP.md` removed from this repo — product vision now
  lives in [`vision/`](../vision/) and priority tracking in
  [`roadmap/`](../roadmap/), so there's one copy of each instead of two that
  can drift.

# Changelog

All notable changes to this project are documented here. Format loosely
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- RFC-0001: Engineering Memory Kernel — scopes v1 to full-text search over
  markdown/ADRs, no AI, embeddings, or agents.
- `docs/architecture/`: ARCHITECTURE.md, DOMAIN_MODEL.md, DATABASE.md.
- `docs/cli/CLI.md`: seven-command CLI design (`init`, `add`, `index`,
  `search`, `ask`, `status`, `doctor`).
- Go module and `cmd/eng` skeleton — `eng version` only, no other logic yet.
- LICENSE (Apache 2.0), CONTRIBUTING.md.

### Changed

- Repository restructured: `RFC/` → `rfcs/`; root-level design docs moved
  under `docs/architecture/` and `docs/cli/`.
- `VISION.md` and `ROADMAP.md` removed from this repo — product vision now
  lives in [`vision/`](../vision/) and priority tracking in
  [`roadmap/`](../roadmap/), so there's one copy of each instead of two that
  can drift.

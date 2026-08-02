---
doc: README
audience: [human, agent]
status: living
owner: ai-memory
last_reviewed: 2026-08-01
---

# AI Memory

## What is this?

The kernel of the [AI Engineering OS](../vision/README.md). Mission: store,
organize, retrieve, and connect engineering knowledge. Everything else in the
OS (context, intelligence, workflows) gets built on top of this, the same way
an OS grows outward from a kernel — it doesn't start as one.

A persistent context layer for agents. It ingests engineering docs, decisions,
and conventions so AI systems can remember, enforce, and assist across sessions.

## Why does it exist?

Agents start fresh every session. Knowledge lives in scattered docs and people's
heads. AI Memory turns `engineering/` (and related repos) into queryable memory
agents can load on demand.

## Who is it for?

- Agents that need durable context (review, codegen, onboarding)
- Engineers building or integrating agent tooling
- Anyone defining how memory is ingested and queried

## Current status

**Planning** — documentation only. No implementation yet.

## Roadmap

See [`ROADMAP.md`](ROADMAP.md). Summary:

| Phase | Focus |
|-------|--------|
| 0 (now) | Docs, RFC-0001, data model |
| 1 | MVP ingest + index + query |
| 2 | Roadmap ingest, agent/MCP integration |
| 3 | Multi-repo, embeddings, observability |

Tracked company-wide in [`roadmap/NOW.md`](../roadmap/NOW.md).

## Contributing

1. Read [`VISION.md`](VISION.md), [`RFC-0001`](RFC/0001-engineering-memory-kernel.md),
   [`ARCHITECTURE.md`](ARCHITECTURE.md), [`DOMAIN_MODEL.md`](DOMAIN_MODEL.md),
   [`DATABASE.md`](DATABASE.md), and [`CLI.md`](CLI.md)
2. Significant design → open an RFC under `RFC/` (start from `0000-template.md`)
3. Org-wide decisions also land in [`engineering/ADR/`](../engineering/ADR/)

Code dirs (`cmd/`, `internal/`, `pkg/`, `tests/`) are reserved — no code yet.

## Related repos

| Repo | Role |
|------|------|
| [`engineering/`](../engineering/) | Source docs & rules consumed by memory |
| [`roadmap/`](../roadmap/) | Company priorities |
| [`vision/`](../vision/) | Company north star |

---
doc: ROADMAP
audience: [human, agent]
status: living
owner: ai-memory
last_reviewed: 2026-07-30
---

# Roadmap

Project phases for AI Memory. Company-wide priorities live in
[`roadmap/NOW.md`](../roadmap/NOW.md).

## Phase 0 — Documentation (now)

- [x] Repo scaffold and vision
- [ ] RFC-0001 — core design and scope
- [ ] Data model draft (`docs/data-model.md`)

## Phase 1 — MVP

- [ ] Ingest markdown from `engineering/` (rules, ADRs, standards)
- [ ] Index by doc type, rule id, and front-matter
- [ ] Query API or CLI: "what rules apply to this path?"
- [ ] Basic examples in `examples/`

## Phase 2 — Expand

- [ ] Ingest `roadmap/` and cross-repo links
- [ ] Agent integration (MCP tool or SDK hook)
- [ ] Incremental updates and staleness detection

## Phase 3 — Scale

- [ ] Multi-tenant / multi-repo
- [ ] Embeddings and semantic retrieval (if needed)
- [ ] Observability: what context was served and when

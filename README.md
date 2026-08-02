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

**Design complete, implementation starting.** RFC-0001, architecture, domain
model, database schema, and CLI design are done. Code so far: a Go module
skeleton (`cmd/eng`) that runs `eng version` and nothing else — indexing,
storage, and search are Sprint 2.

## Roadmap

This repo has no roadmap file of its own — company-wide priority lives in
[`roadmap/NOW.md`](../roadmap/NOW.md) and milestones in
[`roadmap/MILESTONES.md`](../roadmap/MILESTONES.md), so there's exactly one
place to check what's next, not two that can drift out of sync.

## Contributing

1. Read [`RFC-0001`](rfcs/0001-engineering-memory-kernel.md),
   [`RFC-0002`](rfcs/0002-knowledge-engine.md),
   [`KNOWLEDGE_MODEL.md`](docs/architecture/KNOWLEDGE_MODEL.md) (start here —
   what Engineering Knowledge actually is),
   [`ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md),
   [`DOMAIN_MODEL.md`](docs/architecture/DOMAIN_MODEL.md),
   [`DATABASE.md`](docs/architecture/DATABASE.md),
   [`INTERFACES.md`](docs/architecture/INTERFACES.md), and
   [`CLI.md`](docs/cli/CLI.md)
2. Significant design → open an RFC under `rfcs/` (start from `0000-template.md`)
3. Org-wide decisions also land in [`engineering/ADR/`](../engineering/ADR/)

Code dirs (`cmd/`, `internal/`, `pkg/`, `tests/`) are mostly reserved —
`cmd/eng` has a skeleton, the rest have no code yet.

## Related repos

| Repo | Role |
|------|------|
| [`engineering/`](../engineering/) | Source docs & rules consumed by memory |
| [`roadmap/`](../roadmap/) | Company priorities |
| [`vision/`](../vision/) | Company north star |

## Map

```
ai-memory/
├── README.md
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── go.mod
├── rfcs/               ← design proposals (0001: Engineering Memory Kernel)
├── docs/
│   ├── architecture/   ← KNOWLEDGE_MODEL.md, ARCHITECTURE.md, DOMAIN_MODEL.md, DATABASE.md, INTERFACES.md
│   ├── cli/            ← CLI.md
│   └── api/ storage/ search/ sdk/ plugins/ examples/   ← reserved
├── cmd/eng/            ← CLI entrypoint (`eng version` only, so far)
├── internal/           ← reserved — parser, indexer, storage, search
├── pkg/                ← reserved — public libraries
├── examples/           ← reserved — runnable usage examples
├── scripts/            ← reserved — dev/build scripts
└── tests/              ← reserved — integration/e2e tests
```

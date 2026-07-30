---
doc: VISION
audience: [human, agent]
status: placeholder
owner: ai-memory
last_reviewed: 2026-07-30
---

# Vision

## Problem

Agents start fresh every session. Engineering knowledge lives in scattered docs,
chat history, and people's heads. Reviews repeat the same feedback; conventions
aren't enforced consistently.

## Solution

**AI Memory** is a persistent context engine that:

1. **Ingests** — engineering docs, ADRs, rules, roadmap, and related artifacts
2. **Indexes** — structured, searchable memory agents can query by task
3. **Serves** — relevant context to agents at the right time (review, codegen, onboarding)

## Principles

- **Docs are the source of truth** — memory reflects `engineering/`, not replaces it
- **Machine-parseable** — front-matter, rule ids, and stable schemas over prose alone
- **Incremental** — MVP ingests markdown; expand formats and sources later
- **Inspectable** — humans can see what the agent knows and why

## Success (MVP)

An agent working on a PR can load conventions, applicable rules, and related
ADRs without re-reading the entire `engineering/` repo.

## Non-goals (MVP)

- Full RAG over arbitrary codebases
- Real-time sync with every edit
- Replacing human review or decision-making

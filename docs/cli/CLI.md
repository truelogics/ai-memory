---
doc: CLI
audience: [human, agent]
status: draft
owner: ai-memory
last_reviewed: 2026-08-02
---

# CLI

> Companion to [RFC-0001](../../rfcs/0001-engineering-memory-kernel.md),
> [ARCHITECTURE.md](../architecture/ARCHITECTURE.md), and
> [DATABASE.md](../architecture/DATABASE.md).

## Note on scope: 5 commands vs. 7

RFC-0001's CLI section originally listed five commands (`init`, `index`,
`search`, `ask`, `status`). The Week 1 plan's CLI task specified seven,
adding `add` and `doctor`. That was a real discrepancy, not a stylistic one —
RFC-0001 has since been patched to match this document's seven. Keeping the
reasoning here since it's still the source of truth for *why*:

- **`add`** earns its place: without it, a multi-repo Workspace (the whole
  point of indexing `ai-memory` + `engineering` + `roadmap` + `vision`
  together) has no CLI entry point — `init` only bootstraps the workspace
  against the current directory's repo. `add` registers one more Repository
  into an existing Workspace. It does **not** mean "add a single document" —
  `index` already discovers documents by walking a repo, so a
  document-level add would duplicate that.
- **`doctor`** also earns its place: `status` reports whether the index is
  stale (needs re-running `eng index`); `doctor` diagnoses whether the index
  is *broken* (orphaned chunks, dangling foreign keys, a registered repo
  whose `local_path` no longer exists). Different failure class, matches the
  `git fsck` / `brew doctor` convention.

Net: this reconciles to **7 commands**, matching RFC-0001.

## Commands

### `eng init`

**Purpose:** Bootstrap a new Workspace in the current directory. Creates
`.eng/memory.db` and registers the current directory's git repo as the
first Repository.

**Arguments:** none in v1.

**Output:** confirmation and the path to the created database.

**Example:**

```
$ cd ai-memory && eng init
Created workspace at .eng/memory.db
Registered repository: ai-memory (.)
```

### `eng add <path>`

**Purpose:** Register another local repository into the current Workspace,
so `eng index` and `eng search` span it too.

**Arguments:** `path` (required) — path to another git repo, typically a
sibling directory (`../engineering`, `../roadmap`, `../vision`).

**Output:** confirmation and the new repository row summary.

**Example:**

```
$ eng add ../engineering
Registered repository: engineering (../engineering)
Workspace now spans 2 repositories. Run `eng index` to index it.
```

### `eng index`

**Purpose:** Walk every registered repository, parse markdown files, and
populate `documents` / `document_chunks` / `tags` / `relationships` (see
[`../architecture/DATABASE.md`](../architecture/DATABASE.md)). Skips files
whose `content_hash` hasn't changed since the last run.

**Arguments:**
- `--repo <name>` — index only one registered repository
- `--full` — ignore `content_hash`, re-parse everything

**Output:** per-repo summary — files scanned, added, updated, unchanged,
errors.

**Example:**

```
$ eng index
ai-memory:    42 scanned, 3 added, 1 updated, 38 unchanged
engineering:  118 scanned, 0 added, 5 updated, 113 unchanged
Indexed in 0.4s
```

### `eng search "<query>"`

**Purpose:** Ranked full-text search across all indexed chunks in the
Workspace.

**Arguments:**
- `query` (required, positional)
- `--repo <name>` — restrict to one repository
- `--type <doc_type>` — restrict to `adr`, `rule`, `standard`, etc.
- `--limit <n>` — default 10

**Output:** ranked list — file, score, matched snippet, related files.

**Example:**

```
$ eng search "authentication"
1. engineering/ADR/0003-jwt-auth.md          score 0.91
   "...we chose JWT for stateless auth because..."
   related: engineering/ARCHITECTURE.md, ai-memory/rfcs/0001-*.md

2. README.md                                  score 0.42
   "...authentication is handled by..."
```

### `eng ask "<question>"`

**Purpose:** Retriever bundle for a natural-language question — groups
`eng search` results into labeled sections. No LLM, no generated prose (see
[`../architecture/ARCHITECTURE.md`](../architecture/ARCHITECTURE.md)'s Retriever).

**Arguments:** `question` (required, positional)

**Output:** structured bundle, empty sections shown so it's clear what
isn't covered yet (e.g. PRs, until Milestone 2 ingests them).

**Example:**

```
$ eng ask "how does authentication work?"

Architecture docs:
  - engineering/ARCHITECTURE.md

ADRs:
  - engineering/ADR/0003-jwt-auth.md

Rules:
  - engineering/rules/auth-required.yaml

Related PRs:
  (none indexed — PR ingestion is Milestone 2)
```

### `eng status`

**Purpose:** Report index health per registered repository — is it built,
and is it current.

**Arguments:** `--repo <name>` — restrict to one repository

**Output:** table of repos with document count, last indexed commit/time,
and staleness (current `HEAD` vs. `last_indexed_commit`).

**Example:**

```
$ eng status
REPOSITORY    DOCS  LAST INDEXED       STATUS
ai-memory     45    2026-08-02 09:14   clean
engineering   123   2026-08-01 18:02   stale (3 commits behind)
```

### `eng doctor`

**Purpose:** Diagnose problems `status` doesn't catch: orphaned
`document_chunks` with no parent `documents` row, a registered repository
whose `local_path` no longer exists on disk, schema version mismatch.

**Arguments:** none in v1. `--fix` (safe auto-repairs, e.g. pruning orphaned
rows) is deferred — v1 reports only, doesn't mutate.

**Output:** list of issues found, or a clean bill of health.

**Example:**

```
$ eng doctor
✓ 2 repositories registered, both resolve on disk
✓ no orphaned document_chunks
✓ schema version matches (v1)
Workspace healthy.
```

## Conventions

- Exit code `0`: success. `1`: usage/argument error. `2` (doctor only):
  issues found.
- All commands operate on the Workspace found by walking up from the
  current directory to `.eng/` — same convention as `.git`.
- No command in this list makes a network call or reads an API key. If a
  future flag needs one (Milestone 3's AI layer), it's a new command, not a
  flag bolted onto one of these seven.

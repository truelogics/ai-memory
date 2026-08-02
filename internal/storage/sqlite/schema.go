package sqlite

// schema is applied idempotently on every Open — v1's "migration"
// mechanism (see DATABASE.md). A real migration framework is future work
// once the schema needs to evolve after data exists in the field.
const schema = `
CREATE TABLE IF NOT EXISTS repositories (
	id                  TEXT PRIMARY KEY,
	workspace_id        TEXT NOT NULL,
	name                TEXT NOT NULL,
	remote_url          TEXT NOT NULL DEFAULT '',
	local_path          TEXT NOT NULL,
	last_indexed_commit TEXT NOT NULL DEFAULT '',
	last_indexed_at     TIMESTAMP
);

CREATE TABLE IF NOT EXISTS documents (
	id             TEXT PRIMARY KEY,
	repository_id  TEXT NOT NULL REFERENCES repositories(id),
	path           TEXT NOT NULL,
	doc_type       TEXT NOT NULL DEFAULT 'unknown',
	title          TEXT NOT NULL DEFAULT '',
	front_matter   TEXT NOT NULL DEFAULT '{}',
	body           TEXT NOT NULL DEFAULT '',
	content_hash   TEXT NOT NULL DEFAULT '',
	git_author     TEXT NOT NULL DEFAULT '',
	git_updated_at TIMESTAMP,
	indexed_at     TIMESTAMP,
	UNIQUE (repository_id, path)
);

CREATE TABLE IF NOT EXISTS document_chunks (
	id          TEXT PRIMARY KEY,
	document_id TEXT NOT NULL REFERENCES documents(id),
	chunk_index INTEGER NOT NULL,
	heading     TEXT NOT NULL DEFAULT '',
	content     TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_document_chunks_document ON document_chunks(document_id);

CREATE TABLE IF NOT EXISTS tags (
	id          TEXT PRIMARY KEY,
	document_id TEXT NOT NULL REFERENCES documents(id),
	key         TEXT NOT NULL,
	value       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_tags_document ON tags(document_id);

CREATE TABLE IF NOT EXISTS relationships (
	id                TEXT PRIMARY KEY,
	from_document_id  TEXT NOT NULL,
	to_document_id    TEXT NOT NULL,
	relationship_type TEXT NOT NULL,
	source            TEXT NOT NULL,
	created_at        TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_relationships_from ON relationships(from_document_id);
CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_document_id);

CREATE TABLE IF NOT EXISTS index_state (
	repository_id             TEXT PRIMARY KEY REFERENCES repositories(id),
	document_count            INTEGER NOT NULL DEFAULT 0,
	last_full_index_at        TIMESTAMP,
	last_incremental_index_at TIMESTAMP,
	status                    TEXT NOT NULL DEFAULT 'clean'
);

-- Not an "external content" FTS table on purpose: chunk/document ids are
-- text (content-derived, see domain.contentID), and FTS5's external
-- content sync requires an integer rowid mapping. Storing the searchable
-- text alongside plain lookup columns keeps this simple for v1 at the
-- cost of one duplicate copy of chunk content.
CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
	content,
	chunk_id UNINDEXED,
	document_id UNINDEXED,
	heading UNINDEXED
);
`

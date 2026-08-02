package sqlite

import (
	"context"
	"fmt"
)

// DeleteDocument implements kernel.Storage. Explicit application-layer
// deletes inside one transaction, in dependency order — not ON DELETE
// CASCADE (RFC-0003's resolution of DATABASE.md's open question).
func (s *Store) DeleteDocument(ctx context.Context, documentID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: delete document %s: begin tx: %w", documentID, err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmts := []struct {
		sql  string
		args []any
	}{
		{`DELETE FROM chunks_fts WHERE document_id = ?`, []any{documentID}},
		{`DELETE FROM document_chunks WHERE document_id = ?`, []any{documentID}},
		{`DELETE FROM tags WHERE document_id = ?`, []any{documentID}},
		{`DELETE FROM relationships WHERE from_document_id = ? OR to_document_id = ?`, []any{documentID, documentID}},
		{`DELETE FROM documents WHERE id = ?`, []any{documentID}},
	}
	for _, stmt := range stmts {
		if _, err := tx.ExecContext(ctx, stmt.sql, stmt.args...); err != nil {
			return fmt.Errorf("sqlite: delete document %s: %w", documentID, err)
		}
	}
	return tx.Commit()
}

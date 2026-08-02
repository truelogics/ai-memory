package sqlite

import (
	"context"
	"fmt"
)

// TraverseRelationships implements kernel.Storage: a bounded `WITH
// RECURSIVE` walk over the relationships table, not a second storage
// engine (RFC-0003/GRAPH.md).
func (s *Store) TraverseRelationships(ctx context.Context, documentID string, depth int) ([]string, error) {
	if depth < 0 {
		depth = 0
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH RECURSIVE walk(id, hops) AS (
			SELECT ?, 0
			UNION
			SELECT CASE WHEN r.from_document_id = w.id THEN r.to_document_id ELSE r.from_document_id END,
			       w.hops + 1
			FROM relationships r
			JOIN walk w ON (r.from_document_id = w.id OR r.to_document_id = w.id)
			WHERE w.hops < ?
		)
		SELECT DISTINCT id FROM walk WHERE id != ?
	`, documentID, depth, documentID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: traverse relationships from %s: %w", documentID, err)
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("sqlite: scan traversal id: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

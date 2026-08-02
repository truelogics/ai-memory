package domain

import (
	"errors"
	"strings"
)

// Tag is freeform key/value data attached to a Document — anything
// Metadata's schema doesn't cover yet. See KNOWLEDGE_MODEL.md §3 and
// DATABASE.md's `tags` table.
type Tag struct {
	ID         string
	DocumentID string
	Key        string
	Value      string
}

// NewTag validates and constructs a Tag.
func NewTag(documentID, key, value string) (Tag, error) {
	documentID = strings.TrimSpace(documentID)
	key = strings.TrimSpace(key)
	if documentID == "" {
		return Tag{}, errors.New("domain: tag requires a document id")
	}
	if key == "" {
		return Tag{}, errors.New("domain: tag requires a key")
	}
	return Tag{
		ID:         contentID(documentID, key, value),
		DocumentID: documentID,
		Key:        key,
		Value:      value,
	}, nil
}

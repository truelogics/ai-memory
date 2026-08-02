package domain

import (
	"errors"
	"strconv"
	"strings"
)

// Chunk is a sub-Document unit — the granularity Search actually indexes,
// so a query returns a matched section instead of "the whole file
// matched." See DATABASE.md's `document_chunks` table.
type Chunk struct {
	ID         string
	DocumentID string
	Index      int
	Heading    string
	Content    string
}

// NewChunk validates and constructs a Chunk. ID is derived from
// (documentID, index) so re-chunking the same document deterministically
// reproduces the same chunk ids.
func NewChunk(documentID string, index int, heading, content string) (Chunk, error) {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return Chunk{}, errors.New("domain: chunk requires a document id")
	}
	if index < 0 {
		return Chunk{}, errors.New("domain: chunk index must be >= 0")
	}
	return Chunk{
		ID:         contentID(documentID, strconv.Itoa(index)),
		DocumentID: documentID,
		Index:      index,
		Heading:    heading,
		Content:    content,
	}, nil
}

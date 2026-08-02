package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// newID returns a random, UUID-v4-shaped identifier for entities that have
// no natural stable key (Workspace, Repository, Source, Tag, Relationship).
func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// contentID derives a stable, deterministic identifier from its parts, so
// re-indexing the same (repository, path) or (document, chunk index)
// always yields the same id instead of a new row each run. See
// DATABASE.md's open question on document identity — this is v1's answer:
// path-based (via a hash of repository + path), not content-hash-based.
func contentID(parts ...string) string {
	h := sha256.New()
	for i, p := range parts {
		if i > 0 {
			h.Write([]byte{'|'})
		}
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

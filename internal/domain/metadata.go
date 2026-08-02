package domain

// Metadata is structured, schema'd key/value data attached to a Document —
// front-matter fields, an ADR's status, a Rule's severity — as distinct
// from Tag's freeform key/value pairs. See KNOWLEDGE_MODEL.md §3 and
// DATABASE.md's `documents.front_matter`.
type Metadata map[string]string

// NewMetadata returns an empty Metadata, ready to populate.
func NewMetadata() Metadata {
	return Metadata{}
}

// Get returns the value for key and whether it was present.
func (m Metadata) Get(key string) (string, bool) {
	v, ok := m[key]
	return v, ok
}

// Set assigns key to value.
func (m Metadata) Set(key, value string) {
	m[key] = value
}

// Clone returns a shallow copy, so callers can mutate without aliasing.
func (m Metadata) Clone() Metadata {
	out := make(Metadata, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

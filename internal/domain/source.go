package domain

import (
	"errors"
	"strings"
)

// SourceType identifies where a Source's content is collected from.
type SourceType string

// SourceTypeFilesystem is v1's only SourceType — a local git repository.
// Future SourceTypes (github, slack, notion, ...) are additional values
// here plus a matching Collector implementation; Source itself doesn't
// change shape. See KNOWLEDGE_MODEL.md §1 and INTERFACES.md's open
// question on Source vs. Collector.
const SourceTypeFilesystem SourceType = "filesystem"

// Source represents where Documents originate — the general concept
// Repository is a v1 instance of. See DOMAIN_MODEL.md and
// KNOWLEDGE_MODEL.md's Core Entities.
type Source struct {
	ID           string
	RepositoryID string
	Type         SourceType
}

// NewSource validates and constructs a Source.
func NewSource(repositoryID string, sourceType SourceType) (Source, error) {
	repositoryID = strings.TrimSpace(repositoryID)
	if repositoryID == "" {
		return Source{}, errors.New("domain: source requires a repository id")
	}
	if strings.TrimSpace(string(sourceType)) == "" {
		return Source{}, errors.New("domain: source requires a type")
	}
	return Source{
		ID:           newID(),
		RepositoryID: repositoryID,
		Type:         sourceType,
	}, nil
}

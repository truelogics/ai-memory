package domain

import (
	"errors"
	"strings"
	"time"
)

// Workspace is the scope of one Engineering Memory index — one or more
// Repositories indexed together. See DOMAIN_MODEL.md.
type Workspace struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// NewWorkspace validates and constructs a Workspace.
func NewWorkspace(name string) (Workspace, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Workspace{}, errors.New("domain: workspace name is required")
	}
	return Workspace{
		ID:        newID(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}, nil
}

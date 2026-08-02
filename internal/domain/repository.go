package domain

import (
	"errors"
	"strings"
	"time"
)

// Repository is a git repo registered into a Workspace — v1's only Source
// type. See DOMAIN_MODEL.md and DATABASE.md's `repositories` table.
type Repository struct {
	ID                string
	WorkspaceID       string
	Name              string
	RemoteURL         string
	LocalPath         string
	LastIndexedCommit string
	LastIndexedAt     time.Time
}

// NewRepository validates and constructs a Repository. RemoteURL is
// optional (best-effort — a repo may have no configured remote).
func NewRepository(workspaceID, name, localPath string) (Repository, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	name = strings.TrimSpace(name)
	localPath = strings.TrimSpace(localPath)
	if workspaceID == "" {
		return Repository{}, errors.New("domain: repository requires a workspace id")
	}
	if name == "" {
		return Repository{}, errors.New("domain: repository name is required")
	}
	if localPath == "" {
		return Repository{}, errors.New("domain: repository local path is required")
	}
	return Repository{
		ID:          newID(),
		WorkspaceID: workspaceID,
		Name:        name,
		LocalPath:   localPath,
	}, nil
}

// WithRemote returns a copy of r with RemoteURL set.
func (r Repository) WithRemote(url string) Repository {
	r.RemoteURL = strings.TrimSpace(url)
	return r
}

// MarkIndexed returns a copy of r recording a completed index run.
func (r Repository) MarkIndexed(commit string, at time.Time) Repository {
	r.LastIndexedCommit = commit
	r.LastIndexedAt = at
	return r
}

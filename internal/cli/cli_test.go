package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// TestDefinitionOfDone exercises exactly the command sequence Step 7's
// Definition of Done specifies: init, index ., search (twice), status —
// on a directory containing markdown documentation.
func TestDefinitionOfDone(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "---\ndoc: README\nstatus: living\n---\n\n# Demo Project\n\nWe chose JWT for stateless authentication across services.\n")
	writeFile(t, dir, "docs/ARCHITECTURE.md", "# Architecture\n\nThis document describes the pipeline architecture of the system.\n")

	var out bytes.Buffer

	if err := Init(ctx, dir, &out); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if !strings.Contains(out.String(), "Created workspace") {
		t.Errorf("Init output = %q, want it to mention workspace creation", out.String())
	}
	if _, err := os.Stat(dbPath(dir)); err != nil {
		t.Fatalf("expected %s to exist after Init: %v", dbPath(dir), err)
	}

	out.Reset()
	if err := Index(ctx, dir, &out); err != nil {
		t.Fatalf("Index: %v", err)
	}
	if !strings.Contains(out.String(), "2 scanned") || !strings.Contains(out.String(), "2 added") {
		t.Errorf("Index output = %q, want 2 scanned, 2 added", out.String())
	}

	out.Reset()
	if err := Search(ctx, dir, "architecture", &out); err != nil {
		t.Fatalf("Search(architecture): %v", err)
	}
	if !strings.Contains(out.String(), "ARCHITECTURE.md") {
		t.Errorf("Search(architecture) output = %q, want a hit on ARCHITECTURE.md", out.String())
	}

	out.Reset()
	if err := Search(ctx, dir, "authentication", &out); err != nil {
		t.Fatalf("Search(authentication): %v", err)
	}
	if !strings.Contains(out.String(), "README.md") {
		t.Errorf("Search(authentication) output = %q, want a hit on README.md", out.String())
	}

	out.Reset()
	if err := Status(ctx, dir, &out); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !strings.Contains(out.String(), "clean") {
		t.Errorf("Status output = %q, want status 'clean'", out.String())
	}
}

func TestIndexWithoutInitFails(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := Index(context.Background(), dir, &out); err == nil {
		t.Fatal("Index: expected error when no workspace has been initialized")
	}
}

func TestSearchWithoutInitFails(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := Search(context.Background(), dir, "anything", &out); err == nil {
		t.Fatal("Search: expected error when no workspace has been initialized")
	}
}

func TestStatusWithoutInitFails(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	if err := Status(context.Background(), dir, &out); err == nil {
		t.Fatal("Status: expected error when no workspace has been initialized")
	}
}

func TestStatusWithNoRepositoriesRegistered(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	var out bytes.Buffer
	if err := Init(ctx, dir, &out); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// Init itself registers dir as a repository, so status should show
	// one entry even before any index run.
	out.Reset()
	if err := Status(ctx, dir, &out); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !strings.Contains(out.String(), "not indexed") {
		t.Errorf("Status output = %q, want 'not indexed' before any eng index run", out.String())
	}
}

func TestInitIsIdempotent(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	var out bytes.Buffer

	if err := Init(ctx, dir, &out); err != nil {
		t.Fatalf("Init (first): %v", err)
	}
	out.Reset()
	if err := Init(ctx, dir, &out); err != nil {
		t.Fatalf("Init (second): %v", err)
	}
	if !strings.Contains(out.String(), "Already registered") {
		t.Errorf("second Init output = %q, want 'Already registered'", out.String())
	}
}

func TestSearchNoMatches(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Hello\n\nNothing relevant here.\n")
	var out bytes.Buffer
	if err := Init(ctx, dir, &out); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := Index(ctx, dir, &out); err != nil {
		t.Fatalf("Index: %v", err)
	}
	out.Reset()
	if err := Search(ctx, dir, "zzzznonexistent", &out); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if !strings.Contains(out.String(), "No matches") {
		t.Errorf("Search output = %q, want 'No matches'", out.String())
	}
}

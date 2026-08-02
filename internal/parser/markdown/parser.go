// Package markdown implements kernel.Parser for markdown — v1's only
// Parser. See ARCHITECTURE.md and RFC-0002's "why Parsers per format."
package markdown

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"

	"github.com/truelogics/ai-memory/internal/domain"
	"github.com/truelogics/ai-memory/internal/kernel"
)

// Parser converts markdown RawDocuments into CanonicalDocuments: it
// splits YAML front-matter from the body, and understands headings, code
// blocks, links, and tables well enough to extract a Title and structural
// Tags — using goldmark's CommonMark+GFM AST rather than regexing markdown
// as if it were plain text.
type Parser struct {
	md goldmark.Markdown
}

var _ kernel.Parser = (*Parser)(nil)

// New returns a Parser configured for GitHub-flavored markdown (tables,
// strikethrough, autolinks) — the dialect this repo's own docs use.
func New() *Parser {
	return &Parser{md: goldmark.New(goldmark.WithExtensions(extension.GFM))}
}

// CanParse reports whether raw looks like markdown, by extension.
func (p *Parser) CanParse(raw domain.RawDocument) bool {
	ext := strings.ToLower(path.Ext(raw.Path))
	return ext == ".md" || ext == ".markdown"
}

// Parse implements kernel.Parser.
func (p *Parser) Parse(ctx context.Context, raw domain.RawDocument) (domain.CanonicalDocument, error) {
	frontMatter, body := splitFrontMatter(raw.Bytes)

	// v1 has no persisted `sources` table (see DATABASE.md) — Source is a
	// domain concept, not yet an independent row, so SourceID doubles as
	// RepositoryID until a second Source type exists.
	doc, err := domain.NewCanonicalDocument(raw.SourceID, raw.SourceID, raw.Path)
	if err != nil {
		return domain.CanonicalDocument{}, fmt.Errorf("markdown: %s: %w", raw.Path, err)
	}

	if len(frontMatter) > 0 {
		if err := applyFrontMatter(&doc, frontMatter); err != nil {
			return domain.CanonicalDocument{}, fmt.Errorf("markdown: front-matter in %s: %w", raw.Path, err)
		}
	}

	title, stats := analyze(p.md, body)
	if title == "" {
		if t, ok := doc.Metadata.Get("doc"); ok && t != "" {
			title = t
		} else {
			title = raw.Path
		}
	}

	doc.Title = title
	doc.Content = string(body)
	doc.Type = inferDocType(raw.Path, doc.Metadata)
	doc.ContentHash = contentHash(raw.Bytes)
	doc.GitUpdatedAt = raw.FetchedAt // best-effort; Indexer may overwrite with real git metadata

	if err := appendStatsTags(&doc, stats); err != nil {
		return domain.CanonicalDocument{}, fmt.Errorf("markdown: %s: %w", raw.Path, err)
	}

	return doc, nil
}

// splitFrontMatter separates a leading `---\n...\n---\n` YAML block from
// the rest of the file. Returns a nil front-matter slice if none is
// present.
func splitFrontMatter(raw []byte) (frontMatter, body []byte) {
	const delim = "---"
	if !bytes.HasPrefix(raw, []byte(delim)) {
		return nil, raw
	}
	rest := raw[len(delim):]
	rest = bytes.TrimPrefix(rest, []byte("\n"))
	rest = bytes.TrimPrefix(rest, []byte("\r\n"))

	end := bytes.Index(rest, []byte("\n"+delim))
	if end == -1 {
		return nil, raw
	}
	frontMatter = rest[:end]

	afterDelim := rest[end+1+len(delim):]
	afterDelim = bytes.TrimPrefix(afterDelim, []byte("\r\n"))
	afterDelim = bytes.TrimPrefix(afterDelim, []byte("\n"))
	return frontMatter, afterDelim
}

// applyFrontMatter parses YAML front-matter into doc.Metadata (scalar
// fields) and doc.Tags (one Tag per item of a list-valued field — e.g.
// `audience: [human, agent]` becomes two Tags, since Metadata holds one
// value per key and lists don't fit that shape). See KNOWLEDGE_MODEL.md §3
// on the Metadata/Tag split.
func applyFrontMatter(doc *domain.CanonicalDocument, frontMatter []byte) error {
	var fm map[string]any
	if err := yaml.Unmarshal(frontMatter, &fm); err != nil {
		return err
	}
	for k, v := range fm {
		switch val := v.(type) {
		case []any:
			for _, item := range val {
				tag, err := domain.NewTag(doc.ID, k, fmt.Sprint(item))
				if err != nil {
					return err
				}
				doc.Tags = append(doc.Tags, tag)
			}
		case nil:
			// e.g. `supersedes:` left blank — nothing to record.
		default:
			doc.Metadata.Set(k, fmt.Sprint(val))
		}
	}
	return nil
}

type bodyStats struct {
	headings, codeBlocks, links, tables int
}

// analyze walks body's goldmark AST for the first level-1 heading (the
// Title fallback) and structural counts, proving the parser understands
// headings/code blocks/links/tables rather than treating markdown as
// opaque text.
func analyze(md goldmark.Markdown, body []byte) (title string, stats bodyStats) {
	root := md.Parser().Parse(text.NewReader(body))
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch tn := n.(type) {
		case *ast.Heading:
			stats.headings++
			if title == "" && tn.Level == 1 {
				title = strings.TrimSpace(string(tn.Text(body)))
			}
		case *ast.CodeBlock:
			stats.codeBlocks++
		case *ast.FencedCodeBlock:
			stats.codeBlocks++
		case *ast.Link:
			stats.links++
		case *ast.AutoLink:
			stats.links++
		case *extast.Table:
			stats.tables++
		}
		return ast.WalkContinue, nil
	})
	return title, stats
}

// appendStatsTags records bodyStats as Tags — informative, not schema'd,
// so they belong as Tags rather than Metadata.
func appendStatsTags(doc *domain.CanonicalDocument, stats bodyStats) error {
	counts := []struct {
		key string
		n   int
	}{
		{"heading_count", stats.headings},
		{"code_block_count", stats.codeBlocks},
		{"link_count", stats.links},
		{"table_count", stats.tables},
	}
	for _, c := range counts {
		if c.n == 0 {
			continue
		}
		tag, err := domain.NewTag(doc.ID, c.key, fmt.Sprint(c.n))
		if err != nil {
			return err
		}
		doc.Tags = append(doc.Tags, tag)
	}
	return nil
}

// inferDocType applies ARCHITECTURE.md's stated heuristic: front-matter
// `doc:` field and path conventions, falling back to unknown.
func inferDocType(p string, meta domain.Metadata) domain.DocType {
	lower := strings.ToLower(p)
	base := strings.ToLower(path.Base(p))
	segments := strings.Split(lower, "/")
	hasSegment := func(names ...string) bool {
		for _, seg := range segments {
			for _, name := range names {
				if seg == name {
					return true
				}
			}
		}
		return false
	}

	if doc, ok := meta.Get("doc"); ok {
		switch strings.ToUpper(doc) {
		case "RFC":
			return domain.DocTypeRFC
		case "ARCHITECTURE", "DATABASE", "DOMAIN_MODEL", "KNOWLEDGE_MODEL", "INTERFACES", "CLI":
			return domain.DocTypeStandard
		case "ROADMAP", "NOW", "NEXT", "BACKLOG":
			return domain.DocTypeRoadmap
		case "README":
			return domain.DocTypeReadme
		}
	}

	switch {
	case base == "readme.md":
		return domain.DocTypeReadme
	case hasSegment("adr", "adrs"):
		return domain.DocTypeADR
	case hasSegment("rfc", "rfcs"):
		return domain.DocTypeRFC
	case hasSegment("rules"):
		return domain.DocTypeRule
	case hasSegment("roadmap"):
		return domain.DocTypeRoadmap
	default:
		return domain.DocTypeUnknown
	}
}

func contentHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

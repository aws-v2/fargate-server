package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// =====================
// Models (match frontend)
// =====================

type DocItem struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type DocCategory struct {
	Title string    `json:"title"`
	Items []DocItem `json:"items"`
}

type DocManifest struct {
	Service    string        `json:"service"`
	Version    string        `json:"version,omitempty"`
	Categories []DocCategory `json:"categories"`
}

type Metadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	LastUpdated string   `json:"lastUpdated"`
	Tags        []string `json:"tags"`
}

type DocResponse struct {
	Metadata Metadata `json:"metadata"`
	Content  string   `json:"content"`
}

// =====================
// Service
// =====================

type DocsService struct {
	basePath string // e.g. "./docs"
}

func NewDocsService(basePath string) *DocsService {
	return &DocsService{basePath: basePath}
}

// =====================
// Public API
// =====================

// GetManifest loads manifest.json from public/internal folder
func (s *DocsService) GetManifest(internal bool) (*DocManifest, error) {
	scope := s.getScope(internal)

	path := filepath.Join(s.basePath, scope, "manifest.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest DocManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &manifest, nil
}

// GetDoc loads a markdown file and parses frontmatter
func (s *DocsService) GetDoc(slug string, internal bool) (*DocResponse, error) {
	if !isValidSlug(slug) {
		return nil, errors.New("invalid slug")
	}

	scope := s.getScope(internal)
	path := filepath.Join(s.basePath, scope, slug+".md")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("not found")
	}

	meta, content := parseMarkdownWithFrontmatter(string(data))

	return &DocResponse{
		Metadata: meta,
		Content:  content,
	}, nil
}

// =====================
// Helpers
// =====================

func (s *DocsService) getScope(internal bool) string {
	if internal {
		return "internal"
	}
	return "public"
}

// Prevent path traversal attacks
func isValidSlug(slug string) bool {
	if slug == "" {
		return false
	}
	if strings.Contains(slug, "..") || strings.Contains(slug, "/") || strings.Contains(slug, "\\") {
		return false
	}
	return true
}

// =====================
// Markdown Parser
// =====================

func parseMarkdownWithFrontmatter(input string) (Metadata, string) {
	var meta Metadata

	parts := strings.SplitN(input, "---", 3)

	// No frontmatter
	if len(parts) < 3 {
		meta.LastUpdated = time.Now().Format("2006-01-02")
		return meta, strings.TrimSpace(input)
	}

	rawMeta := parts[1]
	content := parts[2]

	lines := strings.Split(rawMeta, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "title:"):
			meta.Title = cleanValue(line, "title:")
		case strings.HasPrefix(line, "description:"):
			meta.Description = cleanValue(line, "description:")
		case strings.HasPrefix(line, "icon:"):
			meta.Icon = cleanValue(line, "icon:")
		case strings.HasPrefix(line, "tags:"):
			meta.Tags = parseTags(cleanValue(line, "tags:"))
		}
	}

	meta.LastUpdated = time.Now().Format("2006-01-02")

	return meta, strings.TrimSpace(content)
}

func cleanValue(line, prefix string) string {
	val := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	val = strings.Trim(val, `"`)
	return val
}

func parseTags(input string) []string {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")

	var tags []string
	for _, t := range parts {
		tag := strings.TrimSpace(strings.Trim(t, `"`))
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
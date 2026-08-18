package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

const memoryDir = "docs/memory"

// CreateMemory validates and records a new memory entry (CONTRACT.md
// `memory add`), markdown-first: docs/memory/{id}.md is written before the
// derived `memories` index row, mirroring CreateTrace's ordering (R1,
// docs/plans/active/durable-memory.md) — a failed markdown write leaves
// zero DB rows.
func CreateMemory(db *sql.DB, memType, scope, planID, summary string) (id string, err error) {
	entity := domain.Memory{Type: memType, Scope: scope, PlanID: planID, Summary: summary}
	if err := entity.Validate(); err != nil {
		return "", err
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	path := fmt.Sprintf("%s/%s.md", memoryDir, id)
	content := formatMemoryEntry(id, memType, scope, planID, at, summary)

	// Markdown is the write target: the file write runs before the DB
	// write, so a failed markdown write leaves zero DB rows behind it,
	// matching CreateTrace's ordering (R1).
	if err := writeManagedFile(path, []byte(content)); err != nil {
		return "", fmt.Errorf("memory %s: markdown write failed: %w", id, err)
	}

	sha := managedDocSHA256([]byte(content))
	var planIDArg any
	if planID != "" {
		planIDArg = planID
	}
	if _, err := db.Exec(
		`INSERT INTO memories (id, path, type, scope, plan_id, sha256, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, path, memType, scope, planIDArg, sha, at,
	); err != nil {
		return "", fmt.Errorf("memory %s: markdown recorded, but db write failed: %w", id, err)
	}
	return id, nil
}

// formatMemoryEntry renders docs/memory/{id}.md: frontmatter (id, type,
// scope, plan_id when scope=plan, created) followed by the entry body.
func formatMemoryEntry(id, memType, scope, planID, at, summary string) string {
	fm := fmt.Sprintf("---\nid: %s\ntype: %s\nscope: %s\n", id, memType, scope)
	if planID != "" {
		fm += fmt.Sprintf("plan_id: %s\n", planID)
	}
	fm += fmt.Sprintf("created: %s\n---\n\n%s\n", at, summary)
	return fm
}

const memoryFrontmatterLineLimit = 10

// MemoryView is the `memory get --id` view: one docs/memory/{id}.md entry's
// frontmatter plus body, read back off the markdown file itself rather than
// reconstructed from the `memories` index row — markdown is the source of
// truth (R1, docs/plans/active/durable-memory.md), the index is derived.
type MemoryView struct {
	ID        string  `json:"id"`
	Path      string  `json:"path"`
	Type      string  `json:"type"`
	Scope     string  `json:"scope"`
	PlanID    *string `json:"plan_id"`
	CreatedAt string  `json:"created_at"`
	Body      string  `json:"body"`
}

// MemoryGet resolves id to its indexed path, reads the markdown file, and
// parses its frontmatter (R3: direct read-back by ID, no ranking/search).
func MemoryGet(db *sql.DB, id string) (MemoryView, error) {
	var path string
	err := db.QueryRow(`SELECT path FROM memories WHERE id = ?`, id).Scan(&path)
	if err == sql.ErrNoRows {
		return MemoryView{}, &domain.ValidationError{Code: "unknown_memory_id", Message: "memory get: id " + id + " not found"}
	}
	if err != nil {
		return MemoryView{}, fmt.Errorf("memory get %s: query index: %w", id, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return MemoryView{}, fmt.Errorf("memory get %s: read %s: %w", id, path, err)
	}

	view := MemoryView{ID: id, Path: path, Body: extractMemoryBody(string(content))}
	if lines, ok := frontmatterPreview(string(content), memoryFrontmatterLineLimit); ok {
		if v, ok := frontmatterPreviewField(lines, "type"); ok {
			view.Type = v
		}
		if v, ok := frontmatterPreviewField(lines, "scope"); ok {
			view.Scope = v
		}
		if v, ok := frontmatterPreviewField(lines, "plan_id"); ok && v != "" {
			view.PlanID = &v
		}
		if v, ok := frontmatterPreviewField(lines, "created"); ok {
			view.CreatedAt = v
		}
	}
	return view, nil
}

// extractMemoryBody returns the markdown body following the frontmatter's
// closing `---`, or the whole trimmed content when there is no frontmatter
// block — matching frontmatterPreview's own tolerance for a malformed file.
func extractMemoryBody(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return strings.TrimSpace(content)
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		}
	}
	return strings.TrimSpace(content)
}

// MemoryListView is one row of the `memory query` view: path plus metadata
// only, no body content — an index/list operation, not ranking or
// cross-source routing, which stays exclusively P6 (retrieval-router)'s
// scope (R3, NG1, docs/plans/active/durable-memory.md).
type MemoryListView struct {
	ID        string  `json:"id"`
	Path      string  `json:"path"`
	Type      string  `json:"type"`
	Scope     string  `json:"scope"`
	PlanID    *string `json:"plan_id"`
	CreatedAt string  `json:"created_at"`
}

// MemoryQuery lists memories index rows filtered by type (required) and
// optionally scope and/or plan_id, newest first.
func MemoryQuery(db *sql.DB, memType, scope, planID string) ([]MemoryListView, error) {
	if memType == "" {
		return nil, &domain.ValidationError{Code: "missing_required_field", Message: "memory query: --type is required"}
	}
	if scope != "" && !domain.IsValidMemoryScope(scope) {
		return nil, &domain.ValidationError{Code: "invalid_scope", Message: "memory query: invalid scope " + scope + " (want plan|global)"}
	}

	q := `SELECT id, path, type, scope, plan_id, created_at FROM memories WHERE type = ?`
	args := []any{memType}
	if scope != "" {
		q += ` AND scope = ?`
		args = append(args, scope)
	}
	if planID != "" {
		q += ` AND plan_id = ?`
		args = append(args, planID)
	}
	q += ` ORDER BY created_at DESC, id DESC`

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("memory query: %w", err)
	}
	defer rows.Close()

	views := []MemoryListView{}
	for rows.Next() {
		var v MemoryListView
		var planIDCol sql.NullString
		if err := rows.Scan(&v.ID, &v.Path, &v.Type, &v.Scope, &planIDCol, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("memory query: scan row: %w", err)
		}
		v.PlanID = nullableString(planIDCol)
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory query: %w", err)
	}
	return views, nil
}

// rebuildMemoriesFromMarkdown reconstructs the memories index from every
// committed docs/memory/*.md entry, with no read of any non-committed
// state (R4, docs/plans/active/durable-memory.md). Unlike plan_index,
// which refreshes lazily whenever a plan is read, memory rows have no
// read path that already touches the file (MemoryGet/MemoryQuery only
// read the index), so db rebuild is the sole reconstruction path.
func rebuildMemoriesFromMarkdown(db *sql.DB, result *RebuildResult) error {
	paths, err := filepath.Glob(filepath.Join(memoryDir, "*.md"))
	if err != nil {
		return fmt.Errorf("glob %s: %w", memoryDir, err)
	}
	sort.Strings(paths)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		content := string(data)

		lines, ok := frontmatterPreview(content, memoryFrontmatterLineLimit)
		if !ok {
			continue // hand-authored or malformed entry: degrade, skip
		}
		id, ok := frontmatterPreviewField(lines, "id")
		if !ok || id == "" {
			continue
		}
		memType, _ := frontmatterPreviewField(lines, "type")
		scope, _ := frontmatterPreviewField(lines, "scope")
		planID, _ := frontmatterPreviewField(lines, "plan_id")
		createdAt, _ := frontmatterPreviewField(lines, "created")

		var planIDArg any
		if planID != "" {
			planIDArg = planID
		}
		if _, err := db.Exec(
			`INSERT INTO memories (id, path, type, scope, plan_id, sha256, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			id, path, memType, scope, planIDArg, managedDocSHA256(data), createdAt,
		); err != nil {
			return fmt.Errorf("insert memory %s: %w", id, err)
		}
		result.Memories++
	}
	return nil
}

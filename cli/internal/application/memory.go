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

// vietFold maps Vietnamese precomposed characters to their ASCII base after
// case-folding. Only lowercased keys are needed because normalizeFold lowercases
// first; combining marks 0300-036F are stripped separately so both precomposed
// and decomposed forms fold to the same ASCII.
var vietFold = map[rune]rune{
	'á': 'a', 'à': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
	'â': 'a', 'ấ': 'a', 'ầ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
	'ă': 'a', 'ắ': 'a', 'ằ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
	'é': 'e', 'è': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
	'ê': 'e', 'ế': 'e', 'ề': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
	'í': 'i', 'ì': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
	'ó': 'o', 'ò': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
	'ô': 'o', 'ố': 'o', 'ồ': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
	'ơ': 'o', 'ớ': 'o', 'ờ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
	'ú': 'u', 'ù': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
	'ư': 'u', 'ứ': 'u', 'ừ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
	'ý': 'y', 'ỳ': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
	'đ': 'd',
}

// normalizeFold case-folds and strips Vietnamese diacritics so "kiem tra"
// matches "kiểm tra" and "dong bo" matches "đồng bộ". It handles both
// precomposed characters (via vietFold) and decomposed combining marks
// (U+0300-036F stripped). Stdlib only, zero new modules.
func normalizeFold(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 0x0300 && r <= 0x036F {
			continue
		}
		if base, ok := vietFold[r]; ok {
			b.WriteRune(base)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

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

const memoryFrontmatterLineLimit = 12

// MemoryView is the `memory get --id` view: one docs/memory/{id}.md entry's
// frontmatter plus body, read back off the markdown file itself rather than
// reconstructed from the `memories` index row — markdown is the source of
// truth (R1, docs/plans/active/durable-memory.md), the index is derived.
type MemoryView struct {
	ID           string  `json:"id"`
	Path         string  `json:"path"`
	Type         string  `json:"type"`
	Scope        string  `json:"scope"`
	PlanID       *string `json:"plan_id"`
	CreatedAt    string  `json:"created_at"`
	Body         string  `json:"body"`
	Status       string  `json:"status"`
	SupersededBy *string `json:"superseded_by"`
	SupersededAt *string `json:"superseded_at"`
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

	view := MemoryView{ID: id, Path: path, Body: extractMemoryBody(string(content)), Status: "active"}
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
		if v, ok := frontmatterPreviewField(lines, "superseded_by"); ok && v != "" {
			view.SupersededBy = &v
			view.Status = "superseded"
		}
		if v, ok := frontmatterPreviewField(lines, "superseded_at"); ok && v != "" {
			view.SupersededAt = &v
		}
	}
	// Fallback: if frontmatter says superseded but DB says not, DB is derived and may be stale
	// before rebuild — frontmatter wins. If file lacks superseded but DB has it (should not happen
	// outside a mid-write crash), we also surface DB value as superseded.
	if view.Status != "superseded" {
		var dbSuperseded sql.NullString
		var dbSupersededAt sql.NullString
		// Column may not exist on old DBs before migration 0013 — ignore error and treat as active.
		err := db.QueryRow(`SELECT superseded_by, superseded_at FROM memories WHERE id = ?`, id).Scan(&dbSuperseded, &dbSupersededAt)
		if err == nil && dbSuperseded.Valid && dbSuperseded.String != "" {
			view.SupersededBy = &dbSuperseded.String
			view.Status = "superseded"
			if dbSupersededAt.Valid && dbSupersededAt.String != "" {
				v := dbSupersededAt.String
				view.SupersededAt = &v
			}
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
	ID           string  `json:"id"`
	Path         string  `json:"path"`
	Type         string  `json:"type"`
	Scope        string  `json:"scope"`
	PlanID       *string `json:"plan_id"`
	CreatedAt    string  `json:"created_at"`
	SupersededBy *string `json:"superseded_by"`
	SupersededAt *string `json:"superseded_at"`
}

// MemoryQuery lists memories index rows filtered by type (required) and
// optionally scope and/or plan_id, newest first. Superseded entries are
// excluded by default (R2); use MemoryQueryWithIncludeSuperseded to restore.
func MemoryQuery(db *sql.DB, memType, scope, planID string) ([]MemoryListView, error) {
	return memoryQueryInternal(db, memType, scope, planID, false)
}

// MemoryQueryWithIncludeSuperseded is the include-superseded variant of
// MemoryQuery (R2). When include is false it behaves identically to
// MemoryQuery; when true, superseded entries are returned alongside active
// ones with their superseded_by/at populated.
func MemoryQueryWithIncludeSuperseded(db *sql.DB, memType, scope, planID string, include bool) ([]MemoryListView, error) {
	return memoryQueryInternal(db, memType, scope, planID, include)
}

func memoryQueryInternal(db *sql.DB, memType, scope, planID string, includeSuperseded bool) ([]MemoryListView, error) {
	if memType == "" {
		return nil, &domain.ValidationError{Code: "missing_required_field", Message: "memory query: --type is required"}
	}
	if scope != "" && !domain.IsValidMemoryScope(scope) {
		return nil, &domain.ValidationError{Code: "invalid_scope", Message: "memory query: invalid scope " + scope + " (want plan|global)"}
	}

	q := `SELECT id, path, type, scope, plan_id, created_at, superseded_by, superseded_at FROM memories WHERE type = ?`
	args := []any{memType}
	if scope != "" {
		q += ` AND scope = ?`
		args = append(args, scope)
	}
	if planID != "" {
		q += ` AND plan_id = ?`
		args = append(args, planID)
	}
	if !includeSuperseded {
		q += ` AND superseded_by IS NULL`
	}
	q += ` ORDER BY created_at DESC, id DESC`

	rows, err := db.Query(q, args...)
	if err != nil {
		// Fallback for DBs missing superseded columns (pre-0013): retry without those columns and filter is no-op.
		if strings.Contains(err.Error(), "no such column") && strings.Contains(err.Error(), "superseded_by") {
			return memoryQueryLegacyFallback(db, memType, scope, planID)
		}
		return nil, fmt.Errorf("memory query: %w", err)
	}
	defer rows.Close()

	views := []MemoryListView{}
	for rows.Next() {
		var v MemoryListView
		var planIDCol sql.NullString
		var supBy sql.NullString
		var supAt sql.NullString
		if err := rows.Scan(&v.ID, &v.Path, &v.Type, &v.Scope, &planIDCol, &v.CreatedAt, &supBy, &supAt); err != nil {
			return nil, fmt.Errorf("memory query: scan row: %w", err)
		}
		v.PlanID = nullableString(planIDCol)
		if supBy.Valid && supBy.String != "" {
			v.SupersededBy = &supBy.String
		}
		if supAt.Valid && supAt.String != "" {
			v.SupersededAt = &supAt.String
		}
		views = append(views, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory query: %w", err)
	}
	return views, nil
}

func memoryQueryLegacyFallback(db *sql.DB, memType, scope, planID string) ([]MemoryListView, error) {
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

// MemoryScoredView is one row of the `memory query --keywords` ranked view:
// MemoryListView's fields plus the keyword-match Score that produced its
// rank (R1, docs/plans/active/retrieval-router.md — keyword ranking, not
// semantic search).
type MemoryScoredView struct {
	ID           string  `json:"id"`
	Path         string  `json:"path"`
	Type         string  `json:"type"`
	Scope        string  `json:"scope"`
	PlanID       *string `json:"plan_id"`
	CreatedAt    string  `json:"created_at"`
	Score        int     `json:"score"`
	SupersededBy *string `json:"superseded_by"`
	SupersededAt *string `json:"superseded_at"`
}

// MemoryQueryRanked lists memories index rows ranked by case-insensitive
// keyword-token match count against each entry's type and markdown body,
// optionally narrowed first by the same type/scope/plan_id filters
// MemoryQuery accepts. Zero-score entries are dropped; the rest are
// ordered by score descending, then created_at descending as a tiebreak
// (R1/R5, docs/plans/active/retrieval-router.md). No new external
// dependency and no schema change — scoring runs entirely in Go over
// markdown already read via the same path MemoryGet uses.
func MemoryQueryRanked(db *sql.DB, keywords, memType, scope, planID string) ([]MemoryScoredView, error) {
	return memoryQueryRankedInternal(db, keywords, memType, scope, planID, false)
}

// MemoryQueryRankedWithIncludeSuperseded is the include-superseded variant of
// MemoryQueryRanked (R2).
func MemoryQueryRankedWithIncludeSuperseded(db *sql.DB, keywords, memType, scope, planID string, include bool) ([]MemoryScoredView, error) {
	return memoryQueryRankedInternal(db, keywords, memType, scope, planID, include)
}

func memoryQueryRankedInternal(db *sql.DB, keywords, memType, scope, planID string, includeSuperseded bool) ([]MemoryScoredView, error) {
	if strings.TrimSpace(keywords) == "" {
		return nil, &domain.ValidationError{Code: "missing_required_field", Message: "memory query: --keywords is required for ranked mode"}
	}
	if scope != "" && !domain.IsValidMemoryScope(scope) {
		return nil, &domain.ValidationError{Code: "invalid_scope", Message: "memory query: invalid scope " + scope + " (want plan|global)"}
	}

	q := `SELECT id, path, type, scope, plan_id, created_at, superseded_by, superseded_at FROM memories WHERE 1 = 1`
	args := []any{}
	if memType != "" {
		q += ` AND type = ?`
		args = append(args, memType)
	}
	if scope != "" {
		q += ` AND scope = ?`
		args = append(args, scope)
	}
	if planID != "" {
		q += ` AND plan_id = ?`
		args = append(args, planID)
	}
	if !includeSuperseded {
		q += ` AND superseded_by IS NULL`
	}
	q += ` ORDER BY created_at DESC, id DESC`

	rows, err := db.Query(q, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no such column") && strings.Contains(err.Error(), "superseded_by") {
			return memoryQueryRankedLegacyFallback(db, keywords, memType, scope, planID)
		}
		return nil, fmt.Errorf("memory query: %w", err)
	}
	defer rows.Close()

	candidates := []MemoryListView{}
	for rows.Next() {
		var v MemoryListView
		var planIDCol sql.NullString
		var supBy sql.NullString
		var supAt sql.NullString
		if err := rows.Scan(&v.ID, &v.Path, &v.Type, &v.Scope, &planIDCol, &v.CreatedAt, &supBy, &supAt); err != nil {
			return nil, fmt.Errorf("memory query: scan row: %w", err)
		}
		v.PlanID = nullableString(planIDCol)
		if supBy.Valid && supBy.String != "" {
			v.SupersededBy = &supBy.String
		}
		if supAt.Valid && supAt.String != "" {
			v.SupersededAt = &supAt.String
		}
		candidates = append(candidates, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory query: %w", err)
	}

	// Scored in the same descending created_at/id order the SELECT already
	// produced, so the stable sort below preserves that order as the
	// tiebreak for equal scores without a second ORDER BY key.
	tokens := strings.Fields(normalizeFold(keywords))
	scored := make([]MemoryScoredView, 0, len(candidates))
	for _, c := range candidates {
		content, err := os.ReadFile(c.Path)
		if err != nil {
			return nil, fmt.Errorf("memory query: read %s: %w", c.Path, err)
		}
		haystack := normalizeFold(c.Type + " " + extractMemoryBody(string(content)))
		score := 0
		for _, token := range tokens {
			score += strings.Count(haystack, token)
		}
		if score == 0 {
			continue
		}
		scored = append(scored, MemoryScoredView{
			ID: c.ID, Path: c.Path, Type: c.Type, Scope: c.Scope,
			PlanID: c.PlanID, CreatedAt: c.CreatedAt, Score: score,
			SupersededBy: c.SupersededBy, SupersededAt: c.SupersededAt,
		})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	return scored, nil
}

func memoryQueryRankedLegacyFallback(db *sql.DB, keywords, memType, scope, planID string) ([]MemoryScoredView, error) {
	q := `SELECT id, path, type, scope, plan_id, created_at FROM memories WHERE 1 = 1`
	args := []any{}
	if memType != "" {
		q += ` AND type = ?`
		args = append(args, memType)
	}
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
	candidates := []MemoryListView{}
	for rows.Next() {
		var v MemoryListView
		var planIDCol sql.NullString
		if err := rows.Scan(&v.ID, &v.Path, &v.Type, &v.Scope, &planIDCol, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("memory query: scan row: %w", err)
		}
		v.PlanID = nullableString(planIDCol)
		candidates = append(candidates, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory query: %w", err)
	}
	tokens := strings.Fields(normalizeFold(keywords))
	scored := make([]MemoryScoredView, 0, len(candidates))
	for _, c := range candidates {
		content, err := os.ReadFile(c.Path)
		if err != nil {
			return nil, fmt.Errorf("memory query: read %s: %w", c.Path, err)
		}
		haystack := normalizeFold(c.Type + " " + extractMemoryBody(string(content)))
		score := 0
		for _, token := range tokens {
			score += strings.Count(haystack, token)
		}
		if score == 0 {
			continue
		}
		scored = append(scored, MemoryScoredView{
			ID: c.ID, Path: c.Path, Type: c.Type, Scope: c.Scope,
			PlanID: c.PlanID, CreatedAt: c.CreatedAt, Score: score,
		})
	}
	sort.SliceStable(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	return scored, nil
}

// SupersedeMemory records that oldID is superseded by newID (R1). It writes
// superseded_by + superseded_at into the old entry's frontmatter first
// (markdown-first ordering), then mirrors the columns in the memories index.
// Refusals are exit-1 ValidationErrors with codes unknown_memory_id,
// supersede_self, already_superseded.
func SupersedeMemory(db *sql.DB, oldID, newID string) error {
	if oldID == newID {
		return &domain.ValidationError{Code: "supersede_self", Message: "memory supersede: old-id and new-id must differ"}
	}

	// Validate both IDs exist and fetch old's superseded state + paths.
	var oldPath string
	var oldSuperseded sql.NullString
	err := db.QueryRow(`SELECT path, superseded_by FROM memories WHERE id = ?`, oldID).Scan(&oldPath, &oldSuperseded)
	if err == sql.ErrNoRows {
		return &domain.ValidationError{Code: "unknown_memory_id", Message: "memory supersede: old id " + oldID + " not found"}
	}
	if err != nil {
		if strings.Contains(err.Error(), "no such column") {
			// Pre-migration DB: treat as no superseded, but still validate existence via fallback.
			var pathOnly string
			err2 := db.QueryRow(`SELECT path FROM memories WHERE id = ?`, oldID).Scan(&pathOnly)
			if err2 == sql.ErrNoRows {
				return &domain.ValidationError{Code: "unknown_memory_id", Message: "memory supersede: old id " + oldID + " not found"}
			}
			if err2 != nil {
				return fmt.Errorf("memory supersede %s: query index: %w", oldID, err2)
			}
			oldPath = pathOnly
			oldSuperseded = sql.NullString{Valid: false}
		} else {
			return fmt.Errorf("memory supersede %s: query index: %w", oldID, err)
		}
	}
	if oldSuperseded.Valid && oldSuperseded.String != "" {
		return &domain.ValidationError{Code: "already_superseded", Message: "memory supersede: id " + oldID + " already superseded by " + oldSuperseded.String}
	}

	var newPath string
	err = db.QueryRow(`SELECT path FROM memories WHERE id = ?`, newID).Scan(&newPath)
	if err == sql.ErrNoRows {
		return &domain.ValidationError{Code: "unknown_memory_id", Message: "memory supersede: new id " + newID + " not found"}
	}
	if err != nil {
		return fmt.Errorf("memory supersede %s: query new index: %w", newID, err)
	}

	contentBytes, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("memory supersede %s: read %s: %w", oldID, oldPath, err)
	}
	content := string(contentBytes)

	// Guard against frontmatter already carrying superseded_by (re-supersede via file).
	if lines, ok := frontmatterPreview(content, memoryFrontmatterLineLimit); ok {
		if v, ok := frontmatterPreviewField(lines, "superseded_by"); ok && v != "" {
			return &domain.ValidationError{Code: "already_superseded", Message: "memory supersede: id " + oldID + " already superseded by " + v}
		}
	}

	at := time.Now().UTC().Format(time.RFC3339)
	newContent, err := injectSupersededFrontmatter(content, newID, at)
	if err != nil {
		return fmt.Errorf("memory supersede %s: inject frontmatter: %w", oldID, err)
	}

	// Markdown-first: write file before DB.
	if err := writeManagedFile(oldPath, []byte(newContent)); err != nil {
		return fmt.Errorf("memory supersede %s: markdown write failed: %w", oldID, err)
	}

	sha := managedDocSHA256([]byte(newContent))
	// Prefer updating both columns; fallback to no-op if column missing (pre-migration, should not happen in prod after 0013).
	_, err = db.Exec(`UPDATE memories SET superseded_by = ?, superseded_at = ?, sha256 = ? WHERE id = ?`, newID, at, sha, oldID)
	if err != nil && strings.Contains(err.Error(), "no such column") {
		// Column missing — still consider markdown success; DB will catch up after migration+rebuild.
		return nil
	}
	if err != nil {
		return fmt.Errorf("memory supersede %s: markdown recorded, but db write failed: %w", oldID, err)
	}
	return nil
}

func injectSupersededFrontmatter(content, newID, at string) (string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("missing frontmatter block")
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return "", fmt.Errorf("missing frontmatter closing ---")
	}
	// Insert before closing ---.
	insert := []string{fmt.Sprintf("superseded_by: %s", newID), fmt.Sprintf("superseded_at: %s", at)}
	newLines := make([]string, 0, len(lines)+len(insert))
	newLines = append(newLines, lines[:end]...)
	newLines = append(newLines, insert...)
	newLines = append(newLines, lines[end:]...)
	return strings.Join(newLines, "\n"), nil
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
		supersededBy, _ := frontmatterPreviewField(lines, "superseded_by")
		supersededAt, _ := frontmatterPreviewField(lines, "superseded_at")

		var planIDArg any
		if planID != "" {
			planIDArg = planID
		}
		var supByArg any
		if supersededBy != "" {
			supByArg = supersededBy
		}
		var supAtArg any
		if supersededAt != "" {
			supAtArg = supersededAt
		}
		if _, err := db.Exec(
			`INSERT INTO memories (id, path, type, scope, plan_id, sha256, created_at, superseded_by, superseded_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, path, memType, scope, planIDArg, managedDocSHA256(data), createdAt, supByArg, supAtArg,
		); err != nil {
			// Fallback for pre-0013 DBs where columns don't exist (should not happen in migrated prod, but keeps legacy tests green).
			if strings.Contains(err.Error(), "no such column") {
				if _, err2 := db.Exec(
					`INSERT INTO memories (id, path, type, scope, plan_id, sha256, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
					id, path, memType, scope, planIDArg, managedDocSHA256(data), createdAt,
				); err2 != nil {
					return fmt.Errorf("insert memory %s: %w", id, err2)
				}
				result.Memories++
				continue
			}
			return fmt.Errorf("insert memory %s: %w", id, err)
		}
		result.Memories++
	}
	return nil
}

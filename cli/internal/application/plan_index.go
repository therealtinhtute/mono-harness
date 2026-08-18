package application

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/oklog/ulid/v2"
)

// planIndexRow is the plan_index state for one path, as last refreshed.
type planIndexRow struct {
	ID        string
	SHA256    string
	Status    string
	UpdatedAt string
}

// planIndexStaleness compares path's on-disk content against its indexed
// plan_index row, reusing managedDocSHA256 (the managed_docs hashing
// helper) instead of a second hashing path. Staleness is always the 3-way
// comparison on-disk hash vs indexed hash vs indexed row presence — never a
// timestamp guess (R9, docs/plans/active/harness-markdown-truth.md). A path
// with no row yet is stale by definition: nothing has ever indexed it.
func planIndexStaleness(db *sql.DB, path string) (onDiskSHA256 string, indexed bool, stale bool, row planIndexRow, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", false, false, planIndexRow{}, err
	}
	onDiskSHA256 = managedDocSHA256(content)

	row, indexed, err = queryPlanIndexRow(db, path)
	if err != nil {
		return "", false, false, planIndexRow{}, err
	}
	if !indexed {
		return onDiskSHA256, false, true, planIndexRow{}, nil
	}
	return onDiskSHA256, true, onDiskSHA256 != row.SHA256, row, nil
}

func queryPlanIndexRow(db *sql.DB, path string) (row planIndexRow, indexed bool, err error) {
	err = db.QueryRow(`SELECT id, sha256, status, updated_at FROM plan_index WHERE path = ?`, path).
		Scan(&row.ID, &row.SHA256, &row.Status, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		return planIndexRow{}, false, nil
	}
	if err != nil {
		return planIndexRow{}, false, err
	}
	return row, true, nil
}

// refreshPlanIndex reconciles plan_index for path against content's current
// hash, upserting a row when none exists yet or the on-disk hash has moved
// — so any caller reading plan_index next sees fresh data instead of a
// stale cached answer (R9, docs/plans/active/harness-markdown-truth.md,
// wave 3). wasStale reports whether the row was absent or out of date
// before this call, so a caller (resume's stale_index drift finding) can
// name the event instead of refreshing silently.
func refreshPlanIndex(db *sql.DB, path, content string) (wasStale bool, err error) {
	sha := managedDocSHA256([]byte(content))

	row, indexed, err := queryPlanIndexRow(db, path)
	if err != nil {
		return false, err
	}
	if indexed && row.SHA256 == sha {
		return false, nil
	}

	id := row.ID
	if !indexed {
		id = ulid.Make().String()
	}
	status := planFrontmatterStatus(content)
	at := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`
		INSERT INTO plan_index (id, path, sha256, status, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET sha256 = excluded.sha256, status = excluded.status, updated_at = excluded.updated_at
	`, id, path, sha, status, at); err != nil {
		return false, fmt.Errorf("refresh plan_index for %s: %w", path, err)
	}
	return true, nil
}

// planFrontmatterStatus reads content's frontmatter `status:` field via the
// same Tier 1 preview ambiguous-plan disambiguation already uses
// (plan_resolve.go) — empty when frontmatter is missing, unparseable, or
// has no status line.
func planFrontmatterStatus(content string) string {
	lines, ok := frontmatterPreview(content, frontmatterPreviewLines)
	if !ok {
		return ""
	}
	status, _ := frontmatterPreviewField(lines, "status")
	return status
}

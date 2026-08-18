package application

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// ImportResult mirrors CONTRACT.md's `import --json` shape.
type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

// Import parses legacy `{legacyDir}/workflow-state.yml` (plus the
// planning markdown it points to) into DB rows, per STATE.md's Legacy
// Field Mapping. It only creates `stories` rows for the slugs actually
// required by the FK graph (current_phase, entry_phase, and the phase
// embedded in latest_cook_run's filename) — not a full historical import
// of every roadmap phase. It deliberately never synthesizes a `checks` row
// for latest_check_report: checks.verdict is NOT NULL and STATE.md's
// mapping only covers yml fields, never check-report body parsing, so
// meta.latest_check_id is left NULL.
//
// Idempotent by pre-check: each entity is compared against current DB
// state before writing anything, so a second run with unchanged legacy
// input writes nothing new.
func Import(db *sql.DB, legacyDir string) (ImportResult, error) {
	var result ImportResult

	ymlPath := filepath.Join(legacyDir, "workflow-state.yml")
	fields, err := parseFlatYAML(ymlPath)
	if err != nil {
		return result, err
	}
	if err := checkKnownLegacyFields(fields); err != nil {
		return result, err
	}

	currentPhase := normalizePhase(fields["current_phase"])
	entryPhase := normalizePhase(fields["entry_phase"])
	roadmapPath := fields["roadmap"]
	now := time.Now().UTC().Format(time.RFC3339)

	var runSlug, runCreatedAt, runArtifactPath string
	if p := fields["latest_cook_run"]; p != "" {
		if slug, createdAt, ok := parseArtifactFilename(p); ok {
			runSlug, runCreatedAt, runArtifactPath = slug, createdAt, p
		}
	}

	// Compute every slug's status up front so a slug playing multiple
	// roles (e.g. current_phase == the run's phase) still lands on the
	// correct status regardless of processing order.
	assignments := map[string]string{}
	if runSlug != "" {
		assignments[runSlug] = domain.StoryInProgress
	}
	if currentPhase != "" {
		if _, ok := assignments[currentPhase]; !ok {
			assignments[currentPhase] = domain.StoryPlanned
		}
	}
	if entryPhase != "" {
		if _, ok := assignments[entryPhase]; !ok {
			assignments[entryPhase] = domain.StoryPlanned
		}
	}

	slugs := make([]string, 0, len(assignments))
	for slug := range assignments {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		changed, err := ensureStory(db, slug, assignments[slug], roadmapPath, now)
		if err != nil {
			return result, fmt.Errorf("ensure story %s: %w", slug, err)
		}
		recordOutcome(&result, changed)
	}

	var runID string
	if runSlug != "" && runArtifactPath != "" {
		id, changed, err := ensureRun(db, runSlug, runArtifactPath, runCreatedAt)
		if err != nil {
			return result, fmt.Errorf("ensure run: %w", err)
		}
		runID = id
		recordOutcome(&result, changed)
	}

	changed, err := ensureMeta(db, currentPhase, entryPhase, runID)
	if err != nil {
		return result, fmt.Errorf("ensure meta: %w", err)
	}
	recordOutcome(&result, changed)

	return result, nil
}

func recordOutcome(result *ImportResult, changed bool) {
	if changed {
		result.Imported++
		return
	}
	result.Skipped++
}

func storyByExactSlug(db *sql.DB, slug string) (id, status string, exists bool, err error) {
	err = db.QueryRow(`SELECT id, status FROM stories WHERE slug = ?`, slug).Scan(&id, &status)
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("query story %q: %w", slug, err)
	}
	return id, status, true, nil
}

// ensureStory creates or updates the `stories` row for slug so it matches
// wantStatus, writing to the db only when something actually changes.
func ensureStory(db *sql.DB, slug, wantStatus, roadmapPath, at string) (changed bool, err error) {
	id, existingStatus, exists, err := storyByExactSlug(db, slug)
	if err != nil {
		return false, err
	}

	if exists {
		if existingStatus == wantStatus {
			return false, nil
		}
		// Markdown is the write target for the story's phase-block status,
		// same as story create/run create/check record/handoff (R8, P3
		// wave 1, docs/plans/active/harness-markdown-truth.md) — a no-op
		// when slug has no matching phase block, which is the common case
		// for a legacy import predating this plan format.
		writeStatus, err := preparePlanPhaseStatus(db, slug, wantStatus)
		if err != nil {
			return false, err
		}
		if err := writeStatus(); err != nil {
			return false, fmt.Errorf("plan write failed: %w", err)
		}
		if _, err := db.Exec(`UPDATE stories SET status = ? WHERE id = ?`, wantStatus, id); err != nil {
			return false, fmt.Errorf("update story %s: %w", slug, err)
		}
		return true, nil
	}

	goal, ok := parseRoadmapGoal(roadmapPath, slug)
	if !ok {
		goal = fmt.Sprintf("imported from legacy workflow-state.yml (no roadmap goal found for %s)", slug)
	}

	writeStatus, err := preparePlanPhaseStatus(db, slug, wantStatus)
	if err != nil {
		return false, err
	}
	if err := writeStatus(); err != nil {
		return false, fmt.Errorf("plan write failed: %w", err)
	}

	if _, err := db.Exec(
		`INSERT INTO stories (id, slug, goal, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		ulid.Make().String(), slug, goal, wantStatus, at,
	); err != nil {
		return false, fmt.Errorf("insert story %s: %w", slug, err)
	}
	return true, nil
}

func runByArtifactPath(db *sql.DB, artifactPath string) (id string, exists bool, err error) {
	err = db.QueryRow(`SELECT id FROM runs WHERE artifact_path = ?`, artifactPath).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("query run by artifact_path %q: %w", artifactPath, err)
	}
	return id, true, nil
}

// ensureRun creates the `runs` row for the legacy `latest_cook_run`
// pointer, keyed by artifact_path (there's no other stable natural key
// available from a bare legacy path string).
func ensureRun(db *sql.DB, storySlug, artifactPath, createdAt string) (id string, changed bool, err error) {
	existingID, exists, err := runByArtifactPath(db, artifactPath)
	if err != nil {
		return "", false, err
	}
	if exists {
		return existingID, false, nil
	}

	newID := ulid.Make().String()
	if _, err := db.Exec(
		`INSERT INTO runs (id, story_slug, artifact_path, created_at) VALUES (?, ?, ?, ?)`,
		newID, storySlug, artifactPath, createdAt,
	); err != nil {
		return "", false, fmt.Errorf("insert run: %w", err)
	}
	return newID, true, nil
}

// ensureMeta updates only the meta columns that actually differ from
// current DB state. latest_check_id is never touched here — see Import's
// doc comment for why checks are out of scope for this import.
func ensureMeta(db *sql.DB, currentPhase, entryPhase, runID string) (changed bool, err error) {
	var existingCurrent, existingEntry, existingRunID sql.NullString
	err = db.QueryRow(`SELECT current_phase, entry_phase, latest_run_id FROM meta LIMIT 1`).
		Scan(&existingCurrent, &existingEntry, &existingRunID)
	if err != nil {
		return false, fmt.Errorf("read meta: %w", err)
	}

	newCurrent, newEntry, newRunID := existingCurrent.String, existingEntry.String, existingRunID.String
	changed = false
	if currentPhase != "" && existingCurrent.String != currentPhase {
		newCurrent = currentPhase
		changed = true
	}
	if entryPhase != "" && existingEntry.String != entryPhase {
		newEntry = entryPhase
		changed = true
	}
	if runID != "" && existingRunID.String != runID {
		newRunID = runID
		changed = true
	}
	if !changed {
		return false, nil
	}

	if _, err := db.Exec(
		`UPDATE meta SET current_phase = ?, entry_phase = ?, latest_run_id = ?`,
		newCurrent, newEntry, newRunID,
	); err != nil {
		return false, fmt.Errorf("update meta: %w", err)
	}
	return true, nil
}

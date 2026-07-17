package application

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// ImportResult mirrors CONTRACT.md's `import --json` shape.
type ImportResult struct {
	Imported          int      `json:"imported"`
	Skipped           int      `json:"skipped"`
	ChangesetsWritten []string `json:"changesets_written"`
}

// Import parses legacy `{legacyDir}/workflow-state.yml` (plus the
// planning markdown it points to) into changesets applied to db, per
// STATE.md's Legacy Field Mapping. It only creates `stories` rows for the
// slugs actually required by the FK graph (current_phase, entry_phase,
// and the phase embedded in latest_cook_run's filename) — not a full
// historical import of every roadmap phase. It deliberately never
// synthesizes a `checks` row for latest_check_report: checks.verdict is
// NOT NULL and STATE.md's mapping only covers yml fields, never
// check-report body parsing, so meta.latest_check_id is left NULL.
//
// Idempotent by pre-check, not just by the changeset engine's own fence:
// each entity is compared against current DB state before writing
// anything, so a second run with unchanged legacy input produces zero
// new changesets.
func Import(db *sql.DB, legacyDir, changesetDir string) (ImportResult, error) {
	result := ImportResult{ChangesetsWritten: []string{}}

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
		changed, path, err := ensureStory(db, changesetDir, slug, assignments[slug], roadmapPath, now)
		if err != nil {
			return result, fmt.Errorf("ensure story %s: %w", slug, err)
		}
		recordOutcome(&result, changed, path)
	}

	var runID string
	if runSlug != "" && runArtifactPath != "" {
		id, changed, path, err := ensureRun(db, changesetDir, runSlug, runArtifactPath, runCreatedAt)
		if err != nil {
			return result, fmt.Errorf("ensure run: %w", err)
		}
		runID = id
		recordOutcome(&result, changed, path)
	}

	changed, path, err := ensureMeta(db, changesetDir, currentPhase, entryPhase, runID, now)
	if err != nil {
		return result, fmt.Errorf("ensure meta: %w", err)
	}
	recordOutcome(&result, changed, path)

	return result, nil
}

func recordOutcome(result *ImportResult, changed bool, path string) {
	if changed {
		result.Imported++
		result.ChangesetsWritten = append(result.ChangesetsWritten, path)
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
// wantStatus, writing a changeset only when something actually changes.
func ensureStory(db *sql.DB, changesetDir, slug, wantStatus, roadmapPath, at string) (changed bool, path string, err error) {
	id, existingStatus, exists, err := storyByExactSlug(db, slug)
	if err != nil {
		return false, "", err
	}

	if exists {
		if existingStatus == wantStatus {
			return false, "", nil
		}
		path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
			{Op: "update", Entity: "story", ID: id, Fields: map[string]any{"status": wantStatus}, At: at},
		})
		return true, path, err
	}

	goal, ok := parseRoadmapGoal(roadmapPath, slug)
	if !ok {
		goal = fmt.Sprintf("imported from legacy workflow-state.yml (no roadmap goal found for %s)", slug)
	}

	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{
			Op:     "create",
			Entity: "story",
			ID:     ulid.Make().String(),
			Fields: map[string]any{
				"slug":       slug,
				"goal":       goal,
				"status":     wantStatus,
				"created_at": at,
			},
			At: at,
		},
	})
	return true, path, err
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
func ensureRun(db *sql.DB, changesetDir, storySlug, artifactPath, createdAt string) (id string, changed bool, path string, err error) {
	existingID, exists, err := runByArtifactPath(db, artifactPath)
	if err != nil {
		return "", false, "", err
	}
	if exists {
		return existingID, false, "", nil
	}

	newID := ulid.Make().String()
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{
			Op:     "create",
			Entity: "run",
			ID:     newID,
			Fields: map[string]any{
				"story_slug":    storySlug,
				"artifact_path": artifactPath,
				"created_at":    createdAt,
			},
			At: createdAt,
		},
	})
	if err != nil {
		return "", false, "", err
	}
	return newID, true, path, nil
}

// ensureMeta updates only the meta columns that actually differ from
// current DB state. latest_check_id is never touched here — see Import's
// doc comment for why checks are out of scope for this import.
func ensureMeta(db *sql.DB, changesetDir, currentPhase, entryPhase, runID, at string) (changed bool, path string, err error) {
	var existingCurrent, existingEntry, existingRunID sql.NullString
	err = db.QueryRow(`SELECT current_phase, entry_phase, latest_run_id FROM meta LIMIT 1`).
		Scan(&existingCurrent, &existingEntry, &existingRunID)
	if err != nil {
		return false, "", fmt.Errorf("read meta: %w", err)
	}

	fields := map[string]any{}
	if currentPhase != "" && existingCurrent.String != currentPhase {
		fields["current_phase"] = currentPhase
	}
	if entryPhase != "" && existingEntry.String != entryPhase {
		fields["entry_phase"] = entryPhase
	}
	if runID != "" && existingRunID.String != runID {
		fields["latest_run_id"] = runID
	}
	if len(fields) == 0 {
		return false, "", nil
	}

	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "update", Entity: "meta", Fields: fields, At: at},
	})
	return true, path, err
}

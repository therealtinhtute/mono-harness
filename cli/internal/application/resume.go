package application

import (
	"database/sql"
	"fmt"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// Position mirrors CONTRACT.md's `resume --json` position sub-object: the
// current phase (story slug) and that story's status, if any.
type Position struct {
	CurrentPhase *string `json:"current_phase"`
	Status       *string `json:"status"`
}

// DriftFinding is one entry of `resume --json`'s drift array.
type DriftFinding struct {
	Type     string `json:"type"`
	Detail   string `json:"detail"`
	Recovery string `json:"recovery"`
}

// StaleDocsRecovery is the single source of truth for the stale_docs
// recovery instruction — STATE.md's stale-pointer table quotes this
// constant verbatim rather than duplicating the string (the #24 lesson).
const StaleDocsRecovery = "zharness init --refresh-docs"

// docsVersionMinSchema is the schema_version that introduced
// meta.docs_version (infrastructure migration 0002_meta_docs_version).
// Below this, the column doesn't exist yet — an un-migrated project is
// unversioned, not an error.
const docsVersionMinSchema = 2

// ResumeView mirrors CONTRACT.md's locked `resume --json` shape.
type ResumeView struct {
	Position        Position       `json:"position"`
	LatestRunID     *string        `json:"latest_run_id"`
	LatestCheckID   *string        `json:"latest_check_id"`
	LatestHandoffID *string        `json:"latest_handoff_id"`
	Drift           []DriftFinding `json:"drift"`
	Readiness       string         `json:"readiness"`
}

func runArtifactPathByID(db *sql.DB, id string) (path string, exists bool, err error) {
	err = db.QueryRow(`SELECT artifact_path FROM runs WHERE id = ?`, id).Scan(&path)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("query run %q: %w", id, err)
	}
	return path, true, nil
}

func checkRunIDByID(db *sql.DB, id string) (runID string, exists bool, err error) {
	err = db.QueryRow(`SELECT run_id FROM checks WHERE id = ?`, id).Scan(&runID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("query check %q: %w", id, err)
	}
	return runID, true, nil
}

func latestHandoffID(db *sql.DB) (id string, exists bool, err error) {
	err = db.QueryRow(`SELECT id FROM handoffs ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("query latest handoff: %w", err)
	}
	return id, true, nil
}

// Resume derives CONTRACT.md's `resume` snapshot from the current meta
// pointers plus lightweight cross-checks against stories/runs/checks. It
// only reads `meta`'s pointer columns — nothing here creates or updates
// them (see cli-domain's implementation-notes.md entry on the
// meta-pointer-maintenance gap: no command in this phase's 19-command
// surface writes latest_run_id/latest_check_id going forward, only the
// existing `import`). cliVersion is the running binary's own version
// (`"dev"` for unreleased builds) — compared against meta.docs_version to
// detect stale_docs drift; resume never writes it back.
func Resume(db *sql.DB, cliVersion string) (ResumeView, error) {
	view := ResumeView{Drift: []DriftFinding{}}

	state, err := QueryState(db)
	if err != nil {
		return view, fmt.Errorf("resume: %w", err)
	}
	view.LatestRunID = state.LatestRunID
	view.LatestCheckID = state.LatestCheckID
	view.Position.CurrentPhase = state.CurrentPhase

	if state.CurrentPhase != nil {
		_, status, exists, err := storyByExactSlug(db, *state.CurrentPhase)
		if err != nil {
			return view, fmt.Errorf("resume: %w", err)
		}
		if !exists {
			view.Drift = append(view.Drift, DriftFinding{
				Type:     "unknown_phase",
				Detail:   fmt.Sprintf("current_phase %q has no matching story", *state.CurrentPhase),
				Recovery: fmt.Sprintf("record it via `zharness story --slug %s --goal ...`, or correct workflow-state's current_phase", *state.CurrentPhase),
			})
		} else {
			s := status
			view.Position.Status = &s
		}
	}

	if id, exists, err := latestHandoffID(db); err != nil {
		return view, fmt.Errorf("resume: %w", err)
	} else if exists {
		view.LatestHandoffID = &id
	}

	if state.LatestRunID != nil {
		artifactPath, exists, err := runArtifactPathByID(db, *state.LatestRunID)
		if err != nil {
			return view, fmt.Errorf("resume: %w", err)
		}
		if exists && artifactPath != "" && !infrastructure.Exists(artifactPath) {
			view.Drift = append(view.Drift, DriftFinding{
				Type:     "missing_file",
				Detail:   fmt.Sprintf("run %s artifact_path %q not found on disk", *state.LatestRunID, artifactPath),
				Recovery: "re-run `work` for the current phase, or correct the run's artifact_path",
			})
		}
	}

	if state.LatestCheckID != nil {
		runID, exists, err := checkRunIDByID(db, *state.LatestCheckID)
		if err != nil {
			return view, fmt.Errorf("resume: %w", err)
		}
		if exists && state.LatestRunID != nil && runID != *state.LatestRunID {
			view.Drift = append(view.Drift, DriftFinding{
				Type:     "out_of_order",
				Detail:   fmt.Sprintf("latest_check %s belongs to run %s, but latest_run_id is %s", *state.LatestCheckID, runID, *state.LatestRunID),
				Recovery: "record a new check for the latest run via `check record`, or correct latest_run_id/latest_check_id",
			})
		}
	}

	if state.SchemaVersion >= docsVersionMinSchema {
		var writtenVersion sql.NullString
		if err := db.QueryRow(`SELECT docs_version FROM meta LIMIT 1`).Scan(&writtenVersion); err != nil {
			return view, fmt.Errorf("resume: read meta.docs_version: %w", err)
		}
		if written := writtenVersion.String; writtenVersion.Valid && written != "" &&
			written != "dev" && cliVersion != "dev" && written != cliVersion {
			view.Drift = append(view.Drift, DriftFinding{
				Type:     "stale_docs",
				Detail:   fmt.Sprintf("docs written at version %q, CLI is version %q", written, cliVersion),
				Recovery: StaleDocsRecovery,
			})
		}
	}

	switch {
	case len(view.Drift) > 0:
		view.Readiness = "drifted"
	case view.Position.Status != nil && (*view.Position.Status == domain.StoryInProgress || *view.Position.Status == domain.StoryPlanned):
		view.Readiness = "in-progress"
	default:
		view.Readiness = "clean"
	}

	return view, nil
}

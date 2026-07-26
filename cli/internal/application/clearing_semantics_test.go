package application

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/embedded"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// TestClearingSemantics_RefreshDocsResolvesStaleDocsDrift proves
// cli-stale-drift-PLAN.md's T3 acceptance: a project stamped with an older
// docs_version reports stale_docs via Resume; running the named recovery
// (`init --refresh-docs`, exercised here as ScaffoldDocs(refresh=true))
// clears the drift; and lifecycle rows/pointers remain byte-stable while
// managed root docs and their hash records advance.
func TestClearingSemantics_RefreshDocsResolvesStaleDocsDrift(t *testing.T) {
	db, changesetDir := freshDB(t)
	root := t.TempDir()
	kitDir := filepath.Join(root, ".kit")

	// Seed non-docs state that must survive the refresh untouched. Uses a
	// real, existing artifact_path (unlike the shared seedRun fixture's
	// fake path) so the only drift this test can observe is stale_docs —
	// no incidental missing_file noise.
	runArtifact := filepath.Join(root, "run.md")
	if err := os.WriteFile(runArtifact, []byte("run log"), 0o644); err != nil {
		t.Fatalf("write run artifact fixture: %v", err)
	}
	at := "2026-07-18T12:00:00Z"
	storyID := ulid.Make().String()
	runID := ulid.Make().String()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "story", ID: storyID, Fields: map[string]any{
			"slug": "cli-domain", "goal": "test fixture", "status": "done", "created_at": at,
		}, At: at},
		{Op: "create", Entity: "run", ID: runID, Fields: map[string]any{
			"story_slug": "cli-domain", "artifact_path": runArtifact, "created_at": at,
		}, At: at},
	}); err != nil {
		t.Fatalf("seed story + run: %v", err)
	}
	setMeta(t, db, changesetDir, map[string]any{
		"current_phase": "cli-domain", "entry_phase": "cli-domain", "latest_run_id": runID,
	})

	if _, err := ScaffoldDocs(db, changesetDir, root, kitDir, embedded.FS, "0.2.0", false, false); err != nil {
		t.Fatalf("initial ScaffoldDocs: %v", err)
	}

	// Simulate upgrading the CLI to 0.3.0 without refreshing this project's docs.
	before, err := Resume(db, "0.3.0")
	if err != nil {
		t.Fatalf("Resume (before refresh): %v", err)
	}
	if before.Readiness != "drifted" {
		t.Fatalf("readiness before refresh = %q, want drifted", before.Readiness)
	}
	foundStale := false
	for _, d := range before.Drift {
		if d.Type == "stale_docs" {
			foundStale = true
			if d.Recovery != StaleDocsRecovery {
				t.Errorf("recovery = %q, want %q", d.Recovery, StaleDocsRecovery)
			}
		}
	}
	if !foundStale {
		t.Fatalf("drift before refresh = %v, want a stale_docs finding", before.Drift)
	}

	preRefreshStoryStatus := queryStoryStatus(t, db, "cli-domain")
	preRefreshRunArtifact := queryRunArtifactPath(t, db, runID)

	if _, err := ScaffoldDocs(db, changesetDir, root, kitDir, embedded.FS, "0.3.0", true, false); err != nil {
		t.Fatalf("refresh ScaffoldDocs: %v", err)
	}

	after, err := Resume(db, "0.3.0")
	if err != nil {
		t.Fatalf("Resume (after refresh): %v", err)
	}
	if after.Readiness != "clean" {
		t.Fatalf("readiness after refresh = %q, want clean; drift = %v", after.Readiness, after.Drift)
	}
	if len(after.Drift) != 0 {
		t.Fatalf("drift after refresh = %v, want none", after.Drift)
	}

	// Managed docs/hash metadata may change; stories/runs/meta pointers stay stable.
	var currentPhase, entryPhase, latestRunID string
	if err := db.QueryRow(`SELECT current_phase, entry_phase, latest_run_id FROM meta`).Scan(&currentPhase, &entryPhase, &latestRunID); err != nil {
		t.Fatalf("query meta pointers: %v", err)
	}
	if currentPhase != "cli-domain" || entryPhase != "cli-domain" || latestRunID != runID {
		t.Errorf("meta pointers after refresh = (%q, %q, %q), want unchanged (%q, %q, %q)", currentPhase, entryPhase, latestRunID, "cli-domain", "cli-domain", runID)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != preRefreshStoryStatus {
		t.Errorf("story status after refresh = %q, want unchanged %q", got, preRefreshStoryStatus)
	}
	if got := queryRunArtifactPath(t, db, runID); got != preRefreshRunArtifact {
		t.Errorf("run artifact_path after refresh = %q, want unchanged %q", got, preRefreshRunArtifact)
	}

	var docsVersion string
	if err := db.QueryRow(`SELECT docs_version FROM meta`).Scan(&docsVersion); err != nil {
		t.Fatalf("query meta.docs_version: %v", err)
	}
	if docsVersion != "0.3.0" {
		t.Errorf("meta.docs_version after refresh = %q, want %q", docsVersion, "0.3.0")
	}

	if _, err := os.Stat(filepath.Join(root, "docs", "WORKFLOW.md")); err != nil {
		t.Errorf("docs/WORKFLOW.md missing after refresh: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("root AGENTS.md missing after refresh: %v", err)
	}
}

func queryStoryStatus(t *testing.T, db *sql.DB, slug string) string {
	t.Helper()
	var status string
	if err := db.QueryRow(`SELECT status FROM stories WHERE slug = ?`, slug).Scan(&status); err != nil {
		t.Fatalf("query story %q status: %v", slug, err)
	}
	return status
}

func queryRunArtifactPath(t *testing.T, db *sql.DB, runID string) string {
	t.Helper()
	var path string
	if err := db.QueryRow(`SELECT artifact_path FROM runs WHERE id = ?`, runID).Scan(&path); err != nil {
		t.Fatalf("query run %q artifact_path: %v", runID, err)
	}
	return path
}

package application

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// TestRunCreateReplaySafety proves run create's three-line changeset (run
// row + story status + meta.latest_run_id pointer) replays identically: build state
// incrementally against db1 as the commands are issued, then rebuild db2
// purely by replaying the changeset directory from empty, and assert
// Resume sees the same position/pointers either way.
func TestRunCreateReplaySafety(t *testing.T) {
	root := t.TempDir()
	changesetDir := filepath.Join(root, ".kit", "changesets")

	db1, err := infrastructure.Open(filepath.Join(root, "db1.sqlite"))
	if err != nil {
		t.Fatalf("open db1: %v", err)
	}
	defer db1.Close()
	if _, _, err := infrastructure.Migrate(db1); err != nil {
		t.Fatalf("migrate db1: %v", err)
	}

	if _, _, err := CreateStory(db1, changesetDir, "write-boundary", "add run create", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, _, err := CreateRun(db1, changesetDir, "write-boundary", ".kit/runs/work/x.md", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	db2, err := infrastructure.Open(filepath.Join(root, "db2.sqlite"))
	if err != nil {
		t.Fatalf("open db2: %v", err)
	}
	defer db2.Close()
	if _, _, err := infrastructure.Migrate(db2); err != nil {
		t.Fatalf("migrate db2: %v", err)
	}
	totalApplied, err := infrastructure.Replay(db2, changesetDir)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if totalApplied != 4 { // story create + run create + story status update + meta update
		t.Fatalf("Replay totalApplied = %d, want 4", totalApplied)
	}

	view1, err := Resume(db1, "dev")
	if err != nil {
		t.Fatalf("Resume(db1): %v", err)
	}
	view2, err := Resume(db2, "dev")
	if err != nil {
		t.Fatalf("Resume(db2): %v", err)
	}

	json1, err := json.Marshal(view1)
	if err != nil {
		t.Fatalf("marshal view1: %v", err)
	}
	json2, err := json.Marshal(view2)
	if err != nil {
		t.Fatalf("marshal view2: %v", err)
	}
	if string(json1) != string(json2) {
		t.Fatalf("resume view diverged after replay:\nincremental: %s\nreplayed:    %s", json1, json2)
	}

	if view2.LatestRunID == nil || *view2.LatestRunID != runID {
		t.Fatalf("replayed latest_run_id = %v, want %s", view2.LatestRunID, runID)
	}
}

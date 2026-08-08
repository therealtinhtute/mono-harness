package application

import (
	"testing"

	"github.com/oklog/ulid/v2"
)

func TestQueryArtifactsPreservesLegacyArtifactPaths(t *testing.T) {
	db, changesetDir := freshDB(t)
	firstID := seedRun(t, db, changesetDir)
	if _, err := db.Exec(`UPDATE runs SET artifact_path = '' WHERE id = ?`, firstID); err != nil {
		t.Fatalf("clear artifact_path: %v", err)
	}

	missingPath := ".kit/runs/work/nonexistent.md"
	secondID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO runs (id, story_slug, trace_ids, artifact_path, created_at)
		VALUES (?, 'cli-domain', '[]', ?, '2026-07-27T12:01:00Z')
	`, secondID, missingPath); err != nil {
		t.Fatalf("insert second run: %v", err)
	}

	artifacts, err := QueryArtifacts(db, "cli-domain")
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("artifacts = %+v, want two typed run rows", artifacts)
	}
	if artifacts[0].ID != firstID || artifacts[0].ArtifactPath != "" {
		t.Fatalf("first artifact = %+v, want empty artifact_path preserved", artifacts[0])
	}
	if artifacts[1].ID != secondID || artifacts[1].ArtifactPath != missingPath {
		t.Fatalf("second artifact = %+v, want nonexistent artifact_path preserved", artifacts[1])
	}
}

func TestQueryTracesChronologicalOrderAndFilters(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	for i, at := range []string{"2026-07-27T12:00:00Z", "2026-07-27T12:01:00Z", "2026-07-27T12:02:00Z"} {
		id, _, err := CreateTrace(db, changesetDir, i+1, "wave summary", runID, "", "")
		if err != nil {
			t.Fatalf("CreateTrace(%d): %v", i, err)
		}
		if _, err := db.Exec(`UPDATE traces SET created_at = ? WHERE id = ?`, at, id); err != nil {
			t.Fatalf("backdate trace %d: %v", i, err)
		}
	}

	all, err := QueryTraces(db, "", 0)
	if err != nil {
		t.Fatalf("QueryTraces: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("traces = %+v, want 3 rows", all)
	}
	for i, want := range []int{1, 2, 3} {
		if all[i].Wave != want {
			t.Fatalf("traces[%d].Wave = %d, want %d (chronological order)", i, all[i].Wave, want)
		}
	}
	if all[0].RunID == nil || *all[0].RunID != runID {
		t.Fatalf("traces[0].RunID = %v, want %s", all[0].RunID, runID)
	}

	tailed, err := QueryTraces(db, "", 2)
	if err != nil {
		t.Fatalf("QueryTraces tail=2: %v", err)
	}
	if len(tailed) != 2 || tailed[0].Wave != 2 || tailed[1].Wave != 3 {
		t.Fatalf("tailed traces = %+v, want waves [2 3]", tailed)
	}

	filtered, err := QueryTraces(db, runID, 0)
	if err != nil {
		t.Fatalf("QueryTraces run-id filter: %v", err)
	}
	if len(filtered) != 3 {
		t.Fatalf("filtered traces = %+v, want 3 rows for the seeded run", filtered)
	}

	none, err := QueryTraces(db, "01ARZ3NDEKTSV4RRFFQ69G5FAV", 0)
	if err != nil {
		t.Fatalf("QueryTraces unknown run-id: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("traces for unknown run = %+v, want empty", none)
	}
}

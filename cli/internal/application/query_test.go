package application

import (
	"fmt"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
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

// TestQueryChecksChronologicalOrderAndFilters is P6's own success-signal
// fix: `query check --latest` only ever exposed the most recent verdict,
// leaving Validation's append-only history unqueryable — the one
// append-only section that was still write-only. QueryChecks closes that
// gap with the same phase/tail filter shape as QueryTraces/QueryDecisions.
func TestQueryChecksChronologicalOrderAndFilters(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")

	var ids []string
	for i, verdict := range []string{"REQUEST_CHANGES", "APPROVED"} {
		id, _, err := RecordCheck(db, changesetDir, runID, verdict, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true", OutputRef: "x"}})
		if err != nil {
			t.Fatalf("RecordCheck(%d): %v", i, err)
		}
		if _, err := db.Exec(`UPDATE checks SET created_at = ? WHERE id = ?`, fmt.Sprintf("2026-07-27T12:0%d:00Z", i), id); err != nil {
			t.Fatalf("backdate check %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	all, err := QueryChecks(db, "", 0)
	if err != nil {
		t.Fatalf("QueryChecks: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("checks = %+v, want 2 rows", all)
	}
	if all[0].ID != ids[0] || all[1].ID != ids[1] {
		t.Fatalf("checks order = [%s %s], want chronological %v", all[0].ID, all[1].ID, ids)
	}
	if all[0].Verdict != "REQUEST_CHANGES" || all[1].Verdict != "APPROVED" {
		t.Fatalf("checks verdicts = [%s %s], want [REQUEST_CHANGES APPROVED]", all[0].Verdict, all[1].Verdict)
	}
	if all[0].Phase != "cli-domain" || all[0].RunID != runID {
		t.Fatalf("checks[0] phase/run = %s/%s, want cli-domain/%s", all[0].Phase, all[0].RunID, runID)
	}
	if all[0].Judge == nil || *all[0].Judge != domain.JudgeIndependent {
		t.Fatalf("checks[0].Judge = %v, want %s", all[0].Judge, domain.JudgeIndependent)
	}

	tailed, err := QueryChecks(db, "", 1)
	if err != nil {
		t.Fatalf("QueryChecks tail=1: %v", err)
	}
	if len(tailed) != 1 || tailed[0].ID != ids[1] {
		t.Fatalf("tailed checks = %+v, want only the most recent check", tailed)
	}

	filtered, err := QueryChecks(db, "cli-domain", 0)
	if err != nil {
		t.Fatalf("QueryChecks phase filter: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("filtered checks = %+v, want 2 rows for cli-domain", filtered)
	}

	none, err := QueryChecks(db, "no-such-phase", 0)
	if err != nil {
		t.Fatalf("QueryChecks unknown phase: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("checks for unknown phase = %+v, want empty", none)
	}
}

package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCheckRecord(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	id, path, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, []domain.ProofLink{
		{Command: "go test ./...", OutputRef: "ok", ArtifactPath: ".kit/runs/work/x.md"},
	})
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if path == "" {
		t.Fatal("changeset path is empty, want a written .jsonl file")
	}

	var verdict, proofLinks string
	if err := db.QueryRow(`SELECT verdict, proof_links FROM checks WHERE id = ?`, id).Scan(&verdict, &proofLinks); err != nil {
		t.Fatalf("query check: %v", err)
	}
	if verdict != domain.VerdictApproved {
		t.Fatalf("verdict = %q, want %q", verdict, domain.VerdictApproved)
	}
	if proofLinks == "[]" || proofLinks == "" {
		t.Fatalf("proof_links = %q, want a non-empty JSON array", proofLinks)
	}

	var latestCheckID string
	if err := db.QueryRow(`SELECT latest_check_id FROM meta LIMIT 1`).Scan(&latestCheckID); err != nil {
		t.Fatalf("query meta: %v", err)
	}
	if latestCheckID != id {
		t.Fatalf("meta.latest_check_id = %q, want %q", latestCheckID, id)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryChecked {
		t.Fatalf("story status = %q, want checked", got)
	}
}

func TestCheckRecordRequestChangesAllowsEmptyProofLinks(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictRequestChanges, nil)
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	if got := queryStoryStatus(t, db, "cli-domain"); got != domain.StoryInProgress {
		t.Fatalf("story status = %q, want in-progress after REQUEST_CHANGES", got)
	}
}

func TestCheckRecordEmptyProofLinks(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, nil)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "empty_proof_links" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: empty_proof_links}", err)
	}
}

func TestCheckRecordInvalidVerdict(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)

	_, _, err := RecordCheck(db, changesetDir, runID, "MAYBE", []domain.ProofLink{{Command: "x"}})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_verdict" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_verdict}", err)
	}
}

func TestCheckRecordUnknownRunID(t *testing.T) {
	db, changesetDir := freshDB(t)

	_, _, err := RecordCheck(db, changesetDir, "01HZZZZZZZZZZZZZZZZZZZZZZZ", domain.VerdictApproved, []domain.ProofLink{{Command: "x"}})
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_run_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_run_id}", err)
	}
	if got := countRows(t, db, "checks"); got != 0 {
		t.Fatalf("checks rows = %d, want 0", got)
	}
}

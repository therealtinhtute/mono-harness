package application

import (
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestScoreTraceMinimal(t *testing.T) {
	db, changesetDir := freshDB(t)
	id, _, err := CreateTrace(db, changesetDir, 1, "did stuff", "")
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}

	score, err := ScoreTrace(db, id)
	if err != nil {
		t.Fatalf("ScoreTrace: %v", err)
	}
	if score.Tier != "minimal" {
		t.Fatalf("tier = %q, want minimal (no run_id link)", score.Tier)
	}
	if len(score.Reasons) == 0 {
		t.Fatal("reasons is empty, want at least one")
	}
}

func TestScoreTraceStandard(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	id, _, err := CreateTrace(db, changesetDir, 1, "wave 1 completed cleanly", runID)
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}

	score, err := ScoreTrace(db, id)
	if err != nil {
		t.Fatalf("ScoreTrace: %v", err)
	}
	if score.Tier != "standard" {
		t.Fatalf("tier = %q, want standard (summary >= 10 chars, linked to run)", score.Tier)
	}
}

func TestScoreTraceDetailed(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := seedRun(t, db, changesetDir)
	longSummary := "wave 1 completed with full detail across every planned task and verification step"
	if len(longSummary) < 40 {
		t.Fatalf("test setup: longSummary must be >= 40 chars, got %d", len(longSummary))
	}
	if _, _, err := CreateTrace(db, changesetDir, 1, longSummary, runID); err != nil {
		t.Fatalf("CreateTrace (first): %v", err)
	}
	id, _, err := CreateTrace(db, changesetDir, 2, longSummary, runID)
	if err != nil {
		t.Fatalf("CreateTrace (second): %v", err)
	}

	score, err := ScoreTrace(db, id)
	if err != nil {
		t.Fatalf("ScoreTrace: %v", err)
	}
	if score.Tier != "detailed" {
		t.Fatalf("tier = %q, want detailed (summary >= 40 chars, >1 trace on run)", score.Tier)
	}
}

func TestScoreTraceUnknownID(t *testing.T) {
	db, _ := freshDB(t)
	_, err := ScoreTrace(db, "01HZZZZZZZZZZZZZZZZZZZZZZZ")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_trace_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_trace_id}", err)
	}
}

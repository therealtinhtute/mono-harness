package infrastructure_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// TestTraceAddBatchReplayEquivalence is R5's replay-equivalence proof
// (docs/audit/sdlc-token-cache-audit.md, p3-fewer-round-trips wave 1 task
// 2): batching `trace add --tasks` must not change what ends up in the
// index. Two things have to hold: (1) a batched write and the same content
// written one call at a time produce rows equal in every substantive field
// (wave/task/task_status/summary) — the batch is a transport optimization,
// not a different code path with different semantics; (2) a DB rebuilt
// purely by replaying the changeset directory from empty agrees exactly
// with the DB the commands built incrementally, including the multi-line
// changeset a batch call writes in one file (run_create_replay_test.go's
// established pattern, extended to a multi-entity writer).
func TestTraceAddBatchReplayEquivalence(t *testing.T) {
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

	if _, _, err := application.CreateStory(db1, changesetDir, "replay-phase", "prove batch replay", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, _, err := application.CreateRun(db1, changesetDir, "replay-phase", ".kit/runs/work/x.md", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	if _, _, err := application.CreateTraces(db1, changesetDir, 1, runID, []domain.TraceTask{
		{Task: "task A", TaskStatus: domain.TaskStatusDone, Summary: "did A"},
		{Task: "task B", TaskStatus: domain.TaskStatusDoneWithConcerns, Summary: "did B"},
	}); err != nil {
		t.Fatalf("CreateTraces (batch): %v", err)
	}
	if _, _, err := application.CreateTrace(db1, changesetDir, 1, "did A", runID, "task A", domain.TaskStatusDone); err != nil {
		t.Fatalf("CreateTrace (per-call A): %v", err)
	}
	if _, _, err := application.CreateTrace(db1, changesetDir, 1, "did B", runID, "task B", domain.TaskStatusDoneWithConcerns); err != nil {
		t.Fatalf("CreateTrace (per-call B): %v", err)
	}

	traces, err := application.QueryTracesByPhase(db1, "replay-phase", 0)
	if err != nil {
		t.Fatalf("QueryTracesByPhase: %v", err)
	}
	if len(traces) != 4 {
		t.Fatalf("QueryTracesByPhase = %d rows, want 4 (2 batched + 2 per-call)", len(traces))
	}

	byTask := map[string][]application.TraceView{}
	for _, tr := range traces {
		if tr.Task == nil {
			t.Fatalf("trace row missing task: %+v", tr)
		}
		byTask[*tr.Task] = append(byTask[*tr.Task], tr)
	}
	for _, task := range []string{"task A", "task B"} {
		pair := byTask[task]
		if len(pair) != 2 {
			t.Fatalf("task %q has %d rows, want 2 (one batched, one per-call)", task, len(pair))
		}
		a, b := pair[0], pair[1]
		if a.ID == b.ID {
			t.Fatalf("task %q: batch and per-call rows share an id: %s", task, a.ID)
		}
		if a.Wave != b.Wave || a.Summary != b.Summary || a.TaskStatus == nil || b.TaskStatus == nil || *a.TaskStatus != *b.TaskStatus {
			t.Fatalf("task %q: batch vs per-call rows diverge on substantive fields: %+v vs %+v", task, a, b)
		}
		if a.RunID == nil || b.RunID == nil || *a.RunID != runID || *b.RunID != runID {
			t.Fatalf("task %q: run_id mismatch: %+v vs %+v, want both %s", task, a.RunID, b.RunID, runID)
		}
	}

	db2, err := infrastructure.Open(filepath.Join(root, "db2.sqlite"))
	if err != nil {
		t.Fatalf("open db2: %v", err)
	}
	defer db2.Close()
	if _, _, err := infrastructure.Migrate(db2); err != nil {
		t.Fatalf("migrate db2: %v", err)
	}
	if _, err := infrastructure.Replay(db2, changesetDir); err != nil {
		t.Fatalf("Replay: %v", err)
	}

	replayed, err := application.QueryTracesByPhase(db2, "replay-phase", 0)
	if err != nil {
		t.Fatalf("QueryTracesByPhase(db2): %v", err)
	}
	json1, err := json.Marshal(traces)
	if err != nil {
		t.Fatalf("marshal traces: %v", err)
	}
	json2, err := json.Marshal(replayed)
	if err != nil {
		t.Fatalf("marshal replayed: %v", err)
	}
	if string(json1) != string(json2) {
		t.Fatalf("traces diverged after replay:\nincremental: %s\nreplayed:    %s", json1, json2)
	}
}

package application

import "testing"

// TestBuildContextPacketWatzupOmitsPhases proves the packet is
// stage-shaped (R4): watzup's playbook never calls `query phases`, so its
// packet must not carry a Phases field at all.
func TestBuildContextPacketWatzupOmitsPhases(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")

	pkg, err := BuildContextPacket(db, "watzup", "dev")
	if err != nil {
		t.Fatalf("BuildContextPacket: %v", err)
	}
	if pkg.Phases != nil {
		t.Fatalf("watzup packet Phases = %v, want nil (watzup.md never calls query phases)", pkg.Phases)
	}
	if pkg.Position.CurrentPhase == nil || *pkg.Position.CurrentPhase != "cli-domain" {
		t.Fatalf("watzup packet Position.CurrentPhase = %v, want cli-domain", pkg.Position.CurrentPhase)
	}
	if pkg.LatestRunID == nil || *pkg.LatestRunID != runID {
		t.Fatalf("watzup packet LatestRunID = %v, want %q", pkg.LatestRunID, runID)
	}
}

// TestBuildContextPacketWorkAndHandoffIncludePhases proves work, handoff,
// and check (R6) all get the phases list — the packet replacing their
// separate `query phases`/`resume` calls.
func TestBuildContextPacketWorkAndHandoffIncludePhases(t *testing.T) {
	db, changesetDir := freshDB(t)
	createLifecycleRun(t, db, changesetDir, "cli-domain")
	if _, _, err := CreateStory(db, changesetDir, "next-phase", "goal", "cli-domain"); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}

	for _, stage := range []string{"work", "handoff", "check"} {
		pkg, err := BuildContextPacket(db, stage, "dev")
		if err != nil {
			t.Fatalf("BuildContextPacket(%s): %v", stage, err)
		}
		if len(pkg.Phases) != 2 {
			t.Fatalf("%s packet Phases = %+v, want 2 rows", stage, pkg.Phases)
		}
	}
}

// TestBuildContextPacketTracesWindowedToCurrentPhase proves the window
// policy (P4 wave 2): the packet's Traces are the current phase's own
// traces, not another phase's, and not the whole table.
func TestBuildContextPacketTracesWindowedToCurrentPhase(t *testing.T) {
	db, changesetDir := freshDB(t)
	runA := createLifecycleRun(t, db, changesetDir, "phase-a")
	if _, _, err := CreateTrace(db, changesetDir, 1, "phase-a wave 1", runA, "", ""); err != nil {
		t.Fatalf("CreateTrace phase-a: %v", err)
	}

	if _, _, err := CreateStory(db, changesetDir, "phase-b", "goal", "phase-a"); err != nil {
		t.Fatalf("CreateStory phase-b: %v", err)
	}
	runB, _, err := CreateRun(db, changesetDir, "phase-b", "", "")
	if err != nil {
		t.Fatalf("CreateRun phase-b: %v", err)
	}
	if _, _, err := CreateTrace(db, changesetDir, 1, "phase-b wave 1", runB, "", ""); err != nil {
		t.Fatalf("CreateTrace phase-b: %v", err)
	}

	pkg, err := BuildContextPacket(db, "watzup", "dev")
	if err != nil {
		t.Fatalf("BuildContextPacket: %v", err)
	}
	if pkg.Position.CurrentPhase == nil || *pkg.Position.CurrentPhase != "phase-b" {
		t.Fatalf("Position.CurrentPhase = %v, want phase-b (meta.current_phase follows the latest run create)", pkg.Position.CurrentPhase)
	}
	if len(pkg.Traces) != 1 || pkg.Traces[0].Summary != "phase-b wave 1" {
		t.Fatalf("Traces = %+v, want only phase-b's trace", pkg.Traces)
	}
	if len(pkg.Omitted) != 0 {
		t.Fatalf("Omitted = %+v, want none (1 trace is well under the cap)", pkg.Omitted)
	}
}

// TestBuildContextPacketTracesCappedDeclaresOmitted proves R5: once a
// phase's trace history exceeds the window, the packet still returns the
// capped set but also declares what it left out and how to fetch it.
func TestBuildContextPacketTracesCappedDeclaresOmitted(t *testing.T) {
	db, changesetDir := freshDB(t)
	runID := createLifecycleRun(t, db, changesetDir, "cli-domain")
	const total = contextTraceTail + 5
	for i := 0; i < total; i++ {
		if _, _, err := CreateTrace(db, changesetDir, 1, "wave note", runID, "", ""); err != nil {
			t.Fatalf("CreateTrace #%d: %v", i, err)
		}
	}

	pkg, err := BuildContextPacket(db, "watzup", "dev")
	if err != nil {
		t.Fatalf("BuildContextPacket: %v", err)
	}
	if len(pkg.Traces) != contextTraceTail {
		t.Fatalf("Traces = %d rows, want capped at %d", len(pkg.Traces), contextTraceTail)
	}
	if len(pkg.Omitted) != 1 || pkg.Omitted[0].Field != "traces" {
		t.Fatalf("Omitted = %+v, want one traces entry", pkg.Omitted)
	}
	if pkg.Omitted[0].Fetch == "" {
		t.Fatal("Omitted[0].Fetch is empty, want a command an agent can run to see the rest")
	}
}

// TestBuildContextPacketNoCurrentPhaseHasEmptyTraces proves a fresh repo
// (no current_phase yet) gets an empty, not nil-panicking, Traces window.
func TestBuildContextPacketNoCurrentPhaseHasEmptyTraces(t *testing.T) {
	db, _ := freshDB(t)

	pkg, err := BuildContextPacket(db, "watzup", "dev")
	if err != nil {
		t.Fatalf("BuildContextPacket: %v", err)
	}
	if pkg.Position.CurrentPhase != nil {
		t.Fatalf("Position.CurrentPhase = %v, want nil", pkg.Position.CurrentPhase)
	}
	if len(pkg.Traces) != 0 {
		t.Fatalf("Traces = %+v, want none", pkg.Traces)
	}
}

// TestBuildContextPacketReusesResumeDrift proves the packet's
// position/drift/readiness fields are Resume's own derivation, not a
// second independent computation that could disagree with `resume --json`.
func TestBuildContextPacketReusesResumeDrift(t *testing.T) {
	db, changesetDir := freshDB(t)
	createLifecycleRun(t, db, changesetDir, "cli-domain")
	if _, err := db.Exec(`UPDATE meta SET docs_version = 'stale-version'`); err != nil {
		t.Fatalf("seed stale docs_version: %v", err)
	}

	resumeView, err := Resume(db, "0.9.9")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	pkg, err := BuildContextPacket(db, "watzup", "0.9.9")
	if err != nil {
		t.Fatalf("BuildContextPacket: %v", err)
	}
	if pkg.Readiness != resumeView.Readiness || len(pkg.Drift) != len(resumeView.Drift) {
		t.Fatalf("packet readiness/drift = %q/%v, want to match Resume() = %q/%v", pkg.Readiness, pkg.Drift, resumeView.Readiness, resumeView.Drift)
	}
	if len(pkg.Drift) == 0 {
		t.Fatal("test setup: expected stale_docs drift from version mismatch to prove reuse, got none")
	}
}

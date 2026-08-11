package interfaces

import (
	"encoding/json"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/application"
)

// TestCheckRecordAndQueryChecksRoundTrip proves P6's own success-signal
// fix: `query check --latest` only ever exposed the most recent verdict,
// leaving Validation's append-only history unqueryable. `query checks`
// closes that gap.
func TestCheckRecordAndQueryChecksRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := runDBCommand(t, "story", "--slug", "cli-domain", "--goal", "goal", "--json"); err != nil {
		t.Fatalf("story: %v", err)
	}

	runOut, err := runDBCommand(t, "run", "create", "--slug", "cli-domain", "--json")
	if err != nil {
		t.Fatalf("run create: %v (output=%s)", err, runOut)
	}
	var runResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(runOut), &runResult); err != nil {
		t.Fatalf("decode run create output %q: %v", runOut, err)
	}

	checkOut, err := runDBCommand(t, "check", "record", "--verdict", "REQUEST_CHANGES", "--run-id", runResult.ID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", `[{"command":"go test ./...","output_ref":"fail"}]`, "--json")
	if err != nil {
		t.Fatalf("check record (1): %v (output=%s)", err, checkOut)
	}

	checkOut2, err := runDBCommand(t, "check", "record", "--verdict", "APPROVED", "--run-id", runResult.ID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", `[{"command":"true","output_ref":"pass"}]`, "--json")
	if err != nil {
		t.Fatalf("check record (2): %v (output=%s)", err, checkOut2)
	}

	out, err := runDBCommand(t, "query", "checks", "--json")
	if err != nil {
		t.Fatalf("query checks: %v (output=%s)", err, out)
	}
	var checks []struct {
		ID      string `json:"id"`
		Phase   string `json:"phase"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal([]byte(out), &checks); err != nil {
		t.Fatalf("decode query checks output %q: %v", out, err)
	}
	if len(checks) != 2 {
		t.Fatalf("query checks = %+v, want 2 rows", checks)
	}
	if checks[0].Verdict != "REQUEST_CHANGES" || checks[1].Verdict != "APPROVED" {
		t.Fatalf("query checks verdicts = [%s %s], want [REQUEST_CHANGES APPROVED]", checks[0].Verdict, checks[1].Verdict)
	}
	if checks[0].Phase != "cli-domain" {
		t.Fatalf("query checks[0].Phase = %q, want cli-domain", checks[0].Phase)
	}

	tailedOut, err := runDBCommand(t, "query", "checks", "--tail", "1", "--json")
	if err != nil {
		t.Fatalf("query checks --tail 1: %v (output=%s)", err, tailedOut)
	}
	var tailed []struct {
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal([]byte(tailedOut), &tailed); err != nil {
		t.Fatalf("decode query checks --tail 1 output %q: %v", tailedOut, err)
	}
	if len(tailed) != 1 || tailed[0].Verdict != "APPROVED" {
		t.Fatalf("query checks --tail 1 = %+v, want only the most recent check", tailed)
	}

	// R2 (docs/audit/sdlc-gap-analysis.md, C4): a --phase filter naming a
	// story that does exist still returns an empty array, not an error —
	// the distinction unknown_phase draws is "never defined", not "no
	// rows yet".
	knownEmptyOut, err := runDBCommand(t, "query", "checks", "--phase", "cli-domain", "--tail", "1", "--json")
	if err != nil {
		t.Fatalf("query checks --phase cli-domain --tail 1: %v (output=%s)", err, knownEmptyOut)
	}
}

// TestQueryPhaseFilteredViewsRejectUnknownPhase proves R2
// (docs/audit/sdlc-gap-analysis.md, C4): traces/decisions/checks filtered
// by a --phase naming no story row fail loudly with unknown_phase instead
// of returning an indistinguishable empty array with exit 0.
func TestQueryPhaseFilteredViewsRejectUnknownPhase(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	for _, view := range []string{"checks", "traces", "decisions"} {
		_, err := runDBCommand(t, "query", view, "--phase", "no-such-phase", "--json")
		ce, ok := err.(*cliError)
		if !ok || ce.Code != "unknown_phase" {
			t.Fatalf("query %s --phase no-such-phase = %v, want unknown_phase cliError", view, err)
		}
	}
}

// TestQueryTracesKnownPhaseWithNoRowsReturnsEmptyArray proves the R2 fix
// doesn't overreach: a phase that does exist but has no traces yet still
// returns `[]`, not an error.
func TestQueryTracesKnownPhaseWithNoRowsReturnsEmptyArray(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := runDBCommand(t, "story", "--slug", "empty-phase", "--goal", "goal", "--json"); err != nil {
		t.Fatalf("story: %v", err)
	}

	out, err := runDBCommand(t, "query", "traces", "--phase", "empty-phase", "--json")
	if err != nil {
		t.Fatalf("query traces --phase empty-phase: %v (output=%s)", err, out)
	}
	var v []application.TraceView
	if jsonErr := json.Unmarshal([]byte(out), &v); jsonErr != nil {
		t.Fatalf("decode output %q: %v", out, jsonErr)
	}
	if len(v) != 0 {
		t.Fatalf("query traces --phase empty-phase = %+v, want an empty array", v)
	}
}

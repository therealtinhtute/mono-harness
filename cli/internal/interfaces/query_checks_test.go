package interfaces

import (
	"encoding/json"
	"testing"
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
}

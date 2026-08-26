package interfaces

import (
	"encoding/json"
	"testing"
)

func TestCheckRecordRejectsInvalidModeFlag(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := runDBCommand(t, "story", "--slug", "mode-flag", "--goal", "goal", "--json"); err != nil {
		t.Fatalf("story: %v", err)
	}
	runOut, err := runDBCommand(t, "run", "create", "--slug", "mode-flag", "--json")
	if err != nil {
		t.Fatalf("run create: %v (output=%s)", err, runOut)
	}
	var runResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(runOut), &runResult); err != nil {
		t.Fatalf("decode run create output %q: %v", runOut, err)
	}

	_, err = runDBCommand(t, "check", "record", "--verdict", "APPROVED", "--run-id", runResult.ID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", `[{"command":"true","output_ref":"pass"}]`, "--mode", "banana", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "invalid_check_mode" {
		t.Fatalf("check record --mode banana = %v, want invalid_check_mode cliError", err)
	}
}

func TestCheckRecordModeDefaultsToGateAndQueryChecksExposesIt(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := runDBCommand(t, "story", "--slug", "mode-view", "--goal", "goal", "--json"); err != nil {
		t.Fatalf("story: %v", err)
	}
	runOut, err := runDBCommand(t, "run", "create", "--slug", "mode-view", "--json")
	if err != nil {
		t.Fatalf("run create: %v (output=%s)", err, runOut)
	}
	var runResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(runOut), &runResult); err != nil {
		t.Fatalf("decode run create output %q: %v", runOut, err)
	}

	gateProof := `[{"command":"true","output_ref":"pass"}]`
	if _, err := runDBCommand(t, "check", "record", "--verdict", "APPROVED", "--run-id", runResult.ID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", gateProof, "--json"); err != nil {
		t.Fatalf("default-mode check record: %v", err)
	}
	fullOut, err := runDBCommand(t, "check", "record", "--verdict", "APPROVED", "--run-id", runResult.ID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", gateProof, "--mode", "full", "--json")
	if err != nil {
		t.Fatalf("full check record on checked story: %v (output=%s)", err, fullOut)
	}

	out, err := runDBCommand(t, "query", "checks", "--json")
	if err != nil {
		t.Fatalf("query checks: %v (output=%s)", err, out)
	}
	var checks []struct {
		ID      string  `json:"id"`
		Verdict string  `json:"verdict"`
		Mode    *string `json:"mode"`
	}
	if err := json.Unmarshal([]byte(out), &checks); err != nil {
		t.Fatalf("decode query checks output %q: %v", out, err)
	}
	if len(checks) != 2 {
		t.Fatalf("query checks = %+v, want 2 rows", checks)
	}
	if checks[0].Mode == nil || *checks[0].Mode != "gate" {
		t.Fatalf("query checks[0].mode = %v, want gate", checks[0].Mode)
	}
	if checks[1].Mode == nil || *checks[1].Mode != "full" {
		t.Fatalf("query checks[1].mode = %v, want full", checks[1].Mode)
	}
}

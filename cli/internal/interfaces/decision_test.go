package interfaces

import (
	"encoding/json"
	"testing"
)

func TestDecisionAddAndQueryDecisionsRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	out, err := runDBCommand(t, "decision", "add", "--decisions",
		`[{"decision":"picked A over B","rationale":"A matched the constraint"},{"decision":"deferred C","rationale":"out of scope"}]`,
		"--json")
	if err != nil {
		t.Fatalf("decision add: %v (output=%s)", err, out)
	}
	var addResult struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal([]byte(out), &addResult); err != nil {
		t.Fatalf("decode decision add output %q: %v", out, err)
	}
	if len(addResult.IDs) != 2 {
		t.Fatalf("decision add ids = %v, want 2", addResult.IDs)
	}

	out, err = runDBCommand(t, "query", "decisions", "--json")
	if err != nil {
		t.Fatalf("query decisions: %v (output=%s)", err, out)
	}
	var decisions []struct {
		ID        string `json:"id"`
		Decision  string `json:"decision"`
		Rationale string `json:"rationale"`
	}
	if err := json.Unmarshal([]byte(out), &decisions); err != nil {
		t.Fatalf("decode query decisions output %q: %v", out, err)
	}
	if len(decisions) != 2 {
		t.Fatalf("query decisions = %+v, want 2 rows", decisions)
	}
	if decisions[0].Decision != "picked A over B" || decisions[0].Rationale != "A matched the constraint" {
		t.Fatalf("query decisions[0] = %+v, want the first recorded decision", decisions[0])
	}
}

func TestDecisionAddRejectsEmptyBatch(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "decision", "add", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "empty_decisions" {
		t.Fatalf("decision add with no --decisions = %v, want empty_decisions cliError", err)
	}
}

func TestDecisionAddRejectsInvalidJSON(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "decision", "add", "--decisions", "{not json", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "invalid_decisions" {
		t.Fatalf("decision add with malformed JSON = %v, want invalid_decisions cliError", err)
	}
}

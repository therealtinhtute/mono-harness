package interfaces

import (
	"encoding/json"
	"testing"
)

// TestHandoffRecordNextActionCLIRoundTrip is the CLI-level counterpart of
// TestHandoffRecordNextActionRoundTripsThroughResumeAndQuery: --next-action
// persists through the real command surface and comes back out of both
// `resume` (latest_handoff_id) and `query handoff --latest` (exact_next_action).
func TestHandoffRecordNextActionCLIRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	out, err := runDBCommand(t, "handoff", "record", "--next-action", "start p2-complete-the-index wave 1",
		"--open-items", `["owner decision pending"]`, "--json")
	if err != nil {
		t.Fatalf("handoff record: %v (output=%s)", err, out)
	}
	var recorded struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &recorded); err != nil {
		t.Fatalf("decode handoff record output %q: %v", out, err)
	}

	out, err = runDBCommand(t, "resume", "--json")
	if err != nil {
		t.Fatalf("resume: %v (output=%s)", err, out)
	}
	var resumeView struct {
		LatestHandoffID *string `json:"latest_handoff_id"`
	}
	if err := json.Unmarshal([]byte(out), &resumeView); err != nil {
		t.Fatalf("decode resume output %q: %v", out, err)
	}
	if resumeView.LatestHandoffID == nil || *resumeView.LatestHandoffID != recorded.ID {
		t.Fatalf("resume latest_handoff_id = %v, want %q", resumeView.LatestHandoffID, recorded.ID)
	}

	out, err = runDBCommand(t, "query", "handoff", "--latest", "--json")
	if err != nil {
		t.Fatalf("query handoff --latest: %v (output=%s)", err, out)
	}
	var handoffView struct {
		ID         string   `json:"id"`
		NextAction *string  `json:"exact_next_action"`
		OpenItems  []string `json:"open_items"`
	}
	if err := json.Unmarshal([]byte(out), &handoffView); err != nil {
		t.Fatalf("decode query handoff output %q: %v", out, err)
	}
	if handoffView.ID != recorded.ID {
		t.Fatalf("query handoff id = %q, want %q", handoffView.ID, recorded.ID)
	}
	if handoffView.NextAction == nil || *handoffView.NextAction != "start p2-complete-the-index wave 1" {
		t.Fatalf("query handoff exact_next_action = %v, want %q", handoffView.NextAction, "start p2-complete-the-index wave 1")
	}
	if len(handoffView.OpenItems) != 1 || handoffView.OpenItems[0] != "owner decision pending" {
		t.Fatalf("query handoff open_items = %v, want [owner decision pending]", handoffView.OpenItems)
	}
}

func TestQueryHandoffRequiresLatestFlag(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := runDBCommand(t, "init"); err != nil {
		t.Fatalf("init: %v", err)
	}

	_, err := runDBCommand(t, "query", "handoff", "--json")
	ce, ok := err.(*cliError)
	if !ok || ce.Code != "unknown_view" {
		t.Fatalf("query handoff without --latest = %v, want unknown_view cliError", err)
	}
}

package application

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type gateFixtureInput struct {
	Lane     string   `json:"lane"`
	Gathered []string `json:"gathered"`
}

func loadGateFixture(t *testing.T, name string) gateFixtureInput {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "gate-missing-proof", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var in gateFixtureInput
	if err := json.Unmarshal(data, &in); err != nil {
		t.Fatalf("unmarshal fixture %s: %v", name, err)
	}
	return in
}

// TestGateMissingProof proves the matrix names the exact missing evidence
// class for a sample "normal" lane chain that gathered integration and
// command-output proof but never ran a unit test — the one cell that
// becomes required starting at the normal lane (validation-gate-PLAN.md T4,
// "sample chain minus one required proof → expect FAIL naming that proof").
func TestGateMissingProof(t *testing.T) {
	in := loadGateFixture(t, "gathered.json")
	gathered := map[string]bool{}
	for _, class := range in.Gathered {
		gathered[class] = true
	}

	got, err := EvaluateGate(in.Lane, gathered)
	if err != nil {
		t.Fatalf("EvaluateGate: %v", err)
	}

	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal verdict: %v", err)
	}
	wantJSON, err := os.ReadFile(filepath.Join("..", "..", "testdata", "gate-missing-proof", "expected-verdict.json"))
	if err != nil {
		t.Fatalf("read expected-verdict.json: %v", err)
	}
	if string(gotJSON)+"\n" != string(wantJSON) {
		t.Fatalf("verdict mismatch:\n got:  %s\n want: %s", gotJSON, wantJSON)
	}
}

// TestGateDeterminism proves identical input produces byte-identical
// verdict JSON across repeated calls — validation-gate-PLAN.md T4's
// determinism requirement at the full gate-flow level (T2's
// TestAuditDeterministic already covers audit-composition; this covers
// the matrix-evaluation step itself).
func TestGateDeterminism(t *testing.T) {
	in := loadGateFixture(t, "gathered.json")
	gathered := map[string]bool{}
	for _, class := range in.Gathered {
		gathered[class] = true
	}

	first, err := EvaluateGate(in.Lane, gathered)
	if err != nil {
		t.Fatalf("EvaluateGate (first): %v", err)
	}
	second, err := EvaluateGate(in.Lane, gathered)
	if err != nil {
		t.Fatalf("EvaluateGate (second): %v", err)
	}

	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("non-deterministic verdict:\n first:  %s\n second: %s", firstJSON, secondJSON)
	}
}

// TestGateNoMissingRequired proves a lane whose gathered evidence covers
// every required cell verdicts PASS with an empty (not null) list.
func TestGateNoMissingRequired(t *testing.T) {
	gathered := map[string]bool{"command-output": true}
	got, err := EvaluateGate("tiny", gathered)
	if err != nil {
		t.Fatalf("EvaluateGate: %v", err)
	}
	if got.Verdict != "PASS" {
		t.Fatalf("verdict = %q, want PASS", got.Verdict)
	}
	if len(got.MissingRequired) != 0 {
		t.Fatalf("missing_required = %v, want empty", got.MissingRequired)
	}
}

func TestGateInvalidLane(t *testing.T) {
	_, err := EvaluateGate("bogus", map[string]bool{})
	if err == nil {
		t.Fatal("expected error for invalid lane, got nil")
	}
}

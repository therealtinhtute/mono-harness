package application

import (
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func wantValidationCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("RenderRecap: err = nil, want ValidationError code %q", wantCode)
	}
	ve, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("RenderRecap: err = %v (%T), want *domain.ValidationError", err, err)
	}
	if ve.Code != wantCode {
		t.Fatalf("RenderRecap: code = %q, want %q", ve.Code, wantCode)
	}
}

func TestRenderRecapEmptyState(t *testing.T) {
	view := ResumeView{Readiness: "clean"}
	facts := RecapFacts{Branch: "master"}

	got, err := RenderRecap(view, facts)
	if err != nil {
		t.Fatalf("RenderRecap: %v", err)
	}
	want := "Nhánh sạch — không có thay đổi nào so với main.\n" +
		"Next: Bắt đầu task mới hoặc kéo thay đổi mới nhất.\n"
	if got != want {
		t.Fatalf("RenderRecap = %q, want %q", got, want)
	}
}

func TestRenderRecapTitleFormat(t *testing.T) {
	view := ResumeView{Readiness: "clean"}
	facts := RecapFacts{Branch: "feature/inbox-ui", Ahead: 4}

	got, err := RenderRecap(view, facts)
	if err != nil {
		t.Fatalf("RenderRecap: %v", err)
	}
	firstLine := strings.SplitN(got, "\n", 2)[0]
	if !strings.HasPrefix(firstLine, "Recap — feature/inbox-ui (") || !strings.HasSuffix(firstLine, ")") {
		t.Fatalf("title line = %q, want %q prefix/suffix", firstLine, "Recap — feature/inbox-ui (...)")
	}
}

func TestRenderRecapRiskTable(t *testing.T) {
	view := ResumeView{Readiness: "in-progress"}
	facts := RecapFacts{
		Branch: "master", Ahead: 1,
		Risks: []RecapRisk{{Risk: "thiếu test", Severity: "vừa", Action: "viết test cho X"}},
	}

	got, err := RenderRecap(view, facts)
	if err != nil {
		t.Fatalf("RenderRecap: %v", err)
	}
	wantTable := "Risks\n| Risk | Mức độ | Action |\n|------|--------|--------|\n| thiếu test | vừa | viết test cho X |\n"
	if !strings.Contains(got, wantTable) {
		t.Fatalf("RenderRecap = %q, want to contain %q", got, wantTable)
	}
}

func TestRenderRecapInvalidSeverity(t *testing.T) {
	view := ResumeView{Readiness: "in-progress"}
	facts := RecapFacts{
		Branch: "master",
		Risks:  []RecapRisk{{Risk: "x", Severity: "high", Action: "y"}},
	}

	_, err := RenderRecap(view, facts)
	wantValidationCode(t, err, "invalid_severity")
}

func TestRenderRecapForbiddenPhrase(t *testing.T) {
	view := ResumeView{Readiness: "in-progress"}
	facts := RecapFacts{Branch: "master", NextAction: "run git log to see history"}

	_, err := RenderRecap(view, facts)
	wantValidationCode(t, err, "forbidden_phrase")
}

func TestRenderRecapScorePatternForbidden(t *testing.T) {
	view := ResumeView{Readiness: "in-progress"}
	facts := RecapFacts{Branch: "master", NextAction: "Quality: 8/10, ship it"}

	_, err := RenderRecap(view, facts)
	wantValidationCode(t, err, "forbidden_phrase")
}

func TestRenderRecapDriftOverridesNextAction(t *testing.T) {
	view := ResumeView{
		Readiness: "drifted",
		Drift:     []DriftFinding{{Type: "unknown_phase", Detail: "d", Recovery: "record it via `zharness story ...`"}},
	}
	facts := RecapFacts{Branch: "master", NextAction: "this should be ignored"}

	got, err := RenderRecap(view, facts)
	if err != nil {
		t.Fatalf("RenderRecap: %v", err)
	}
	if !strings.Contains(got, "Next: record it via `zharness story ...`\n") {
		t.Fatalf("RenderRecap = %q, want Next line to be the drift recovery verbatim", got)
	}
	if strings.Contains(got, "this should be ignored") {
		t.Fatalf("RenderRecap = %q, want facts.NextAction ignored when drifted", got)
	}
}

func TestRenderRecapChangesCappedAtFive(t *testing.T) {
	view := ResumeView{Readiness: "in-progress"}
	facts := RecapFacts{
		Branch:  "master",
		Ahead:   1,
		Changes: []string{"a", "b", "c", "d"},
		WIP:     []string{"e", "f", "g"},
	}

	got, err := RenderRecap(view, facts)
	if err != nil {
		t.Fatalf("RenderRecap: %v", err)
	}
	// Changes has 4 items + WIP has 3 = 7 combined; capped at 5, so only the
	// first WIP entry ("e") survives the truncation.
	if !strings.Contains(got, "[WIP] e") {
		t.Fatalf("RenderRecap = %q, want the 5th combined item ([WIP] e) present", got)
	}
	if strings.Contains(got, "[WIP] f") || strings.Contains(got, "[WIP] g") {
		t.Fatalf("RenderRecap = %q, want 6th/7th combined items truncated", got)
	}
}

func TestRenderRecapNoHarnessBranch(t *testing.T) {
	view := ResumeView{Readiness: "no-harness", Drift: []DriftFinding{}}
	facts := RecapFacts{Branch: "feature/legacy-migrate", Ahead: 1}

	got, err := RenderRecap(view, facts)
	if err != nil {
		t.Fatalf("RenderRecap: %v", err)
	}
	if !strings.Contains(got, "Readiness: no-harness\n") {
		t.Fatalf("RenderRecap = %q, want Readiness: no-harness line", got)
	}
	if !strings.Contains(got, "Không có handoff") {
		t.Fatalf("RenderRecap = %q, want default no-handoff context bullet", got)
	}
}

// TestRenderRecapAgainstRealResumeStates exercises RenderRecap against a
// real migrated sqlite db (via freshDB) for the clean and in-progress
// readiness states, satisfying the plan's integration proof-class cell
// the same way next_test.go's freshDB-based cases do.
func TestRenderRecapAgainstRealResumeStates(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		db := freshDB(t)
		seedRun(t, db)
		if _, err := db.Exec(`UPDATE stories SET status = ? WHERE slug = ?`, domain.StoryDone, "cli-domain"); err != nil {
			t.Fatalf("seed story status: %v", err)
		}
		setMeta(t, db, map[string]any{"current_phase": "cli-domain"})

		view, err := Resume(db, "dev")
		if err != nil {
			t.Fatalf("Resume: %v", err)
		}
		got, err := RenderRecap(view, RecapFacts{Branch: "master"})
		if err != nil {
			t.Fatalf("RenderRecap: %v", err)
		}
		if !strings.Contains(got, "Nhánh sạch") {
			t.Fatalf("RenderRecap = %q, want the clean empty-state form", got)
		}
	})

	t.Run("drifted", func(t *testing.T) {
		db := freshDB(t)
		db.SetMaxOpenConns(1)
		if _, err := db.Exec(`PRAGMA foreign_keys=OFF;`); err != nil {
			t.Fatalf("disable foreign_keys: %v", err)
		}
		if _, err := db.Exec(`UPDATE meta SET current_phase = ?`, "no-such-phase"); err != nil {
			t.Fatalf("seed stale current_phase: %v", err)
		}
		if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
			t.Fatalf("re-enable foreign_keys: %v", err)
		}

		view, err := Resume(db, "dev")
		if err != nil {
			t.Fatalf("Resume: %v", err)
		}
		got, err := RenderRecap(view, RecapFacts{Branch: "master", Ahead: 1, NextAction: "ignored"})
		if err != nil {
			t.Fatalf("RenderRecap: %v", err)
		}
		if !strings.Contains(got, "Readiness: drifted\n") {
			t.Fatalf("RenderRecap = %q, want Readiness: drifted", got)
		}
		if !strings.Contains(got, view.Drift[0].Recovery) {
			t.Fatalf("RenderRecap = %q, want the drift recovery verbatim", got)
		}
	})
}

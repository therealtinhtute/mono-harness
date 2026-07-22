package application

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// chdirFixture creates a fresh temp dir, chdirs the test into it (t.Chdir
// restores the original cwd on cleanup), and returns nothing — every next.go
// path is cwd-relative, mirroring how work.md's prose describes them.
func chdirFixture(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func writeRoadmap(t *testing.T, slugs ...string) {
	t.Helper()
	content := ""
	for i, s := range slugs {
		content += "## Phase " + string(rune('1'+i)) + ": " + s + "\n\ngoal\n\n"
	}
	writeFile(t, nextRoadmapPath, content)
}

func writePhaseArtifacts(t *testing.T, slug, planBody string) {
	t.Helper()
	writeFile(t, phasePlanPath(slug), planBody)
	writeFile(t, phaseContextPath(slug), "# context\n")
}

// seedStory writes a story row with an explicit slug + status, since seedRun
// (helpers_test.go) always hardcodes the "cli-domain" slug.
func seedStory(t *testing.T, db *sql.DB, changesetDir, slug, status string) {
	t.Helper()
	if _, _, err := AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "story", ID: ulid.Make().String(), Fields: map[string]any{
			"slug": slug, "goal": "goal", "status": status, "created_at": "2026-07-22T00:00:00Z",
		}, At: "2026-07-22T00:00:00Z"},
	}); err != nil {
		t.Fatalf("seed story %s: %v", slug, err)
	}
}

func TestNextUnknownArgument(t *testing.T) {
	chdirFixture(t)
	_, err := Next(nil, "bogus")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_argument" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_argument}", err)
	}
}

func TestNextSimpleMode(t *testing.T) {
	chdirFixture(t)
	view, err := Next(nil, "simple")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Mode != "simple" || view.Stop != nil {
		t.Fatalf("view = %+v, want mode=simple, no stop", view)
	}
}

func TestNextAutoResolvesSimpleWhenNoSpec(t *testing.T) {
	chdirFixture(t)
	view, err := Next(nil, "")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Mode != "simple" || view.Stop != nil {
		t.Fatalf("view = %+v, want auto to resolve to simple", view)
	}
}

func TestNextAutoAmbiguousWhenSpecAndBrainstorm(t *testing.T) {
	chdirFixture(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeFile(t, ".kit/reports/brainstorm/20260722-x.md", "# brainstorm\n")

	view, err := Next(nil, "auto")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "ambiguous" {
		t.Fatalf("view = %+v, want stop.code=ambiguous", view)
	}
}

func TestNextFullNoSpec(t *testing.T) {
	chdirFixture(t)
	view, err := Next(nil, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "no-spec" {
		t.Fatalf("view = %+v, want stop.code=no-spec", view)
	}
}

func TestNextFullNoPlan(t *testing.T) {
	chdirFixture(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")

	view, err := Next(nil, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "no-plan" {
		t.Fatalf("view = %+v, want stop.code=no-plan", view)
	}
}

func TestNextFullNoPhase(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha")
	seedStory(t, db, changesetDir, "alpha", domain.StoryPlanned)
	// deliberately no phase PLAN.md/CONTEXT.md for "alpha"

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "no-phase" {
		t.Fatalf("view = %+v, want stop.code=no-phase", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "alpha" {
		t.Fatalf("active_phase = %v, want alpha", view.ActivePhase)
	}
}

func TestNextFullPlaceholderPlan(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha")
	seedStory(t, db, changesetDir, "alpha", domain.StoryPlanned)
	writePhaseArtifacts(t, "alpha", "# plan\nTBD: fill this in\n")

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "placeholder-plan" {
		t.Fatalf("view = %+v, want stop.code=placeholder-plan", view)
	}
}

func TestNextFullReady(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha")
	seedStory(t, db, changesetDir, "alpha", domain.StoryPlanned)
	writePhaseArtifacts(t, "alpha", "# plan\nreal steps, no placeholders\n")

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop != nil {
		t.Fatalf("view = %+v, want no stop", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "alpha" {
		t.Fatalf("active_phase = %v, want alpha", view.ActivePhase)
	}
}

func TestNextFullMultipleIncomplete(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha", "beta")
	seedStory(t, db, changesetDir, "alpha", domain.StoryPlanned)
	seedStory(t, db, changesetDir, "beta", domain.StoryPlanned)

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "multiple-incomplete" {
		t.Fatalf("view = %+v, want stop.code=multiple-incomplete", view)
	}
}

func TestNextFullAllPhasesDone(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha")
	seedStory(t, db, changesetDir, "alpha", domain.StoryDone)

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "all-phases-done" {
		t.Fatalf("view = %+v, want stop.code=all-phases-done", view)
	}
}

func TestNextFullExplicitPhaseBypassesSelection(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha", "beta")
	seedStory(t, db, changesetDir, "alpha", domain.StoryDone)
	seedStory(t, db, changesetDir, "beta", domain.StoryDone)
	writePhaseArtifacts(t, "beta", "# plan\nreal steps\n")

	view, err := Next(db, "phase beta")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop != nil {
		t.Fatalf("view = %+v, want no stop (explicit phase bypasses done-ness check)", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "beta" {
		t.Fatalf("active_phase = %v, want beta", view.ActivePhase)
	}
}

func TestNextFullPhaseSelectionWithoutDB(t *testing.T) {
	chdirFixture(t)
	writeFile(t, nextSpecPath, "# spec\nlocked\n")
	writeRoadmap(t, "alpha")
	writePhaseArtifacts(t, "alpha", "# plan\nreal steps\n")

	view, err := Next(nil, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop != nil {
		t.Fatalf("view = %+v, want no stop (nil db treats phase as incomplete-but-selectable)", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "alpha" {
		t.Fatalf("active_phase = %v, want alpha", view.ActivePhase)
	}
}

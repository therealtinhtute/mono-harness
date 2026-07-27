package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func writeActivePlan(t *testing.T, name string, slugs ...string) string {
	t.Helper()
	var content strings.Builder
	content.WriteString("# Active plan\n\n## Phases and Verification\n")
	for i, slug := range slugs {
		fmt.Fprintf(&content, "### Phase %d: %s\n- phase_slug: %s\n- goal: goal\n\n", i+1, slug, slug)
	}
	path := filepath.Join("docs", "plans", "active", name+".md")
	writeFile(t, path, content.String())
	return path
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

func TestNextAutoResolvesSimpleWithoutActivePlan(t *testing.T) {
	chdirFixture(t)
	view, err := Next(nil, "")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Mode != "simple" || view.Stop != nil {
		t.Fatalf("view = %+v, want auto to resolve to simple", view)
	}
}

func TestNextAutoIgnoresEmptyActivePlan(t *testing.T) {
	chdirFixture(t)
	writeFile(t, filepath.Join("docs", "plans", "active", "empty.md"), " \n\t")

	view, err := Next(nil, "auto")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Mode != "simple" || view.Stop != nil {
		t.Fatalf("view = %+v, want empty markdown ignored", view)
	}
}

func TestNextActivePlanOnlyAutoAndFullReady(t *testing.T) {
	for _, argument := range []string{"auto", "full"} {
		t.Run(argument, func(t *testing.T) {
			chdirFixture(t)
			writeActivePlan(t, "initiative", "alpha")

			view, err := Next(nil, argument)
			if err != nil {
				t.Fatalf("Next: %v", err)
			}
			if view.Mode != "full" || view.Stop != nil {
				t.Fatalf("view = %+v, want full and ready", view)
			}
			if view.ActivePhase == nil || *view.ActivePhase != "alpha" {
				t.Fatalf("active_phase = %v, want alpha", view.ActivePhase)
			}
		})
	}
}

func TestNextMultipleActivePlansAreAmbiguous(t *testing.T) {
	for _, argument := range []string{"auto", "full"} {
		t.Run(argument, func(t *testing.T) {
			chdirFixture(t)
			writeActivePlan(t, "beta", "beta-phase")
			writeActivePlan(t, "alpha", "alpha-phase")

			view, err := Next(nil, argument)
			if err != nil {
				t.Fatalf("Next: %v", err)
			}
			if view.Stop == nil || view.Stop.Code != "ambiguous" {
				t.Fatalf("view = %+v, want stop.code=ambiguous", view)
			}
			alpha := strings.Index(view.Stop.Message, "alpha.md")
			beta := strings.Index(view.Stop.Message, "beta.md")
			if alpha < 0 || beta < 0 || alpha > beta {
				t.Fatalf("message = %q, want deterministic path order", view.Stop.Message)
			}
		})
	}
}

func TestNextFullWithoutActivePlanStopsNoPlan(t *testing.T) {
	chdirFixture(t)
	view, err := Next(nil, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "no-plan" {
		t.Fatalf("view = %+v, want stop.code=no-plan", view)
	}
	if !strings.Contains(view.Stop.Recovery, "brainstorm lock") {
		t.Fatalf("recovery = %q, want brainstorm lock", view.Stop.Recovery)
	}
}

func TestNextFullStopsNoPhaseWhenActivePlanHasNoPhaseDefinitions(t *testing.T) {
	chdirFixture(t)
	writeActivePlan(t, "initiative")

	view, err := Next(nil, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "no-phase" || view.Stop.Recovery != "to-plan" {
		t.Fatalf("view = %+v, want no-phase routed to to-plan", view)
	}
}

func TestNextFullStopsNoPhaseForExplicitMissingSlug(t *testing.T) {
	chdirFixture(t)
	writeActivePlan(t, "initiative", "alpha", "beta")

	view, err := Next(nil, "full phase gamma")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "no-phase" || view.Stop.Recovery != "to-plan" {
		t.Fatalf("view = %+v, want no-phase routed to to-plan", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "gamma" {
		t.Fatalf("active_phase = %v, want gamma", view.ActivePhase)
	}
}

func TestNextFullStopsForPlaceholderInActivePlan(t *testing.T) {
	chdirFixture(t)
	path := writeActivePlan(t, "initiative", "alpha")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	writeFile(t, path, string(data)+"- task: TODO fill this in\n")

	view, err := Next(nil, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "placeholder-plan" || view.Stop.Recovery != "to-plan" {
		t.Fatalf("view = %+v, want placeholder-plan routed to to-plan", view)
	}
}

func TestNextSelectsFirstIncompletePhaseInPlanOrder(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeActivePlan(t, "initiative", "alpha", "beta", "gamma")
	seedStory(t, db, changesetDir, "alpha", domain.StoryDone)
	seedStory(t, db, changesetDir, "beta", domain.StoryPlanned)
	seedStory(t, db, changesetDir, "gamma", domain.StoryPlanned)

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop != nil {
		t.Fatalf("view = %+v, want first incomplete phase without ambiguity", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "beta" {
		t.Fatalf("active_phase = %v, want beta", view.ActivePhase)
	}
}

func TestNextFullAllPhasesDone(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeActivePlan(t, "initiative", "alpha", "beta")
	seedStory(t, db, changesetDir, "alpha", domain.StoryDone)
	seedStory(t, db, changesetDir, "beta", domain.StoryDone)

	view, err := Next(db, "full")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop == nil || view.Stop.Code != "all-phases-done" {
		t.Fatalf("view = %+v, want stop.code=all-phases-done", view)
	}
}

func TestNextExplicitPhaseBypassesDoneSelection(t *testing.T) {
	chdirFixture(t)
	db, changesetDir := freshDB(t)
	writeActivePlan(t, "initiative", "alpha", "beta")
	seedStory(t, db, changesetDir, "alpha", domain.StoryDone)
	seedStory(t, db, changesetDir, "beta", domain.StoryDone)

	view, err := Next(db, "phase beta")
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if view.Stop != nil {
		t.Fatalf("view = %+v, want explicit phase to bypass done selection", view)
	}
	if view.ActivePhase == nil || *view.ActivePhase != "beta" {
		t.Fatalf("active_phase = %v, want beta", view.ActivePhase)
	}
}

func TestNextNilDBSelectsFirstPlanPhase(t *testing.T) {
	chdirFixture(t)
	writeActivePlan(t, "initiative", "alpha", "beta")

	view, err := Next(nil, "full")
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

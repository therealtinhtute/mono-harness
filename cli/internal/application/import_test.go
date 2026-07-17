package application

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func copyFixtureFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", dst, err)
	}
}

// TestImportRoundTrip exercises cli-core-PLAN.md's T4 verification: init
// (via Migrate on a fresh db) -> import -> query state, then a second
// import against unchanged legacy input producing zero new changesets.
//
// The fixture under cli/testdata/legacy-kit/ is this repo's own real
// .kit/workflow-state.yml + .kit/planning/ROADMAP.md at the time cli-core
// started (current_phase=cli-core, entry_phase=harness-concept,
// latest_cook_run pointing at the harness-contracts run artifact).
func TestImportRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "legacy-kit")
	repoRoot := t.TempDir()

	copyFixtureFile(t,
		filepath.Join(fixtureDir, "workflow-state.yml"),
		filepath.Join(repoRoot, ".kit", "workflow-state.yml"))
	copyFixtureFile(t,
		filepath.Join(fixtureDir, "planning", "ROADMAP.md"),
		filepath.Join(repoRoot, ".kit", "planning", "ROADMAP.md"))

	t.Chdir(repoRoot)

	db, err := infrastructure.Open(filepath.Join(repoRoot, "harness.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if _, _, err := infrastructure.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	changesetDir := filepath.Join(repoRoot, ".kit", "changesets")

	result, err := Import(db, ".kit/", changesetDir)
	if err != nil {
		t.Fatalf("Import (first): %v", err)
	}
	// 3 stories (cli-core, harness-concept, harness-contracts) + 1 run
	// (harness-contracts' latest_cook_run) + 1 meta update = 5.
	if result.Imported != 5 || result.Skipped != 0 {
		t.Fatalf("first import = (imported=%d, skipped=%d), want (5, 0); changesets=%v",
			result.Imported, result.Skipped, result.ChangesetsWritten)
	}
	if len(result.ChangesetsWritten) != 5 {
		t.Fatalf("first import changesets_written = %d files, want 5", len(result.ChangesetsWritten))
	}

	state, err := QueryState(db)
	if err != nil {
		t.Fatalf("QueryState: %v", err)
	}
	if state.CurrentPhase == nil || *state.CurrentPhase != "cli-core" {
		t.Fatalf("state.CurrentPhase = %v, want cli-core", state.CurrentPhase)
	}
	if state.EntryPhase == nil || *state.EntryPhase != "harness-concept" {
		t.Fatalf("state.EntryPhase = %v, want harness-concept", state.EntryPhase)
	}
	if state.SchemaVersion != infrastructure.CurrentSchemaVersion() {
		t.Fatalf("state.SchemaVersion = %d, want %d", state.SchemaVersion, infrastructure.CurrentSchemaVersion())
	}
	if state.LatestRunID == nil || *state.LatestRunID == "" {
		t.Fatalf("state.LatestRunID = %v, want non-empty", state.LatestRunID)
	}
	if state.LatestCheckID != nil {
		t.Fatalf("state.LatestCheckID = %v, want nil (import never synthesizes a checks row)", state.LatestCheckID)
	}

	phases, err := QueryPhases(db)
	if err != nil {
		t.Fatalf("QueryPhases: %v", err)
	}
	statusBySlug := map[string]string{}
	for _, p := range phases {
		statusBySlug[p.Slug] = p.Status
	}
	if statusBySlug["harness-contracts"] != domain.StoryInProgress {
		t.Fatalf("harness-contracts status = %q, want %q (owns the imported run)",
			statusBySlug["harness-contracts"], domain.StoryInProgress)
	}
	if statusBySlug["cli-core"] != domain.StoryPlanned {
		t.Fatalf("cli-core status = %q, want %q", statusBySlug["cli-core"], domain.StoryPlanned)
	}
	if statusBySlug["harness-concept"] != domain.StoryPlanned {
		t.Fatalf("harness-concept status = %q, want %q", statusBySlug["harness-concept"], domain.StoryPlanned)
	}

	artifacts, err := QueryArtifacts(db, "harness-contracts")
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(artifacts) != 1 || artifacts[0].ArtifactPath != ".kit/runs/work/20260717-1100-harness-contracts.md" {
		t.Fatalf("QueryArtifacts(harness-contracts) = %+v, want 1 row with the legacy artifact_path", artifacts)
	}

	// Re-running import against unchanged legacy input must be a pure
	// no-op: every entity re-examined, nothing rewritten.
	result2, err := Import(db, ".kit/", changesetDir)
	if err != nil {
		t.Fatalf("Import (second): %v", err)
	}
	if result2.Imported != 0 || len(result2.ChangesetsWritten) != 0 {
		t.Fatalf("second import = (imported=%d, changesets=%v), want (0, [])",
			result2.Imported, result2.ChangesetsWritten)
	}
	if result2.Skipped != 5 {
		t.Fatalf("second import skipped=%d, want 5", result2.Skipped)
	}
}

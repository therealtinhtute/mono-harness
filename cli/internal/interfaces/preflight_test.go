package interfaces

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func runPreflightCommand(t *testing.T, version string, args ...string) (string, error) {
	t.Helper()
	jsonOutput = false
	cmd := NewRootCmd(version)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), err
}

func decodePreflight(t *testing.T, output string) application.PreflightView {
	t.Helper()
	var view application.PreflightView
	if err := json.Unmarshal([]byte(output), &view); err != nil {
		t.Fatalf("unmarshal preflight output %q: %v", output, err)
	}
	return view
}

func listPreflightChangesets(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(".kit", "changesets", "*.jsonl"))
	if err != nil {
		t.Fatalf("Glob() changesets error = %v", err)
	}
	sort.Strings(matches)
	return matches
}

func listBoundedLifecycleMarkdown(t *testing.T) []string {
	t.Helper()
	var matches []string
	handoffPath := filepath.Join(".kit", "HANDOFF.md")
	if _, err := os.Stat(handoffPath); err == nil {
		matches = append(matches, handoffPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Stat() handoff error = %v", err)
	}

	roots := []string{
		filepath.Join(".kit", "planning"),
		filepath.Join(".kit", "runs"),
		filepath.Join(".kit", "reports"),
		filepath.Join("docs", "plans", "active"),
		filepath.Join("docs", "plans", "completed"),
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && filepath.Ext(path) == ".md" {
				matches = append(matches, path)
			}
			return nil
		})
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("WalkDir(%q) error = %v", root, err)
		}
	}
	sort.Strings(matches)
	return matches
}

func TestPreflightCommandReducedWithoutHarness(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "simple", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Readiness != application.PreflightReduced || view.DB != application.PreflightDBMissing || view.Docs != application.PreflightDocsMissing {
		t.Fatalf("view = %+v", view)
	}
	if view.Playbook != "" {
		t.Fatalf("view playbook = %q, want empty while docs are missing", view.Playbook)
	}
}

func TestPreflightCommandTreatsMissingStagePlaybookAsMissingDocs(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(preflightDocsPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "simple", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Docs != application.PreflightDocsMissing || view.Playbook != "" {
		t.Fatalf("view = %+v, want missing docs and no playbook", view)
	}
}

func TestPreflightCommandWorkAutoRequiresHarnessForActivePlanOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	planPath := filepath.Join("docs", "plans", "active", "initiative.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(planPath, []byte("# active plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Mode != "durable" || view.Stop == nil || view.Stop.Code != "harness_required" {
		t.Fatalf("view = %+v, want durable harness_required stop", view)
	}
}

func TestPreflightCommandWorkAutoUsesReducedWithoutActivePlan(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Mode != "reduced" || view.Stop != nil {
		t.Fatalf("view = %+v, want reduced route without an active plan", view)
	}
}

// TestPreflightCommandWorkAutoUsesReducedWithActivePlanButNoInProgressPhase
// is the regression for F2/V2 (docs/audit/workflow-harness-ceremony-audit.md):
// a live, readable harness with an active plan file but no story actually
// in-progress must not route a small unrelated change through the full
// durable ceremony path merely because some active plan exists.
func TestPreflightCommandWorkAutoUsesReducedWithActivePlanButNoInProgressPhase(t *testing.T) {
	t.Chdir(t.TempDir())
	seedPreflightWorkPlaybookAndPlan(t)
	seedPreflightStory(t, "planned")

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Mode != "reduced" || view.Stop != nil {
		t.Fatalf("view = %+v, want reduced route: active plan present, but no story in-progress", view)
	}
}

// TestPreflightCommandWorkAutoUsesDurableWithActiveInProgressPhase is the
// positive control for the fix above: an active plan with a genuinely
// in-progress story still resolves to the full durable path.
func TestPreflightCommandWorkAutoUsesDurableWithActiveInProgressPhase(t *testing.T) {
	t.Chdir(t.TempDir())
	seedPreflightWorkPlaybookAndPlan(t)
	seedPreflightStory(t, "in-progress")

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Mode != "durable" || view.Stop != nil {
		t.Fatalf("view = %+v, want durable route: a story is genuinely in-progress", view)
	}
}

func seedPreflightWorkPlaybookAndPlan(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(preflightDocsPath, "playbooks"), 0o755); err != nil {
		t.Fatalf("MkdirAll playbooks: %v", err)
	}
	if err := os.WriteFile(filepath.Join(preflightDocsPath, "playbooks", "work.md"), []byte("# work\n"), 0o644); err != nil {
		t.Fatalf("WriteFile work playbook: %v", err)
	}
	planPath := filepath.Join("docs", "plans", "active", "initiative.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("MkdirAll active plans: %v", err)
	}
	if err := os.WriteFile(planPath, []byte("# active plan\n"), 0o644); err != nil {
		t.Fatalf("WriteFile active plan: %v", err)
	}
}

func seedPreflightStory(t *testing.T, status string) {
	t.Helper()
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
	if _, _, err := infrastructure.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO stories (id, slug, goal, status, created_at)
		VALUES ('01HPREFLIGHTSTORYSEEDULID', 'seeded', 'goal', ?, '2026-08-07T00:00:00Z')`, status); err != nil {
		t.Fatalf("seed story status=%s: %v", status, err)
	}
}

func TestPreflightCommandBlocksDurableWithoutHarness(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "to-plan", "--mode", "full", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Stop == nil || view.Stop.Code != "harness_required" {
		t.Fatalf("view stop = %+v, want harness_required", view.Stop)
	}
}

func TestPreflightCommandReadsReadyStateWithoutMutation(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join(preflightDocsPath, "playbooks"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	workPlaybook := filepath.Join(preflightDocsPath, "playbooks", "work.md")
	if err := os.WriteFile(workPlaybook, []byte("# work\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	beforeDB, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("ReadFile() db error = %v", err)
	}
	beforeHash := sha256.Sum256(beforeDB)
	beforeChangesets := listPreflightChangesets(t)
	if beforeMarkdown := listBoundedLifecycleMarkdown(t); len(beforeMarkdown) != 0 {
		t.Fatalf("bounded lifecycle markdown fixture = %v, want none", beforeMarkdown)
	}

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "bounded", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Stage != "work" || view.Mode != "reduced" || view.Readiness != application.PreflightReady || view.DB != application.PreflightDBReady || view.Docs != application.PreflightDocsReady {
		t.Fatalf("view = %+v, want reduced/ready work route", view)
	}
	if view.Stop != nil {
		t.Fatalf("view stop = %+v, want none", view.Stop)
	}
	if view.Playbook != workPlaybook {
		t.Fatalf("view playbook = %q, want %q", view.Playbook, workPlaybook)
	}

	afterDB, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("ReadFile() db after preflight error = %v", err)
	}
	afterHash := sha256.Sum256(afterDB)
	if beforeHash != afterHash {
		t.Fatalf("preflight mutated closed db: before=%x after=%x", beforeHash, afterHash)
	}
	afterChangesets := listPreflightChangesets(t)
	if len(beforeChangesets) != len(afterChangesets) {
		t.Fatalf("preflight changeset count = %d, want %d", len(afterChangesets), len(beforeChangesets))
	}
	if !slices.Equal(beforeChangesets, afterChangesets) {
		t.Fatalf("preflight changesets = %v, want %v", afterChangesets, beforeChangesets)
	}
	if afterMarkdown := listBoundedLifecycleMarkdown(t); len(afterMarkdown) != 0 {
		t.Fatalf("preflight created bounded lifecycle markdown = %v", afterMarkdown)
	}
}

func TestPreflightCommandReportsStaleDocs(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Join(preflightDocsPath, "playbooks"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(preflightDocsPath, "playbooks", "check.md"), []byte("# check\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := db.Exec(`UPDATE meta SET docs_version = '0.5.0'`); err != nil {
		db.Close()
		t.Fatalf("set docs version: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	out, err := runPreflightCommand(t, "0.6.0", "preflight", "check", "--mode", "full", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Docs != application.PreflightDocsStale || view.Stop == nil || view.Stop.Code != "stale_docs" {
		t.Fatalf("view = %+v", view)
	}
}

func TestPreflightCommandIncludesVersion(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "1.2.3", "preflight", "watzup", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Version != "1.2.3" {
		t.Fatalf("view.Version = %q, want %q", view.Version, "1.2.3")
	}
}

// TestPreflightCommandWatzupIncludesContextWithoutPhases proves R4/wave 1:
// watzup gets a stage-shaped packet (position/latest IDs/drift/readiness),
// but not the Phases field its playbook never references.
func TestPreflightCommandWatzupIncludesContextWithoutPhases(t *testing.T) {
	t.Chdir(t.TempDir())
	seedPreflightStory(t, "planned")

	out, err := runPreflightCommand(t, "dev", "preflight", "watzup", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Context == nil {
		t.Fatal("watzup view.Context = nil, want a stage-shaped packet")
	}
	if view.Context.Phases != nil {
		t.Fatalf("watzup view.Context.Phases = %v, want nil (watzup.md never calls query phases)", view.Context.Phases)
	}
}

// TestPreflightCommandWorkIncludesContextWithPhases proves work's packet
// carries the phases list — the field its playbook's "Load state" step
// references via `query phases`.
func TestPreflightCommandWorkIncludesContextWithPhases(t *testing.T) {
	t.Chdir(t.TempDir())
	seedPreflightWorkPlaybookAndPlan(t)
	seedPreflightStory(t, "in-progress")

	out, err := runPreflightCommand(t, "dev", "preflight", "work", "--mode", "auto", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Context == nil {
		t.Fatal("work view.Context = nil, want a stage-shaped packet")
	}
	if len(view.Context.Phases) != 1 || view.Context.Phases[0].Slug != "seeded" {
		t.Fatalf("work view.Context.Phases = %+v, want the seeded story", view.Context.Phases)
	}
}

// TestPreflightCommandCheckReviewAndBoundedOmitContext proves check's
// response-only modes stay packet-free: review/bounded never call
// `check record`/touch the plan, so they have no lifecycle position to
// prefetch and must not pay the packet's cost.
func TestPreflightCommandCheckReviewAndBoundedOmitContext(t *testing.T) {
	for _, mode := range []string{"review", "bounded", "simple"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir(t.TempDir())
			seedPreflightStory(t, "planned")

			out, err := runPreflightCommand(t, "dev", "preflight", "check", "--mode", mode, "--json")
			if err != nil {
				t.Fatalf("preflight command error = %v", err)
			}
			view := decodePreflight(t, out)
			if view.Context != nil {
				t.Fatalf("check --mode %s view.Context = %+v, want nil (response-only, no lifecycle position to prefetch)", mode, view.Context)
			}
		})
	}
}

// TestPreflightCommandCheckGateAndFullIncludeContextWithPhases is R6's
// routing proof (docs/audit/sdlc-token-cache-audit.md): check's durable
// gate/full modes now get the same stage-shaped packet work/handoff
// already receive, replacing check.md step 1's separate `zharness resume
// --json` call — superseding this initiative's own earlier NG2, which
// kept check's reads entirely separate (docs/audit/workflow-harness-
// ceremony-audit.md).
func TestPreflightCommandCheckGateAndFullIncludeContextWithPhases(t *testing.T) {
	for _, mode := range []string{"gate", "full"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir(t.TempDir())
			seedPreflightWorkPlaybookAndPlan(t)
			if err := os.WriteFile(filepath.Join(preflightDocsPath, "playbooks", "check.md"), []byte("# check\n"), 0o644); err != nil {
				t.Fatalf("WriteFile check playbook: %v", err)
			}
			seedPreflightStory(t, "in-progress")

			out, err := runPreflightCommand(t, "dev", "preflight", "check", "--mode", mode, "--json")
			if err != nil {
				t.Fatalf("preflight command error = %v", err)
			}
			view := decodePreflight(t, out)
			if view.Context == nil {
				t.Fatalf("check --mode %s view.Context = nil, want a stage-shaped packet", mode)
			}
			if len(view.Context.Phases) != 1 || view.Context.Phases[0].Slug != "seeded" {
				t.Fatalf("check --mode %s view.Context.Phases = %+v, want the seeded story", mode, view.Context.Phases)
			}
		})
	}
}

// TestPreflightCommandContextNilWithoutHarness proves the packet degrades
// gracefully (not a crash or a bogus empty packet) when there is no DB to
// build it from.
func TestPreflightCommandContextNilWithoutHarness(t *testing.T) {
	t.Chdir(t.TempDir())

	out, err := runPreflightCommand(t, "dev", "preflight", "watzup", "--json")
	if err != nil {
		t.Fatalf("preflight command error = %v", err)
	}
	view := decodePreflight(t, out)
	if view.Context != nil {
		t.Fatalf("watzup view.Context = %+v, want nil without a harness db", view.Context)
	}
}

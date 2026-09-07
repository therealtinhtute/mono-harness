package embedded

import (
	"strings"
	"testing"
)

func TestBuildManifest_PathsPresentAndNonEmpty(t *testing.T) {
	m, err := BuildManifest("dev")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if len(m.Paths) == 0 {
		t.Fatal("manifest has no paths")
	}
	for _, p := range m.Paths {
		data, err := FS.ReadFile(p)
		if err != nil {
			t.Errorf("ReadFile(%s): %v", p, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("%s is empty", p)
		}
	}
}

func TestPlaybookCount(t *testing.T) {
	count, err := PlaybookCount()
	if err != nil {
		t.Fatalf("PlaybookCount: %v", err)
	}
	if count != 6 {
		t.Fatalf("playbook count = %d, want 6", count)
	}
}

func TestBuildManifest_EntrypointAndWorkflowPresent(t *testing.T) {
	m, err := BuildManifest("dev")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	want := []string{"AGENTS.md", "WORKFLOW.md"}
	for _, w := range want {
		found := false
		for _, p := range m.Paths {
			if p == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("manifest missing %s", w)
		}
	}
}

func TestBuildManifest_DocsVersionExposed(t *testing.T) {
	m, err := BuildManifest("0.2.0")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}
	if m.DocsVersion != "0.2.0" {
		t.Fatalf("DocsVersion = %q, want %q", m.DocsVersion, "0.2.0")
	}
}

// legacyDBName is folded at compile time; spelled this way so the S4 tree
// S4 tree scan does not match its own guard list.
const legacyDBName = "harness" + ".db"

func TestOnePlan_PlaybookContract(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		required  []string
		forbidden []string
	}{
		{
			name: "brainstorm locks honest bootstrap state",
			path: "playbooks/brainstorm.md",
			required: []string{
				"## Outcome",
				"## Authority and Requirements",
				"## Non-goals",
				"The canonical active path is `docs/plans/active/{slug}.md`",
				"Mint two unique identifier tokens locally",
				"Confirm at most one non-empty plan exists under `docs/plans/active/`",
				"explore creates no lifecycle rows, plans, reports, changesets, or markdown artifacts",
				"approach: not-planned",
				"planning_status: not-planned",
				"exact_next_action: to-plan",
				"no literal fake lifecycle placeholders",
				"those definitions are immutable",
				"Append-only `## Progress` is the sole task execution-status source",
			},
		},
		{
			name: "to-plan defines phases as markdown truth",
			path: "playbooks/to-plan.md",
			required: []string{
				"## Approach and Risks",
				"## Phases and Verification",
				"mint a stable `story_id`",
				"one `story_id` per listed phase",
				"docs/plans/active/{slug}.md",
				"After a phase/task definition is written, it is immutable",
				"work/check/handoff may change only that phase's lifecycle status in this file",
				"Do not add task status fields",
				"never write parallel task/phase state anywhere else",
				"Append-only `## Progress` is the sole task execution-status source",
			},
		},
		{
			name: "work appends durable markdown progress",
			path: "playbooks/work.md",
			required: []string{
				"## Progress",
				"## Decisions",
				"## Current State and Next Action",
				"slice `docs/plans/active/{slug}.md` by section",
				"set that phase's plan status to `in-progress`",
				"`task_status=in-progress`",
				"flushes the whole pending list immediately",
				"bounded/simple mode creates no lifecycle rows, plans, reports, changesets, or markdown artifacts",
				"Do not add or update task-definition `status` fields",
				"Append-only `## Progress` is the sole task execution-status source",
			},
		},
		{
			name: "check preserves review intent and records evidence",
			path: "playbooks/check.md",
			required: []string{
				"## Validation",
				"Invocation intent wins",
				"`review` is always response-only",
				"never appends to Validation",
				"set the phase status and Current State lifecycle status to `checked`",
				"keep the phase and Current State lifecycle status `in-progress`",
				"Durable `gate` runs automated checks",
				"`full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review",
				"`gate` does not perform that complete manual review",
				"Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output",
				"so the commit-time guard can find them",
				"The repository's pre-commit hook is the sole proof guarantee",
				"REQUEST_CHANGES entries may cite deliberately failing commands",
				"Append-only `## Progress` is the sole task execution-status source",
			},
			forbidden: []string{
				"Durable `gate`/`full` mode runs real checks and review",
				"Gate/full: applicable commands have captured output, alignment and code review ran",
			},
		},
		{
			name: "handoff closes phases before initiatives",
			path: "playbooks/handoff.md",
			required: []string{
				"## Current State and Next Action",
				"Close every cleanly checked phase",
				"keep frontmatter `status: active` and the same active path",
				"Before closing the final phase, require every prior phase to be `done`",
				"Only after the plan shows every phase `done`",
				"git mv docs/plans/active/{slug}.md docs/plans/completed/{slug}.md",
				"exactly one file may represent the initiative afterwards",
				"docs/plans/active/{slug}.md",
				"docs/plans/completed/{slug}.md",
				"Preserve every phase/task definition",
				"Append-only `## Progress` is the sole task execution-status source",
			},
			forbidden: []string{
				"zharness preflight handoff --mode full --json",
				"Require every phase to be `done` or the final phase",
			},
		},
		{
			name: "watzup recaps without writing",
			path: "playbooks/watzup.md",
			required: []string{
				"Select the active plan by name",
				"exactly one non-empty file may exist under `docs/plans/active/*.md`",
				"Remain read-only",
				"docs/plans/active/{slug}.md",
				"never the whole file",
				"Read task execution status only from append-only `## Progress`",
				"the sole task execution-status source",
			},
			forbidden: []string{
				"zharness resume --json",
			},
		},
	}

	retired := []string{
		".kit/planning",
		"SPEC.md",
		"ROADMAP.md",
		"-CONTEXT.md",
		"-PLAN.md",
		".kit/runs",
		".kit/reports",
		".kit/HANDOFF.md",
		"implementation-notes",
		"--artifact-path",
		"zharness scaffold run",
		"zharness scaffold check",
		"zharness scaffold handoff",
		"RUN artifact",
		"CHECK report",
		// v0.15 slim (R1/R6): every lifecycle verb is deleted from source, and
		// the playbooks are markdown-first with no index-sync block. These
		// strings must never reappear in any embedded playbook.
		"zharness preflight",
		"zharness run create",
		"zharness trace add",
		"zharness decision add",
		"zharness story",
		"zharness intake",
		"zharness id",
		"zharness scaffold plan",
		"zharness query",
		"zharness audit",
		"zharness validate",
		"zharness resume",
		"zharness init",
		"zharness migrate",
		"zharness import",
		"zharness db ",
		"zharness memory",
		"zharness plan complete",
		"zharness handoff record",
		"check record",
		"--close-phase",
		"Optional index-sync",
		legacyDBName,
		"lifecycle ledger",
		"DB-mirroring",
		"mirrored check row",
		"check_id: ULID",
		"latest_run_id",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := FS.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", tt.path, err)
			}
			content := string(data)

			for _, phrase := range tt.required {
				if !strings.Contains(content, phrase) {
					t.Errorf("%s missing one-plan contract phrase %q", tt.path, phrase)
				}
			}
			for _, phrase := range append(retired, tt.forbidden...) {
				if strings.Contains(content, phrase) {
					t.Errorf("%s contains forbidden lifecycle contract %q", tt.path, phrase)
				}
			}
		})
	}
}

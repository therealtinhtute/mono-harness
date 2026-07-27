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
				"zharness preflight brainstorm --mode {explore|lock} --json",
				"zharness scaffold plan --path docs/plans/active/{slug}.md --json",
				"--plan-path docs/plans/active/{slug}.md",
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
			name: "to-plan creates synchronized planned phases",
			path: "playbooks/to-plan.md",
			required: []string{
				"## Approach and Risks",
				"## Phases and Verification",
				"zharness preflight to-plan --mode full --json",
				"zharness story --slug {stable-phase-slug}",
				"set its plan status to `planned`, matching the new DB row",
				"zharness query phases --json",
				"docs/plans/active/{slug}.md",
				"After a phase/task definition is written, it is immutable",
				"only phase lifecycle status to mirror their DB transitions",
				"Do not add task status fields",
				"Append-only `## Progress` is the sole task execution-status source",
			},
		},
		{
			name: "work synchronizes run transition",
			path: "playbooks/work.md",
			required: []string{
				"## Progress",
				"## Decisions",
				"## Current State and Next Action",
				"zharness preflight work --mode {full|bounded} --json",
				"zharness run create --slug {stable-phase-slug} --plan-id {plan-id} --json",
				"set that phase's plan status to `in-progress`",
				"zharness trace add --wave {N}",
				"bounded/simple mode creates no lifecycle rows, plans, reports, changesets, or markdown artifacts",
				"Do not add or update task-definition `status` fields",
				"Append-only `## Progress` is the sole task execution-status source",
				"`task_status=in-progress`",
			},
		},
		{
			name: "check preserves review intent and synchronizes gate",
			path: "playbooks/check.md",
			required: []string{
				"## Validation",
				"zharness preflight check --mode {gate|full|review|bounded} --json",
				"Invocation intent wins",
				"`review` is always response-only",
				"never calls `zharness check record`",
				"set the phase status and Current State lifecycle status to `checked`",
				"keep the phase and Current State lifecycle status `in-progress` to match the DB",
				"zharness audit --json",
				"zharness check record --verdict {verdict} --run-id {run-id}",
				"Durable `gate` runs automated checks",
				"`full` includes the gate and adds the complete Security, Performance, Architecture, and Code Quality review",
				"`gate` does not perform that complete manual review",
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
				"zharness preflight handoff --json",
				"Close every cleanly checked phase",
				"keep frontmatter `status: active` and the same active path",
				"Before closing the final phase, require every prior phase to be `done`",
				"Only after `zharness query phases --json` shows every phase `done`",
				"--close-phase",
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
				"zharness preflight watzup --json",
				"zharness resume --json",
				"Remain read-only",
				"docs/plans/active/{slug}.md",
				"Read task execution status only from append-only `## Progress`",
				"the sole task execution-status source",
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

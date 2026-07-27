package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/embedded"
)

func TestScaffoldArtifact_EachKind(t *testing.T) {
	for _, kind := range []string{"run", "check", "handoff", "spec", "plan"} {
		dst := filepath.Join(t.TempDir(), kind+".md")
		data, err := ScaffoldArtifact(embedded.Templates, kind, dst)
		if err != nil {
			t.Fatalf("kind %s: %v", kind, err)
		}
		if len(data) == 0 {
			t.Fatalf("kind %s: empty template", kind)
		}
		onDisk, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("kind %s: file not written: %v", kind, err)
		}
		if !strings.Contains(string(onDisk), "type: "+kind) {
			t.Fatalf("kind %s: skeleton missing `type: %s` frontmatter", kind, kind)
		}
	}
}

func TestOnePlan_ScaffoldTemplate(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "initiative.md")
	data, err := ScaffoldArtifact(embedded.Templates, "plan", dst)
	if err != nil {
		t.Fatalf("scaffold plan: %v", err)
	}
	content := string(data)

	frontmatter := []string{
		"id: {PLAN_ULID}",
		"type: plan",
		"intake_id: {INTAKE_ULID}",
		"lane: {tiny|normal|high-risk}",
		"status: {active|completed}",
		"created: {YYYY-MM-DD}",
		"updated: {YYYY-MM-DD}",
	}
	for _, phrase := range frontmatter {
		if !strings.Contains(content, phrase) {
			t.Errorf("plan template missing frontmatter contract %q", phrase)
		}
	}

	headings := []string{
		"## Outcome",
		"## Authority and Requirements",
		"## Non-goals",
		"## Approach and Risks",
		"## Phases and Verification",
		"## Progress",
		"## Decisions",
		"## Validation",
		"## Current State and Next Action",
	}
	if got := strings.Count(content, "\n## "); got != len(headings) {
		t.Fatalf("plan template has %d second-level headings, want %d", got, len(headings))
	}
	for _, heading := range headings {
		if !strings.Contains(content, heading) {
			t.Errorf("plan template missing heading %q", heading)
		}
	}

	structured := []string{
		"approach: not-planned",
		"Phase and task definitions are immutable after to-plan",
		"Do not add task status fields",
		"Append-only Progress is the sole task execution-status source",
		"Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done",
		"phase_slug",
		"story_id",
		"status",
		"depends_on",
		"waves",
		"tasks",
		"checks",
		"planning_status: not-planned",
		"phases: none",
		"Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id",
		"Append-only durable entries record timestamp, phase/task, decision, and rationale",
		"exact command/result/output, run_id, check_id, verdict, and proof_gaps",
		"active_phase: none",
		"lifecycle_status: not-planned",
		"latest_run_id: none",
		"latest_trace_ids: []",
		"latest_check_id: none",
		"latest_handoff_id: none",
		"blockers: none",
		"exact_next_action: to-plan",
	}
	for _, phrase := range structured {
		if !strings.Contains(content, phrase) {
			t.Errorf("plan template missing lifecycle contract %q", phrase)
		}
	}

	if strings.Contains(content, "\n        status:") {
		t.Error("plan template contains a mutable task-definition status field")
	}

	for _, phrase := range []string{
		"{stable-phase-slug}",
		"{ULID returned by zharness story}",
		"{RFC3339 timestamp}",
		"{ULID|none}",
		"{exact verification command}",
		"type: run",
		"type: check",
		"type: handoff",
		"RUN artifact",
		"CHECK report",
		"HANDOFF.md",
		".kit/runs",
		".kit/reports",
	} {
		if strings.Contains(content, phrase) {
			t.Errorf("plan template contains fake or dedicated lifecycle output %q", phrase)
		}
	}
}

func TestScaffoldArtifact_UnknownKind(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "x.md")
	_, err := ScaffoldArtifact(embedded.Templates, "bogus", dst)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_kind" {
		t.Fatalf("want unknown_kind ValidationError, got %v", err)
	}
	if _, statErr := os.Stat(dst); statErr == nil {
		t.Fatalf("unknown kind should not write a file")
	}
}

func TestScaffoldArtifact_EmptyPath(t *testing.T) {
	_, err := ScaffoldArtifact(embedded.Templates, "run", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("want missing_required_field ValidationError, got %v", err)
	}
}

func TestScaffoldArtifact_RefuseOverwriteNonEmpty(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "run.md")
	if err := os.WriteFile(dst, []byte("half-filled work in progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ScaffoldArtifact(embedded.Templates, "run", dst)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "file_exists" {
		t.Fatalf("want file_exists ValidationError, got %v", err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "half-filled work in progress\n" {
		t.Fatalf("existing content was clobbered: %q", string(got))
	}
}

func TestScaffoldArtifact_OverwriteEmptyPlaceholder(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "run.md")
	if err := os.WriteFile(dst, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScaffoldArtifact(embedded.Templates, "run", dst); err != nil {
		t.Fatalf("empty placeholder should be fillable: %v", err)
	}
	got, _ := os.ReadFile(dst)
	if !strings.Contains(string(got), "type: run") {
		t.Fatalf("skeleton not written over empty placeholder")
	}
}

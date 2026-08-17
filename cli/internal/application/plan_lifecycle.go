package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

const activePlanCompletedDir = "docs/plans/completed"

// PlanComplete resolves the single active plan (R2,
// docs/audit/consumer-adoption-audit.md D1), refuses when any phase it
// defines is not yet a done story, and otherwise moves it out of
// docs/plans/active/ with status: completed — the first of D1's two exits
// from the "at most one active plan" invariant.
func PlanComplete(db *sql.DB) (string, *StopInfo, error) {
	plan, stop, err := ResolveActivePlan()
	if err != nil {
		return "", nil, err
	}
	if stop != nil {
		return "", stop, nil
	}

	slugs := parseActivePlanPhaseOrder(plan.content)
	openSlug, err := selectActivePhase(db, slugs)
	if err != nil {
		return "", nil, err
	}
	if openSlug != "" {
		return "", nil, &domain.ValidationError{
			Code: "open_phase",
			Message: fmt.Sprintf(
				"%s: phase %q is not a done story: plan complete requires every phase_slug the plan defines to be done (docs/audit/consumer-adoption-audit.md, D1). Finish %q, or run `zharness plan abandon --reason ...` if this plan will never ship.",
				plan.path, openSlug, openSlug,
			),
		}
	}

	dest, err := transitionActivePlan(plan, "completed", "plan completed. rationale: every phase_slug is a done story.")
	if err != nil {
		return "", nil, err
	}
	return dest, nil, nil
}

// PlanAbandon is D1's second exit: the same move with status: abandoned,
// gated on a required --reason instead of an open-phase check — an
// abandoned plan is, by definition, not being finished.
func PlanAbandon(reason string) (string, *StopInfo, error) {
	if strings.TrimSpace(reason) == "" {
		return "", nil, &domain.ValidationError{
			Code:    "missing_required_field",
			Message: "plan abandon: --reason is required",
		}
	}

	plan, stop, err := ResolveActivePlan()
	if err != nil {
		return "", nil, err
	}
	if stop != nil {
		return "", stop, nil
	}

	dest, err := transitionActivePlan(plan, "abandoned", fmt.Sprintf("plan abandoned. rationale: %s.", reason))
	if err != nil {
		return "", nil, err
	}
	return dest, nil, nil
}

// transitionActivePlan rewrites plan's frontmatter status/updated fields,
// records the transition as a `## Decisions` entry (the plan's own durable
// record — P0 has no plan_index/DB table yet, see docs/plans/active
// /harness-markdown-truth.md P2/R8), and moves the file from
// docs/plans/active/ to docs/plans/completed/.
func transitionActivePlan(plan activePlan, status, decisionEntry string) (string, error) {
	content := setFrontmatterFields(plan.content, map[string]string{
		"status":  status,
		"updated": time.Now().UTC().Format("2006-01-02"),
	})

	entry := fmt.Sprintf("- `%s` — %s", time.Now().UTC().Format(time.RFC3339), decisionEntry)
	content, err := AppendToPlanSection(content, "Decisions", entry)
	if err != nil {
		return "", fmt.Errorf("plan %s: record transition: %w", status, err)
	}

	if err := os.MkdirAll(activePlanCompletedDir, 0o755); err != nil {
		return "", fmt.Errorf("plan %s: mkdir %s: %w", status, activePlanCompletedDir, err)
	}
	dest := filepath.Join(activePlanCompletedDir, filepath.Base(plan.path))
	if err := writeFileAtomically(dest, []byte(content)); err != nil {
		return "", fmt.Errorf("plan %s: write %s: %w", status, dest, err)
	}
	if err := os.Remove(plan.path); err != nil {
		return "", fmt.Errorf("plan %s: remove %s: %w", status, plan.path, err)
	}
	return dest, nil
}

// setFrontmatterFields rewrites `key: value` lines strictly between the
// first two `---` delimiters at the top of content. A key not already
// present in the frontmatter block is left absent — the plan template
// guarantees status/updated exist, so silently inventing missing keys
// would only mask a malformed plan. Lines outside the frontmatter block
// (e.g. a phase's own `- status: planned` list item) are never touched.
func setFrontmatterFields(content string, fields map[string]string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return content
	}
	for i := 1; i < end; i++ {
		for key, value := range fields {
			if strings.HasPrefix(lines[i], key+":") {
				lines[i] = fmt.Sprintf("%s: %s", key, value)
			}
		}
	}
	return strings.Join(lines, "\n")
}

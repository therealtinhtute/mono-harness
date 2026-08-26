package application

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func createValidRetainedLifecycle(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()

	intakeID, err := CreateIntake(db, domain.IntakeNewSpec, "validate retained entity IDs", domain.LaneNormal, "", "")
	if err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	storyID, err := CreateStory(db, "retained-entity-ids", "validate retained entity IDs", "")
	if err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, err := CreateRun(db, "retained-entity-ids", ".kit/runs/work/legacy.md", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	traceID, err := CreateTrace(db, 1, "validate retained entity IDs", runID, "", "")
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	checkID, err := RecordCheck(db, runID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{
		{Command: "true", OutputRef: "PASS"},
	}, "")
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	handoffID, err := RecordHandoff(db, runID, checkID, "", nil, false)
	if err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}

	return map[string]string{
		"story":   storyID,
		"run":     runID,
		"check":   checkID,
		"handoff": handoffID,
		"intake":  intakeID,
		"trace":   traceID,
	}
}

func TestValidateWithoutDatabaseIsInvalid(t *testing.T) {
	result, err := Validate(nil)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false when no database is available")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "DB->LIFECYCLE" {
		t.Fatalf("findings = %v, want one DB->LIFECYCLE finding", result.Findings)
	}
}

func TestValidateReportsOverflowLifecycleID(t *testing.T) {
	db := freshDB(t)
	storyID, err := CreateStory(db, "overflow-id", "reject an overflowing ULID", "")
	if err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	const overflowID = "ZZZZZZZZZZZZZZZZZZZZZZZZZZ"
	if _, err := db.Exec(`UPDATE stories SET id = ? WHERE id = ?`, overflowID, storyID); err != nil {
		t.Fatalf("update story id: %v", err)
	}

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	want := ValidateFinding{
		Link:   "DB->STORY",
		Issue:  "missing_key",
		Detail: `story id "ZZZZZZZZZZZZZZZZZZZZZZZZZZ" is not a valid ULID`,
	}
	if result.Valid {
		t.Fatal("valid = true, want false for an overflowing lifecycle ULID")
	}
	if len(result.Findings) != 1 || result.Findings[0] != want {
		t.Fatalf("findings = %v, want [%+v]", result.Findings, want)
	}
}

func TestValidateReportsInvalidIDsForRetainedTables(t *testing.T) {
	invalidIDs := []struct {
		name string
		id   string
	}{
		{name: "malformed", id: "not-a-ulid"},
		{name: "overflow", id: "ZZZZZZZZZZZZZZZZZZZZZZZZZZ"},
	}
	entities := []struct {
		name  string
		table string
		link  string
	}{
		{name: "intake", table: "intakes", link: "DB->INTAKE"},
		{name: "trace", table: "traces", link: "DB->TRACE"},
	}

	for _, invalid := range invalidIDs {
		for _, entity := range entities {
			t.Run(invalid.name+"/"+entity.name, func(t *testing.T) {
				db := freshDB(t)
				ids := createValidRetainedLifecycle(t, db)
				if _, err := db.Exec("UPDATE "+entity.table+" SET id = ? WHERE id = ?", invalid.id, ids[entity.name]); err != nil {
					t.Fatalf("update %s id: %v", entity.name, err)
				}

				result, err := Validate(db)
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				want := ValidateFinding{
					Link:   entity.link,
					Issue:  "missing_key",
					Detail: entity.name + " id \"" + invalid.id + "\" is not a valid ULID",
				}
				if result.Valid {
					t.Fatalf("valid = true, want false for %s %s ID", invalid.name, entity.name)
				}
				if len(result.Findings) != 1 || result.Findings[0] != want {
					t.Fatalf("findings = %v, want [%+v]", result.Findings, want)
				}
			})
		}
	}
}

func TestValidateReportsInvalidLifecycleEnums(t *testing.T) {
	db := freshDB(t)
	ids := createValidRetainedLifecycle(t, db)
	if _, err := db.Exec(`UPDATE stories SET status = 'bogus' WHERE id = ?`, ids["story"]); err != nil {
		t.Fatalf("corrupt story status: %v", err)
	}
	if _, err := db.Exec(`UPDATE checks SET verdict = 'bogus' WHERE id = ?`, ids["check"]); err != nil {
		t.Fatalf("corrupt check verdict: %v", err)
	}

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	want := []ValidateFinding{
		{Link: "DB->STORY", Issue: "invalid_value", Detail: `story "retained-entity-ids" has invalid status "bogus"`},
		{Link: "DB->CHECK", Issue: "invalid_value", Detail: "check " + ids["check"] + ` has invalid verdict "bogus"`},
	}
	if result.Valid {
		t.Fatal("valid = true, want false for invalid persisted enums")
	}
	if len(result.Findings) != len(want) {
		t.Fatalf("findings = %v, want %v", result.Findings, want)
	}
	for i := range want {
		if result.Findings[i] != want[i] {
			t.Fatalf("finding[%d] = %+v, want %+v", i, result.Findings[i], want[i])
		}
	}
}

func TestValidateCleanDatabaseIgnoresLegacyArtifactPaths(t *testing.T) {
	db := freshDB(t)
	createValidRetainedLifecycle(t, db)

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid || len(result.Findings) != 0 {
		t.Fatalf("result = %+v, want valid lifecycle with no findings", result)
	}
}

func TestValidateReportsBrokenDatabaseLink(t *testing.T) {
	db := freshDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	runID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO runs (id, story_slug, trace_ids, artifact_path, created_at)
		VALUES (?, 'missing-story', '[]', '', '2026-07-27T12:00:00Z')
	`, runID); err != nil {
		t.Fatalf("insert broken run: %v", err)
	}

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false for a broken run-to-story link")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "STORY->RUN" || result.Findings[0].Issue != "broken_link" {
		t.Fatalf("findings = %v, want one STORY->RUN broken_link", result.Findings)
	}
}

func TestValidateReportsHandoffRunCheckMismatch(t *testing.T) {
	db := freshDB(t)
	checkRunID := createLifecycleRun(t, db, "cli-domain")
	checkID, err := RecordCheck(db, checkRunID, domain.VerdictApproved, domain.JudgeIndependent, "test-model", []domain.ProofLink{{Command: "true"}}, "")
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}

	handoffRunID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO runs (id, story_slug, trace_ids, artifact_path, created_at)
		VALUES (?, 'cli-domain', '[]', '', '2026-07-27T12:01:00Z')
	`, handoffRunID); err != nil {
		t.Fatalf("insert second run: %v", err)
	}
	handoffID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO handoffs (id, run_id, check_id, anchors, created_at)
		VALUES (?, ?, ?, '{}', '2026-07-27T12:02:00Z')
	`, handoffID, handoffRunID, checkID); err != nil {
		t.Fatalf("insert mismatched handoff: %v", err)
	}

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false for a handoff whose check gates another run")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "CHECK->HANDOFF" || result.Findings[0].Issue != "broken_link" {
		t.Fatalf("findings = %v, want one CHECK->HANDOFF broken_link", result.Findings)
	}
}

// check 9 (R13): zero active plans is a valid idle state and reports no
// finding — the designed result of plan complete/abandon before a new plan
// is locked (matches TestLifecycle_ScratchDirFullChain's post-completion
// validate assertion in cmd/zharness).
func TestValidateNoActivePlanIsClean(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.Valid || len(result.Findings) != 0 {
		t.Fatalf("result = %+v, want valid with no findings for zero active plans", result)
	}
}

// check 9 (R13): 2+ active plans is the actual defect — ambiguity, not
// absence.
func TestValidateReportsAmbiguousActivePlans(t *testing.T) {
	chdirFixture(t)
	writeFile(t, "docs/plans/active/fixture-plan-a.md", scaffoldedPlanFixture)
	writeFile(t, "docs/plans/active/fixture-plan-b.md", scaffoldedPlanFixture)
	db := freshDB(t)

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false with two active plans")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "DOCS->PLAN" || result.Findings[0].Issue != "invalid_value" ||
		!strings.Contains(result.Findings[0].Detail, "contains 2 plans") {
		t.Fatalf("findings = %v, want one DOCS->PLAN invalid_value naming 2 plans", result.Findings)
	}
}

// check 10: every required frontmatter key must carry a non-empty value.
func TestValidateReportsMissingFrontmatterKey(t *testing.T) {
	chdirFixture(t)
	content := strings.Replace(scaffoldedPlanFixture, "lane: normal\n", "", 1)
	writeFile(t, "docs/plans/active/fixture-plan.md", content)
	db := freshDB(t)

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false with a missing frontmatter key")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "DOCS->PLAN" || result.Findings[0].Issue != "missing_key" ||
		!strings.Contains(result.Findings[0].Detail, `"lane"`) {
		t.Fatalf("findings = %v, want one DOCS->PLAN missing_key naming \"lane\"", result.Findings)
	}
}

// check 11: every phase_slug the plan defines must have a matching story row.
func TestValidateReportsPhaseWithoutStory(t *testing.T) {
	chdirFixture(t)
	content := strings.Replace(scaffoldedPlanFixture, "## Progress",
		"## Phases and Verification\n- phase_slug: orphan-phase\n\n## Progress", 1)
	writeFile(t, "docs/plans/active/fixture-plan.md", content)
	db := freshDB(t)

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false with a phase that has no story row")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "STORY->PLAN" || result.Findings[0].Issue != "broken_link" ||
		!strings.Contains(result.Findings[0].Detail, `"orphan-phase"`) {
		t.Fatalf("findings = %v, want one STORY->PLAN broken_link naming \"orphan-phase\"", result.Findings)
	}
}

// check 12: every heading past the structural section must be one of the
// four known append-only headings.
func TestValidateReportsUnknownPlanHeading(t *testing.T) {
	chdirFixture(t)
	content := scaffoldedPlanFixture + "\n## Bogus Heading\n- some text\n"
	writeFile(t, "docs/plans/active/fixture-plan.md", content)
	db := freshDB(t)

	result, err := Validate(db)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.Valid {
		t.Fatal("valid = true, want false with an unrecognized plan heading")
	}
	if len(result.Findings) != 1 || result.Findings[0].Link != "DOCS->PLAN" || result.Findings[0].Issue != "invalid_value" ||
		!strings.Contains(result.Findings[0].Detail, "Bogus Heading") {
		t.Fatalf("findings = %v, want one DOCS->PLAN invalid_value naming \"Bogus Heading\"", result.Findings)
	}
}

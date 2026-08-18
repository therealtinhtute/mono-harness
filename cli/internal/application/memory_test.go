package application

import (
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

func TestCreateMemoryGlobal(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	id, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "sqlite WAL sidecars are read-only-command sensitive")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	assertRowExists(t, db, "memories", id)

	path := "docs/memory/" + id + ".md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), "scope: global") {
		t.Fatalf("memory entry missing scope: global, got %q", content)
	}
	if strings.Contains(string(content), "plan_id:") {
		t.Fatalf("global memory entry must not carry plan_id, got %q", content)
	}

	var gotPath, gotType, gotScope string
	var gotPlanID sql.NullString
	if err := db.QueryRow(`SELECT path, type, scope, plan_id FROM memories WHERE id = ?`, id).Scan(&gotPath, &gotType, &gotScope, &gotPlanID); err != nil {
		t.Fatalf("query memories row: %v", err)
	}
	if gotPath != path || gotType != "gotcha" || gotScope != domain.MemoryScopeGlobal || gotPlanID.Valid {
		t.Fatalf("memories row = (%q, %q, %q, %v), want (%q, gotcha, global, invalid)", gotPath, gotType, gotScope, gotPlanID, path)
	}
}

func TestCreateMemoryPlanScoped(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	id, err := CreateMemory(db, "decision", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX", "chose markdown-first over DB-only storage")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	assertRowExists(t, db, "memories", id)

	var gotPlanID string
	if err := db.QueryRow(`SELECT plan_id FROM memories WHERE id = ?`, id).Scan(&gotPlanID); err != nil {
		t.Fatalf("query plan_id: %v", err)
	}
	if gotPlanID != "01PLANIDFIXTUREXXXXXXXXXX" {
		t.Fatalf("plan_id = %q, want fixture plan id", gotPlanID)
	}
}

func TestCreateMemoryPlanScopeRequiresPlanID(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	_, err := CreateMemory(db, "decision", domain.MemoryScopePlan, "", "missing plan id")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
	if got := countRows(t, db, "memories"); got != 0 {
		t.Fatalf("memories rows = %d, want 0", got)
	}
}

func TestCreateMemoryInvalidScope(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	_, err := CreateMemory(db, "gotcha", "bogus", "", "summary")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "invalid_scope" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: invalid_scope}", err)
	}
	if got := countRows(t, db, "memories"); got != 0 {
		t.Fatalf("memories rows = %d, want 0", got)
	}
}

func TestCreateMemoryMissingSummary(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	_, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
	if got := countRows(t, db, "memories"); got != 0 {
		t.Fatalf("memories rows = %d, want 0", got)
	}
}

func TestMemoryGetRoundTrip(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	id, err := CreateMemory(db, "gotcha", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX", "cwd can drift after a cd compound command")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}

	view, err := MemoryGet(db, id)
	if err != nil {
		t.Fatalf("MemoryGet: %v", err)
	}
	if view.ID != id || view.Type != "gotcha" || view.Scope != domain.MemoryScopePlan {
		t.Fatalf("view = %+v, want id=%q type=gotcha scope=plan", view, id)
	}
	if view.PlanID == nil || *view.PlanID != "01PLANIDFIXTUREXXXXXXXXXX" {
		t.Fatalf("view.PlanID = %v, want 01PLANIDFIXTUREXXXXXXXXXX", view.PlanID)
	}
	if view.Body != "cwd can drift after a cd compound command" {
		t.Fatalf("view.Body = %q, want the summary body", view.Body)
	}
}

func TestMemoryGetUnknownID(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	_, err := MemoryGet(db, "01UNKNOWNMEMORYIDXXXXXXXX")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_memory_id" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: unknown_memory_id}", err)
	}
}

func TestMemoryQueryFilters(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	globalID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "global gotcha")
	if err != nil {
		t.Fatalf("CreateMemory (global): %v", err)
	}
	planID, err := CreateMemory(db, "gotcha", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX", "plan gotcha")
	if err != nil {
		t.Fatalf("CreateMemory (plan): %v", err)
	}
	if _, err := CreateMemory(db, "decision", domain.MemoryScopeGlobal, "", "unrelated type"); err != nil {
		t.Fatalf("CreateMemory (decision): %v", err)
	}

	all, err := MemoryQuery(db, "gotcha", "", "")
	if err != nil {
		t.Fatalf("MemoryQuery (type only): %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("MemoryQuery type=gotcha len = %d, want 2", len(all))
	}

	globalOnly, err := MemoryQuery(db, "gotcha", domain.MemoryScopeGlobal, "")
	if err != nil {
		t.Fatalf("MemoryQuery (scope filter): %v", err)
	}
	if len(globalOnly) != 1 || globalOnly[0].ID != globalID {
		t.Fatalf("MemoryQuery scope=global = %+v, want just %q", globalOnly, globalID)
	}

	planOnly, err := MemoryQuery(db, "gotcha", "", "01PLANIDFIXTUREXXXXXXXXXX")
	if err != nil {
		t.Fatalf("MemoryQuery (plan-id filter): %v", err)
	}
	if len(planOnly) != 1 || planOnly[0].ID != planID {
		t.Fatalf("MemoryQuery plan-id filter = %+v, want just %q", planOnly, planID)
	}
}

func TestMemoryQueryMissingType(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	_, err := MemoryQuery(db, "", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
}

// TestRebuildMemoriesFromMarkdownRoundTrip proves R4
// (docs/plans/active/durable-memory.md): wiping the memories index and
// rebuilding from committed docs/memory/*.md content alone reconstructs
// every entry to match its pre-wipe state.
func TestRebuildMemoriesFromMarkdownRoundTrip(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	globalID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "global gotcha body")
	if err != nil {
		t.Fatalf("CreateMemory (global): %v", err)
	}
	planID, err := CreateMemory(db, "decision", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX", "plan decision body")
	if err != nil {
		t.Fatalf("CreateMemory (plan): %v", err)
	}

	preGlobal, err := MemoryGet(db, globalID)
	if err != nil {
		t.Fatalf("MemoryGet (pre-wipe global): %v", err)
	}
	prePlan, err := MemoryGet(db, planID)
	if err != nil {
		t.Fatalf("MemoryGet (pre-wipe plan): %v", err)
	}

	if _, err := db.Exec(`DELETE FROM memories`); err != nil {
		t.Fatalf("wipe memories: %v", err)
	}
	if got := countRows(t, db, "memories"); got != 0 {
		t.Fatalf("memories rows after wipe = %d, want 0", got)
	}

	result, err := RebuildFromMarkdown(db)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Memories != 2 {
		t.Fatalf("result.Memories = %d, want 2", result.Memories)
	}

	postGlobal, err := MemoryGet(db, globalID)
	if err != nil {
		t.Fatalf("MemoryGet (post-rebuild global): %v", err)
	}
	if !reflect.DeepEqual(postGlobal, preGlobal) {
		t.Fatalf("post-rebuild global view = %+v, want %+v", postGlobal, preGlobal)
	}

	postPlan, err := MemoryGet(db, planID)
	if err != nil {
		t.Fatalf("MemoryGet (post-rebuild plan): %v", err)
	}
	if !reflect.DeepEqual(postPlan, prePlan) {
		t.Fatalf("post-rebuild plan view = %+v, want %+v", postPlan, prePlan)
	}
}

func TestMemoryQueryRankedOrdersByMatchCount(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	lowID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "sqlite WAL sidecars are tricky")
	if err != nil {
		t.Fatalf("CreateMemory (low): %v", err)
	}
	highID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "sqlite WAL sidecars, sqlite WAL journal mode, sqlite WAL checkpoint")
	if err != nil {
		t.Fatalf("CreateMemory (high): %v", err)
	}
	if _, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "unrelated entry about markdown frontmatter"); err != nil {
		t.Fatalf("CreateMemory (unrelated): %v", err)
	}

	results, err := MemoryQueryRanked(db, "sqlite WAL", "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("MemoryQueryRanked len = %d, want 2 (unrelated entry excluded), got %+v", len(results), results)
	}
	if results[0].ID != highID || results[1].ID != lowID {
		t.Fatalf("MemoryQueryRanked order = [%s, %s], want [%s, %s] (higher match count first)", results[0].ID, results[1].ID, highID, lowID)
	}
	if results[0].Score <= results[1].Score {
		t.Fatalf("MemoryQueryRanked scores = [%d, %d], want first > second", results[0].Score, results[1].Score)
	}
}

func TestMemoryQueryRankedExcludesZeroMatches(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	if _, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "nothing relevant here"); err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}

	results, err := MemoryQueryRanked(db, "unmatchedkeyword", "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("MemoryQueryRanked len = %d, want 0, got %+v", len(results), results)
	}
}

func TestMemoryQueryRankedTiebreaksOnCreatedAtDesc(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	firstID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "keyword match one")
	if err != nil {
		t.Fatalf("CreateMemory (first): %v", err)
	}
	secondID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "keyword match two")
	if err != nil {
		t.Fatalf("CreateMemory (second): %v", err)
	}

	results, err := MemoryQueryRanked(db, "keyword", "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	if len(results) != 2 || results[0].Score != results[1].Score {
		t.Fatalf("MemoryQueryRanked = %+v, want two equal-score entries", results)
	}
	if results[0].ID != secondID || results[1].ID != firstID {
		t.Fatalf("MemoryQueryRanked tiebreak order = [%s, %s], want [%s, %s] (most recent first)", results[0].ID, results[1].ID, secondID, firstID)
	}
}

func TestMemoryQueryRankedCombinesWithFilters(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	planID, err := CreateMemory(db, "gotcha", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX", "shared keyword body")
	if err != nil {
		t.Fatalf("CreateMemory (plan): %v", err)
	}
	if _, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "shared keyword body"); err != nil {
		t.Fatalf("CreateMemory (global): %v", err)
	}
	if _, err := CreateMemory(db, "decision", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX", "shared keyword body"); err != nil {
		t.Fatalf("CreateMemory (decision, plan): %v", err)
	}

	results, err := MemoryQueryRanked(db, "keyword", "gotcha", domain.MemoryScopePlan, "01PLANIDFIXTUREXXXXXXXXXX")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	if len(results) != 1 || results[0].ID != planID {
		t.Fatalf("MemoryQueryRanked filtered = %+v, want just %q", results, planID)
	}
}

func TestMemoryQueryRankedMissingKeywords(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	_, err := MemoryQueryRanked(db, "", "gotcha", "", "")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "missing_required_field" {
		t.Fatalf("err = %v, want *domain.ValidationError{Code: missing_required_field}", err)
	}
}

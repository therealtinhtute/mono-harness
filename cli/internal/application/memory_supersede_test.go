package application

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// TestSupersedeMemoryHappyPathWritesFrontmatterAndIndex proves R1:
// supersede writes superseded_by + superseded_at into the old entry's
// frontmatter and mirrors superseded_by in the index, md-first ordering.
func TestSupersedeMemoryHappyPathWritesFrontmatterAndIndex(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old body to be superseded")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new correction body")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}

	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	// Frontmatter round-trip: old file must contain superseded_by + superseded_at.
	oldPath := "docs/memory/" + oldID + ".md"
	content, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatalf("read old file: %v", err)
	}
	if !strings.Contains(string(content), "superseded_by: "+newID) {
		t.Fatalf("old frontmatter missing superseded_by %q, got %q", newID, string(content))
	}
	if !strings.Contains(string(content), "superseded_at:") {
		t.Fatalf("old frontmatter missing superseded_at, got %q", string(content))
	}

	// Index mirror: memories.superseded_by must match.
	var gotSupersededBy sql.NullString
	if err := db.QueryRow(`SELECT superseded_by FROM memories WHERE id = ?`, oldID).Scan(&gotSupersededBy); err != nil {
		t.Fatalf("query superseded_by: %v", err)
	}
	if !gotSupersededBy.Valid || gotSupersededBy.String != newID {
		t.Fatalf("superseded_by = %v, want %q", gotSupersededBy, newID)
	}
	var gotSupersededAt sql.NullString
	if err := db.QueryRow(`SELECT superseded_at FROM memories WHERE id = ?`, oldID).Scan(&gotSupersededAt); err != nil {
		t.Fatalf("query superseded_at: %v", err)
	}
	if !gotSupersededAt.Valid || gotSupersededAt.String == "" {
		t.Fatalf("superseded_at = %v, want non-empty", gotSupersededAt)
	}

	// New entry must remain active (not superseded itself).
	var newSuperseded sql.NullString
	if err := db.QueryRow(`SELECT superseded_by FROM memories WHERE id = ?`, newID).Scan(&newSuperseded); err != nil {
		t.Fatalf("query new superseded_by: %v", err)
	}
	if newSuperseded.Valid {
		t.Fatalf("new entry superseded_by = %v, want NULL", newSuperseded)
	}
}

// TestSupersedeMemoryUnknownOldID proves R1 refusal for unknown old-id.
func TestSupersedeMemoryUnknownOldID(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new body")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	err = SupersedeMemory(db, "01UNKNOWNOLDIDXXXXXXXXXX", newID)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_memory_id" {
		t.Fatalf("err = %v, want *ValidationError{Code: unknown_memory_id}", err)
	}
}

// TestSupersedeMemoryUnknownNewID proves R1 refusal for unknown new-id.
func TestSupersedeMemoryUnknownNewID(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old body")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	err = SupersedeMemory(db, oldID, "01UNKNOWNNEWIDXXXXXXXXXX")
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "unknown_memory_id" {
		t.Fatalf("err = %v, want *ValidationError{Code: unknown_memory_id}", err)
	}
}

// TestSupersedeMemorySelfSupersede proves R1 refusal for self-supersession.
func TestSupersedeMemorySelfSupersede(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	id, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "body")
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	err = SupersedeMemory(db, id, id)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "supersede_self" {
		t.Fatalf("err = %v, want *ValidationError{Code: supersede_self}", err)
	}
}

// TestSupersedeMemoryAlreadySuperseded proves R1 refusal for re-superseding.
func TestSupersedeMemoryAlreadySuperseded(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old body")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new body")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	newerID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "newer body")
	if err != nil {
		t.Fatalf("CreateMemory newer: %v", err)
	}

	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("first SupersedeMemory: %v", err)
	}
	err = SupersedeMemory(db, oldID, newerID)
	ve, ok := err.(*domain.ValidationError)
	if !ok || ve.Code != "already_superseded" {
		t.Fatalf("err = %v, want *ValidationError{Code: already_superseded}", err)
	}
}

// TestSupersedeQueryDefaultExcludesSuperseded proves R2: default query hides superseded.
func TestSupersedeQueryDefaultExcludesSuperseded(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old superseded body")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new correction body")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	views, err := MemoryQuery(db, "gotcha", "", "")
	if err != nil {
		t.Fatalf("MemoryQuery: %v", err)
	}
	for _, v := range views {
		if v.ID == oldID {
			t.Fatalf("MemoryQuery returned superseded id %q, want excluded by default", oldID)
		}
	}
	foundNew := false
	for _, v := range views {
		if v.ID == newID {
			foundNew = true
		}
	}
	if !foundNew {
		t.Fatalf("MemoryQuery did not return new id %q", newID)
	}
}

// TestSupersedeQueryIncludeSupersededRestores proves R2: --include-superseded restores.
func TestSupersedeQueryIncludeSupersededRestores(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old superseded body")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new correction body")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	views, err := MemoryQuery(db, "gotcha", "", "")
	if err != nil {
		t.Fatalf("MemoryQuery default: %v", err)
	}
	// default must hide old
	for _, v := range views {
		if v.ID == oldID {
			t.Fatalf("default query should hide superseded %q", oldID)
		}
	}
	// include should show it — we call the include variant if it exists,
	// otherwise we test via direct SQL that superseded_by chain is present.
	viewsAll, err := MemoryQueryWithIncludeSuperseded(db, "gotcha", "", "", true)
	if err != nil {
		t.Fatalf("MemoryQueryWithIncludeSuperseded: %v", err)
	}
	foundOld := false
	for _, v := range viewsAll {
		if v.ID == oldID {
			foundOld = true
			if v.SupersededBy == nil || *v.SupersededBy != newID {
				t.Fatalf("include query old.SupersededBy = %v, want %q", v.SupersededBy, newID)
			}
		}
	}
	if !foundOld {
		t.Fatalf("include query did not restore superseded %q", oldID)
	}
}

// TestSupersedeRankedDefaultExcludesSuperseded proves R2 for ranked query.
func TestSupersedeRankedDefaultExcludesSuperseded(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old superseded keyword alpha")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new correction keyword alpha")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	results, err := MemoryQueryRanked(db, "keyword alpha", "", "", "")
	if err != nil {
		t.Fatalf("MemoryQueryRanked: %v", err)
	}
	for _, r := range results {
		if r.ID == oldID {
			t.Fatalf("ranked query returned superseded %q, want excluded", oldID)
		}
	}
	foundNew := false
	for _, r := range results {
		if r.ID == newID {
			foundNew = true
		}
	}
	if !foundNew {
		t.Fatalf("ranked query missing new %q", newID)
	}
}

// TestSupersedeRankedIncludeSupersededRestores proves R2 ranked include.
func TestSupersedeRankedIncludeSupersededRestores(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old superseded keyword beta")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new correction keyword beta")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	resultsAll, err := MemoryQueryRankedWithIncludeSuperseded(db, "keyword beta", "", "", "", true)
	if err != nil {
		t.Fatalf("MemoryQueryRankedWithIncludeSuperseded: %v", err)
	}
	foundOld := false
	for _, r := range resultsAll {
		if r.ID == oldID {
			foundOld = true
		}
	}
	if !foundOld {
		t.Fatalf("ranked include did not restore superseded %q", oldID)
	}
}

// TestSupersedeMemoryGetReportsStatus proves R2: get still resolves and reports status.
func TestSupersedeMemoryGetReportsStatus(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old body for status")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new body for status")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	oldView, err := MemoryGet(db, oldID)
	if err != nil {
		t.Fatalf("MemoryGet old: %v", err)
	}
	if oldView.Status != "superseded" {
		t.Fatalf("oldView.Status = %q, want superseded", oldView.Status)
	}
	if oldView.SupersededBy == nil || *oldView.SupersededBy != newID {
		t.Fatalf("oldView.SupersededBy = %v, want %q", oldView.SupersededBy, newID)
	}
	if oldView.SupersededAt == nil || *oldView.SupersededAt == "" {
		t.Fatalf("oldView.SupersededAt = %v, want non-empty", oldView.SupersededAt)
	}

	newView, err := MemoryGet(db, newID)
	if err != nil {
		t.Fatalf("MemoryGet new: %v", err)
	}
	if newView.Status != "active" {
		t.Fatalf("newView.Status = %q, want active", newView.Status)
	}
	if newView.SupersededBy != nil {
		t.Fatalf("newView.SupersededBy = %v, want nil", newView.SupersededBy)
	}
}

// TestSupersedeContextPacketExcludesSuperseded proves R2 for preflight packet.
func TestSupersedeContextPacketExcludesSuperseded(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old context packet body")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new context packet body")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	pkg, err := BuildContextPacket(db, "work", "dev")
	if err != nil {
		t.Fatalf("BuildContextPacket: %v", err)
	}
	for _, m := range pkg.Memories {
		if m.ID == oldID {
			t.Fatalf("BuildContextPacket included superseded %q, want excluded by default", oldID)
		}
	}
	foundNew := false
	for _, m := range pkg.Memories {
		if m.ID == newID {
			foundNew = true
		}
	}
	if !foundNew {
		t.Fatalf("BuildContextPacket missing new active memory %q, want present", newID)
	}
}

// TestMemoryRebuildSupersedeChain proves R5: db rebuild reproduces chain from markdown.
func TestMemoryRebuildSupersedeChain(t *testing.T) {
	chdirFixture(t)
	db := freshDB(t)

	oldID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "old rebuild body")
	if err != nil {
		t.Fatalf("CreateMemory old: %v", err)
	}
	newID, err := CreateMemory(db, "gotcha", domain.MemoryScopeGlobal, "", "new rebuild body")
	if err != nil {
		t.Fatalf("CreateMemory new: %v", err)
	}
	if err := SupersedeMemory(db, oldID, newID); err != nil {
		t.Fatalf("SupersedeMemory: %v", err)
	}

	// Wipe index and rebuild from markdown alone.
	if _, err := db.Exec(`DELETE FROM memories`); err != nil {
		t.Fatalf("wipe memories: %v", err)
	}
	result, err := RebuildFromMarkdown(db)
	if err != nil {
		t.Fatalf("RebuildFromMarkdown: %v", err)
	}
	if result.Memories != 2 {
		t.Fatalf("result.Memories = %d, want 2", result.Memories)
	}

	var gotSupersededBy sql.NullString
	if err := db.QueryRow(`SELECT superseded_by FROM memories WHERE id = ?`, oldID).Scan(&gotSupersededBy); err != nil {
		t.Fatalf("query after rebuild superseded_by: %v", err)
	}
	if !gotSupersededBy.Valid || gotSupersededBy.String != newID {
		t.Fatalf("after rebuild superseded_by = %v, want %q", gotSupersededBy, newID)
	}
	var gotSupersededAt sql.NullString
	if err := db.QueryRow(`SELECT superseded_at FROM memories WHERE id = ?`, oldID).Scan(&gotSupersededAt); err != nil {
		t.Fatalf("query after rebuild superseded_at: %v", err)
	}
	if !gotSupersededAt.Valid || gotSupersededAt.String == "" {
		t.Fatalf("after rebuild superseded_at = %v, want non-empty", gotSupersededAt)
	}

	// Also verify frontmatter survived wipe — rebuild used it.
	oldView, err := MemoryGet(db, oldID)
	if err != nil {
		t.Fatalf("MemoryGet after rebuild: %v", err)
	}
	if oldView.Status != "superseded" {
		t.Fatalf("after rebuild Status = %q, want superseded", oldView.Status)
	}
}

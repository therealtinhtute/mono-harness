package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// F04: existence and content length are different facts. An empty file that
// existed before the update must come back as an empty file that exists.
func TestStash_EmptyExistingFile_RestoredNotDeleted(t *testing.T) {
	root := tempRepo(t, true)
	pf(t, root, gitignoreTarget, "")

	if err := stashCapture(root, []string{gitignoreTarget}); err != nil {
		t.Fatalf("capture: %v", err)
	}
	pf(t, root, gitignoreTarget, "changed by the update\n")

	if err := stashRestore(root); err != nil {
		t.Fatalf("restore: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, gitignoreTarget))
	if err != nil {
		t.Fatalf("an existing empty file was deleted by restore: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("restored content = %q, want empty", got)
	}
}

// F04 control: a file that did not exist before the update is still removed.
func TestStash_AbsentFile_RemovedOnRestore(t *testing.T) {
	root := tempRepo(t, true)

	if err := stashCapture(root, []string{gitignoreTarget}); err != nil {
		t.Fatalf("capture: %v", err)
	}
	pf(t, root, gitignoreTarget, "written by the update\n")

	if err := stashRestore(root); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, gitignoreTarget)); !os.IsNotExist(err) {
		t.Error("a file absent before the update survived restore")
	}
}

// F04 control: ordinary non-empty content round-trips byte-for-byte.
func TestStash_NonEmptyFile_RoundTrips(t *testing.T) {
	root := tempRepo(t, true)
	want := "original bytes\n"
	pf(t, root, gitignoreTarget, want)

	if err := stashCapture(root, []string{gitignoreTarget}); err != nil {
		t.Fatalf("capture: %v", err)
	}
	pf(t, root, gitignoreTarget, "changed\n")
	if err := stashRestore(root); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got := rf(t, root, gitignoreTarget); got != want {
		t.Errorf("restored %q, want %q", got, want)
	}
}

// F05: an incomplete stash must surface as an actionable failure with the
// recovery data intact, never as a silent success that also destroys the
// remaining backups. Asserted through the exported abort flow.
func TestUpdate_Abort_IncompleteStash_FailsAndKeepsEvidence(t *testing.T) {
	root := tempRepo(t, true)
	pf(t, root, "a.txt", "original a\n")
	pf(t, root, "b.txt", "original b\n")

	if err := stashCapture(root, []string{"a.txt", "b.txt"}); err != nil {
		t.Fatalf("capture: %v", err)
	}
	pf(t, root, "a.txt", "changed a\n")
	pf(t, root, "b.txt", "changed b\n")

	// simulate damaged recovery data: one payload is gone
	entries, _, err := stashLoad(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 stash entries, got %d", len(entries))
	}
	if err := os.Remove(filepath.Join(root, stashDir, entries[0].name)); err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	err = RunUpdate(updateOptions{Root: root, Version: "t", Abort: true}, &sb)
	if err == nil {
		t.Fatalf("abort reported success with an incomplete stash\n%s", sb.String())
	}
	if !strings.Contains(err.Error(), entries[0].rel) {
		t.Errorf("error does not name the unrestorable path: %v", err)
	}
	if strings.Contains(sb.String(), "restored byte-for-byte") {
		t.Errorf("abort announced a restoration that did not happen:\n%s", sb.String())
	}
	if _, serr := os.Stat(filepath.Join(root, stashDir)); serr != nil {
		t.Error("the stash was destroyed after a failed restore; recovery data must be kept")
	}
	// nothing was written: a partial restore is worse than none
	if got := rf(t, root, "b.txt"); got != "changed b\n" {
		t.Errorf("restore wrote %s despite an invalid inventory: %q", "b.txt", got)
	}
}

// R9: a v1 stash cannot distinguish absent from empty, so it declines to
// guess rather than deleting, and reports the ambiguity.
func TestStash_LegacyV1ZeroByteBlob_IsNotDeleted(t *testing.T) {
	root := tempRepo(t, true)
	pf(t, root, "a.txt", "")
	if err := stashCapture(root, []string{"a.txt"}); err != nil {
		t.Fatalf("capture: %v", err)
	}
	// rewrite the inventory in the v1 two-column format
	entries, _, err := stashLoad(root)
	if err != nil {
		t.Fatal(err)
	}
	v1 := entries[0].rel + "\t" + entries[0].name + "\n"
	pf(t, root, filepath.Join(stashDir, "stash.tsv"), v1)
	pf(t, root, "a.txt", "changed\n")

	rerr := stashRestore(root)
	if rerr == nil {
		t.Fatal("a v1 stash resolved an ambiguous entry silently")
	}
	if !strings.Contains(rerr.Error(), "a.txt") {
		t.Errorf("error does not name the ambiguous path: %v", rerr)
	}
	if _, err := os.Stat(filepath.Join(root, "a.txt")); err != nil {
		t.Error("a v1 zero-byte blob deleted the file instead of declining to guess")
	}
}

// The happy path still clears the stash: keeping it is failure handling, not
// the default.
func TestUpdate_Abort_CompleteStash_RestoresAndClears(t *testing.T) {
	root := tempRepo(t, true)
	pf(t, root, "a.txt", "original a\n")
	if err := stashCapture(root, []string{"a.txt"}); err != nil {
		t.Fatalf("capture: %v", err)
	}
	pf(t, root, "a.txt", "changed a\n")

	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t", Abort: true}, &sb); err != nil {
		t.Fatalf("abort: %v\n%s", err, sb.String())
	}
	if got := rf(t, root, "a.txt"); got != "original a\n" {
		t.Errorf("abort left %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, stashDir)); !os.IsNotExist(err) {
		t.Error("a successful abort should clear the stash")
	}
}

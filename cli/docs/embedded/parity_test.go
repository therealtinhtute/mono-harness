package embedded

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// projectionRoot is the checked-out docs/ tree that WORKFLOW.md and
// playbooks/*.md are projected into. See cli/docs/CONTRACT.md:71. AGENTS.md
// is embedded too but is not part of this projection (its counterpart is
// the repo-root AGENTS.md, authored separately), so it is out of scope
// here.
const projectionRoot = "../../../docs"

// reporter is the subset of *testing.T that compareProjection needs, so
// TestProjectionParity_DetectsDrift can capture failures without failing
// the outer test.
type reporter interface {
	Helper()
	Errorf(format string, args ...any)
}

// compareProjection asserts WORKFLOW.md and every playbooks/*.md in
// embedded are byte-identical to their counterpart under root, and that
// root's playbooks/ directory has no extra file embedded lacks. It reports
// via t.Errorf rather than skipping so drift cannot pass silently — see
// docs/patterns/encoding-invariants.md:55-56.
func compareProjection(t reporter, embedded fs.FS, root string) {
	t.Helper()

	compareFile(t, embedded, "WORKFLOW.md", root)

	embeddedPlaybooks, err := fs.ReadDir(embedded, "playbooks")
	if err != nil {
		t.Errorf("read embedded playbooks/: %v", err)
		return
	}
	seen := map[string]bool{}
	for _, entry := range embeddedPlaybooks {
		if entry.IsDir() {
			continue
		}
		rel := filepath.Join("playbooks", entry.Name())
		seen[entry.Name()] = true
		compareFile(t, embedded, rel, root)
	}

	rootPlaybooks, err := os.ReadDir(filepath.Join(root, "playbooks"))
	if err != nil {
		t.Errorf("read projected playbooks/: %v", err)
		return
	}
	for _, entry := range rootPlaybooks {
		if entry.IsDir() {
			continue
		}
		if !seen[entry.Name()] {
			t.Errorf("playbooks/%s: projected file has no embedded counterpart", entry.Name())
		}
	}
}

// compareFile reports a mismatch if the embedded file at rel differs from
// its projected counterpart under root, naming rel either way.
func compareFile(t reporter, embedded fs.FS, rel string, root string) {
	want, err := fs.ReadFile(embedded, rel)
	if err != nil {
		t.Errorf("read embedded %s: %v", rel, err)
		return
	}

	got, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Errorf("%s: missing projection at %s: %v", rel, filepath.Join(root, rel), err)
		return
	}

	if !bytes.Equal(want, got) {
		t.Errorf("%s: embedded and projected copies differ", rel)
	}
}

// TestProjectionParity fails on any one-byte drift between the embedded
// WORKFLOW.md/playbooks and their projection under docs/, in either
// direction. Reading a missing projectionRoot is a hard failure, not a
// skip — a skip lets the gate pass silently, which is how invariant S2 was
// lost.
func TestProjectionParity(t *testing.T) {
	if _, err := os.Stat(projectionRoot); err != nil {
		t.Fatalf("projection root %s unreadable: %v", projectionRoot, err)
	}
	compareProjection(t, FS, projectionRoot)
}

// TestProjectionParity_DetectsDrift proves the comparator used above
// actually reports a mismatch instead of passing vacuously on a green run
// with nothing to catch.
func TestProjectionParity_DetectsDrift(t *testing.T) {
	embedded := fstest.MapFS{
		"WORKFLOW.md":         &fstest.MapFile{Data: []byte("same\n")},
		"playbooks/watzup.md": &fstest.MapFile{Data: []byte("drifted\n")},
	}

	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "WORKFLOW.md"), []byte("same\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "playbooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "playbooks", "watzup.md"), []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := &recordingReporter{}
	compareProjection(rec, embedded, tmp)

	if len(rec.errors) == 0 {
		t.Fatal("expected compareProjection to report the drifted file, got no failure")
	}
	found := false
	for _, msg := range rec.errors {
		if bytes.Contains([]byte(msg), []byte("playbooks/watzup.md")) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a failure naming playbooks/watzup.md, got: %v", rec.errors)
	}
}

// recordingReporter captures compareProjection's Errorf calls instead of
// failing the test that owns it.
type recordingReporter struct {
	errors []string
}

func (r *recordingReporter) Helper() {}

func (r *recordingReporter) Errorf(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

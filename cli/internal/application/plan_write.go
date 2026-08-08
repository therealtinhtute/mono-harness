package application

import (
	"fmt"
	"os"

	"github.com/oklog/ulid/v2"
)

// activePlanForWrite resolves the single active plan a durable write
// should append to, reusing findActivePlans (next.go, V3's parse
// precedent). Zero or more than one active plan is not an error: a
// bounded/simple write legitimately has no active plan, and an ambiguous
// multi-plan repository is a drift condition this feature does not try to
// resolve — both cases proceed DB-only, exactly as before this feature
// existed.
func activePlanForWrite() (path string, ok bool, err error) {
	plans, err := findActivePlans()
	if err != nil {
		return "", false, err
	}
	if len(plans) != 1 {
		return "", false, nil
	}
	return plans[0].path, true, nil
}

// preparePlanAppend resolves the active plan (if exactly one exists) and
// validates that entry can be appended to section, without writing
// anything yet. Callers run this BEFORE their changeset/DB write: a
// malformed or missing section then fails with zero side effects, which
// is how P3 wave 2's "index and markdown cannot diverge" requirement is
// actually achieved — a SQL transaction and a text-file write have no
// shared atomic commit, so the honest guarantee is that the common
// failure mode (a plan the CLI cannot safely write to) never reaches the
// DB write at all. The returned write func performs the real file write
// and must only be called after the DB write has succeeded; when no
// single active plan is resolvable, write is a no-op returning nil.
func preparePlanAppend(section, entry string) (write func() error, err error) {
	path, ok, err := activePlanForWrite()
	if err != nil {
		return nil, err
	}
	if !ok {
		return func() error { return nil }, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read active plan %s: %w", path, err)
	}
	newContent, err := AppendToPlanSection(string(data), section, entry)
	if err != nil {
		return nil, fmt.Errorf("append to active plan %s: %w", path, err)
	}
	return func() error { return writeFileAtomically(path, []byte(newContent)) }, nil
}

// writeFileAtomically writes data to a sibling temp file and renames it
// over path. Rename is atomic on the same filesystem, so a write failure
// (disk full, permissions) never leaves path partially written — the
// worst case is the old content survives untouched, not corrupted.
func writeFileAtomically(path string, data []byte) error {
	tmp := path + ".tmp-" + ulid.Make().String()
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) //nolint:errcheck // best-effort cleanup; the rename error is what matters
		return err
	}
	return nil
}

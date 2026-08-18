package application

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/oklog/ulid/v2"
)

// activePlanForWrite resolves the single active plan a durable write
// should append to, reusing ResolveActivePlan (R2,
// docs/audit/consumer-adoption-audit.md D1) so this caller never
// reconstructs its own ok=false collapse of "zero" and "ambiguous" into one
// indistinguishable state. Zero or more than one active plan is not a
// hard error here: a bounded/simple write legitimately has no active plan,
// and an ambiguous multi-plan repository is a drift condition this feature
// does not try to resolve — both cases proceed DB-only, exactly as before
// this feature existed, but the caller now receives the Stop naming which
// case occurred instead of a bare boolean.
func activePlanForWrite() (path string, stop *StopInfo, err error) {
	plan, stop, err := ResolveActivePlan()
	if err != nil {
		return "", nil, err
	}
	if stop != nil {
		return "", stop, nil
	}
	return plan.path, nil, nil
}

// preparePlanAppend resolves the active plan (if exactly one exists) and
// validates that entry can be appended to section, without writing
// anything yet. Every durable-write caller (trace, decision, check
// record, handoff) calls the returned write func BEFORE its DB
// write, not after: markdown is the write target, and the DB row is
// derived from what was actually written to it (R8,
// docs/plans/active/harness-markdown-truth.md). A SQL transaction and a
// text-file write still have no shared atomic commit, but the ordering
// guarantee is real in the direction that matters — a failed markdown
// write (missing section, read-only file, race) always leaves zero DB
// rows behind it; a DB write failure after a successful markdown write
// leaves the markdown line as the durable fact, with no row to derive
// from it yet. When no single active plan is resolvable, write is a
// no-op returning nil. The returned write func also refreshes plan_index
// against the content it just wrote (wave 3, R9) — these are the only
// callers that both read and durably rewrite the plan, so refreshing here
// is the one place the index can track every real change with no
// speculative caller wiring elsewhere.
func preparePlanAppend(db *sql.DB, section, entry string) (write func() error, err error) {
	path, stop, err := activePlanForWrite()
	if err != nil {
		return nil, err
	}
	if stop != nil {
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
	return func() error {
		if err := writeFileAtomically(path, []byte(newContent)); err != nil {
			return err
		}
		if _, err := refreshPlanIndex(db, path, newContent); err != nil {
			return fmt.Errorf("refresh plan_index for %s: %w", path, err)
		}
		return nil
	}, nil
}

// preparePlanPhaseStatus mirrors preparePlanAppend's read-now/write-later
// split, but for a phase block's `status:` line instead of a section
// append — the write closure story create/run create/check record/handoff
// call before their DB write, so stories stay markdown-first the
// same way trace/decision/check/handoff did in P2 (R8, P3 wave 1,
// docs/plans/active/harness-markdown-truth.md). found is false when slug
// has no phase block in `## Phases and Verification`, or the block has no
// status line — expected for a story created ad hoc via `story create`
// with no matching phase entry, not an error condition — in which case the
// write closure is a no-op, same as preparePlanAppend's no-active-plan case.
func preparePlanPhaseStatus(db *sql.DB, slug, newStatus string) (write func() error, err error) {
	path, stop, err := activePlanForWrite()
	if err != nil {
		return nil, err
	}
	if stop != nil {
		return func() error { return nil }, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read active plan %s: %w", path, err)
	}
	newContent, found := SetPlanPhaseStatus(string(data), slug, newStatus)
	if !found {
		return func() error { return nil }, nil
	}
	return func() error {
		if err := writeFileAtomically(path, []byte(newContent)); err != nil {
			return err
		}
		if _, err := refreshPlanIndex(db, path, newContent); err != nil {
			return fmt.Errorf("refresh plan_index for %s: %w", path, err)
		}
		return nil
	}, nil
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

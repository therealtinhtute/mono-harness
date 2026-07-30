package infrastructure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryChangesetsReplayFromEmpty(t *testing.T) {
	db := freshDB(t)
	changesets := filepath.Join("..", "..", "..", ".kit", "changesets")
	// `.kit/` is gitignored per-machine state, so a fresh clone has no
	// changesets to replay. Replay-from-empty itself is covered by
	// TestChangesetReplay; this test only guards local changesets when present.
	if _, err := os.Stat(changesets); os.IsNotExist(err) {
		t.Skipf("no local %s; nothing to replay", changesets)
	}
	applied, err := Replay(db, changesets)
	if err != nil {
		t.Fatalf("Replay(%s) after schema migration: %v", changesets, err)
	}
	if applied == 0 {
		t.Fatalf("Replay(%s) applied zero lines", changesets)
	}
}

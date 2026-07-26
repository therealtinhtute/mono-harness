package infrastructure

import (
	"path/filepath"
	"testing"
)

func TestRepositoryChangesetsReplayFromEmpty(t *testing.T) {
	db := freshDB(t)
	changesets := filepath.Join("..", "..", "..", ".kit", "changesets")
	applied, err := Replay(db, changesets)
	if err != nil {
		t.Fatalf("Replay(%s) after schema migration: %v", changesets, err)
	}
	if applied == 0 {
		t.Fatalf("Replay(%s) applied zero lines", changesets)
	}
}

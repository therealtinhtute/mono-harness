package application

import (
	"testing"

	"github.com/oklog/ulid/v2"
)

func TestQueryArtifactsPreservesLegacyArtifactPaths(t *testing.T) {
	db, changesetDir := freshDB(t)
	firstID := seedRun(t, db, changesetDir)
	if _, err := db.Exec(`UPDATE runs SET artifact_path = '' WHERE id = ?`, firstID); err != nil {
		t.Fatalf("clear artifact_path: %v", err)
	}

	missingPath := ".kit/runs/work/nonexistent.md"
	secondID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO runs (id, story_slug, trace_ids, artifact_path, created_at)
		VALUES (?, 'cli-domain', '[]', ?, '2026-07-27T12:01:00Z')
	`, secondID, missingPath); err != nil {
		t.Fatalf("insert second run: %v", err)
	}

	artifacts, err := QueryArtifacts(db, "cli-domain")
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("artifacts = %+v, want two typed run rows", artifacts)
	}
	if artifacts[0].ID != firstID || artifacts[0].ArtifactPath != "" {
		t.Fatalf("first artifact = %+v, want empty artifact_path preserved", artifacts[0])
	}
	if artifacts[1].ID != secondID || artifacts[1].ArtifactPath != missingPath {
		t.Fatalf("second artifact = %+v, want nonexistent artifact_path preserved", artifacts[1])
	}
}

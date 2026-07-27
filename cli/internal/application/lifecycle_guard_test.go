package application

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

type lifecycleSnapshot struct {
	stories            int
	runs               int
	checks             int
	handoffs           int
	storyStatus        string
	schemaVersion      int
	currentPhase       sql.NullString
	entryPhase         sql.NullString
	latestRunID        sql.NullString
	latestCheckID      sql.NullString
	lastAppliedChanges sql.NullString
	changesets         int
}

func takeLifecycleSnapshot(t *testing.T, db *sql.DB, changesetDir, storySlug string) lifecycleSnapshot {
	t.Helper()

	snapshot := lifecycleSnapshot{
		stories:     countRows(t, db, "stories"),
		runs:        countRows(t, db, "runs"),
		checks:      countRows(t, db, "checks"),
		handoffs:    countRows(t, db, "handoffs"),
		storyStatus: queryStoryStatus(t, db, storySlug),
		changesets:  countLifecycleChangesets(t, changesetDir),
	}
	if err := db.QueryRow(`
		SELECT schema_version, current_phase, entry_phase, latest_run_id, latest_check_id, last_applied_changeset
		FROM meta
		LIMIT 1
	`).Scan(
		&snapshot.schemaVersion,
		&snapshot.currentPhase,
		&snapshot.entryPhase,
		&snapshot.latestRunID,
		&snapshot.latestCheckID,
		&snapshot.lastAppliedChanges,
	); err != nil {
		t.Fatalf("query lifecycle meta: %v", err)
	}
	return snapshot
}

func countLifecycleChangesets(t *testing.T, changesetDir string) int {
	t.Helper()
	entries, err := os.ReadDir(changesetDir)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("read changeset dir: %v", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".changeset.jsonl") {
			count++
		}
	}
	return count
}

func assertLifecycleUnchanged(t *testing.T, before, after lifecycleSnapshot) {
	t.Helper()
	if before != after {
		t.Fatalf("lifecycle changed after rejected transition:\nbefore = %+v\nafter  = %+v", before, after)
	}
}

func assertLifecycleValidationError(t *testing.T, err error, code, message string) {
	t.Helper()
	ve, ok := err.(*domain.ValidationError)
	if !ok {
		t.Fatalf("error = %v, want *domain.ValidationError", err)
	}
	if ve.Code != code || ve.Message != message {
		t.Fatalf("validation error = %#v, want code=%q message=%q", ve, code, message)
	}
}

func createLifecycleRun(t *testing.T, db *sql.DB, changesetDir, storySlug string) string {
	t.Helper()
	if _, _, err := CreateStory(db, changesetDir, storySlug, "exercise lifecycle guards", ""); err != nil {
		t.Fatalf("CreateStory: %v", err)
	}
	runID, _, err := CreateRun(db, changesetDir, storySlug, "", "")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	return runID
}

func recordCleanLifecycleCheck(t *testing.T, db *sql.DB, changesetDir, runID string) string {
	t.Helper()
	checkID, _, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, []domain.ProofLink{{Command: "go test ./...", OutputRef: "ok"}})
	if err != nil {
		t.Fatalf("RecordCheck: %v", err)
	}
	return checkID
}

func TestLifecycleGuardRunCreateRejectsCheckedAndDone(t *testing.T) {
	for _, status := range []string{domain.StoryChecked, domain.StoryDone} {
		t.Run(status, func(t *testing.T) {
			db, changesetDir := freshDB(t)
			storySlug := "terminal-run-create"
			runID := createLifecycleRun(t, db, changesetDir, storySlug)
			checkID := recordCleanLifecycleCheck(t, db, changesetDir, runID)
			if status == domain.StoryDone {
				if _, _, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true); err != nil {
					t.Fatalf("RecordHandoff: %v", err)
				}
			}

			before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
			id, path, err := CreateRun(db, changesetDir, storySlug, "", "")
			assertLifecycleValidationError(t, err, "story_not_runnable", "run create: story must be planned or in-progress")
			if id != "" || path != "" {
				t.Fatalf("rejected CreateRun returned id=%q path=%q, want empty values", id, path)
			}
			after := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
			assertLifecycleUnchanged(t, before, after)
		})
	}
}

func TestLifecycleGuardCheckUsesLatestInProgressRun(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "latest-run-check"
	olderRunID := createLifecycleRun(t, db, changesetDir, storySlug)
	latestRunID, _, err := CreateRun(db, changesetDir, storySlug, "", "")
	if err != nil {
		t.Fatalf("second CreateRun: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
	id, path, err := RecordCheck(db, changesetDir, olderRunID, domain.VerdictApproved, []domain.ProofLink{{Command: "go test ./...", OutputRef: "ok"}})
	assertLifecycleValidationError(t, err, "run_not_latest", "check record: run_id is not the latest run for its story")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordCheck returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))

	if _, _, err := RecordCheck(db, changesetDir, latestRunID, domain.VerdictApproved, []domain.ProofLink{{Command: "go test ./...", OutputRef: "ok"}}); err != nil {
		t.Fatalf("RecordCheck(latest run): %v", err)
	}
	if got := queryStoryStatus(t, db, storySlug); got != domain.StoryChecked {
		t.Fatalf("story status = %q, want checked", got)
	}
}

func TestLifecycleGuardCheckRejectsCheckedAndDoneStory(t *testing.T) {
	for _, status := range []string{domain.StoryChecked, domain.StoryDone} {
		t.Run(status, func(t *testing.T) {
			db, changesetDir := freshDB(t)
			storySlug := "terminal-check"
			runID := createLifecycleRun(t, db, changesetDir, storySlug)
			checkID := recordCleanLifecycleCheck(t, db, changesetDir, runID)
			if status == domain.StoryDone {
				if _, _, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true); err != nil {
					t.Fatalf("RecordHandoff: %v", err)
				}
			}

			before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
			id, path, err := RecordCheck(db, changesetDir, runID, domain.VerdictApproved, []domain.ProofLink{{Command: "go test ./...", OutputRef: "ok"}})
			assertLifecycleValidationError(t, err, "story_not_checkable", "check record: story must be in-progress")
			if id != "" || path != "" {
				t.Fatalf("rejected RecordCheck returned id=%q path=%q, want empty values", id, path)
			}
			assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
		})
	}
}

func TestLifecycleGuardClosingHandoffRejectsNonLatestRun(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "stale-run-close"
	olderRunID := createLifecycleRun(t, db, changesetDir, storySlug)
	checkID := recordCleanLifecycleCheck(t, db, changesetDir, olderRunID)
	if _, err := db.Exec(`
		INSERT INTO runs (id, story_slug, artifact_path, created_at)
		VALUES (?, ?, '', '9999-01-01T00:00:00Z')
	`, ulid.Make().String(), storySlug); err != nil {
		t.Fatalf("insert newer legacy run: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
	id, path, err := RecordHandoff(db, changesetDir, olderRunID, checkID, nil, true)
	assertLifecycleValidationError(t, err, "run_not_latest", "handoff record: run_id is not the latest run for its story")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
}

func TestLifecycleGuardClosingHandoffRejectsNonLatestCheck(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "stale-check-close"
	runID := createLifecycleRun(t, db, changesetDir, storySlug)
	olderCheckID := recordCleanLifecycleCheck(t, db, changesetDir, runID)
	if _, err := db.Exec(`
		INSERT INTO checks (id, run_id, verdict, created_at)
		VALUES (?, ?, ?, '9999-01-01T00:00:00Z')
	`, ulid.Make().String(), runID, domain.VerdictApproved); err != nil {
		t.Fatalf("insert newer legacy check: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
	id, path, err := RecordHandoff(db, changesetDir, runID, olderCheckID, nil, true)
	assertLifecycleValidationError(t, err, "check_not_latest", "handoff record: check_id is not the latest check for its run")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
}

func TestLifecycleGuardClosingHandoffRequiresCheckedStory(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "unchecked-close"
	runID := createLifecycleRun(t, db, changesetDir, storySlug)
	checkID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO checks (id, run_id, verdict, created_at)
		VALUES (?, ?, ?, '9999-01-01T00:00:00Z')
	`, checkID, runID, domain.VerdictApproved); err != nil {
		t.Fatalf("insert legacy check without status transition: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true)
	assertLifecycleValidationError(t, err, "phase_not_checked", "handoff record: story must be checked before phase close")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
}

func TestLifecycleGuardClosingHandoffRejectsUnknownVerdict(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "unknown-verdict-close"
	runID := createLifecycleRun(t, db, changesetDir, storySlug)
	checkID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO checks (id, run_id, verdict, created_at)
		VALUES (?, ?, 'MAYBE', '9999-01-01T00:00:00Z')
	`, checkID, runID); err != nil {
		t.Fatalf("insert legacy check with unknown verdict: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
	id, path, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true)
	assertLifecycleValidationError(t, err, "check_not_clean", "handoff record: cannot close a phase with REQUEST_CHANGES")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
}

func TestLifecycleGuardClosingHandoffRejectsCheckRunMismatch(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "mismatched-check-close"
	checkedRunID := createLifecycleRun(t, db, changesetDir, storySlug)
	checkID := recordCleanLifecycleCheck(t, db, changesetDir, checkedRunID)
	otherRunID := ulid.Make().String()
	if _, err := db.Exec(`
		INSERT INTO runs (id, story_slug, artifact_path, created_at)
		VALUES (?, ?, '', '9999-01-01T00:00:00Z')
	`, otherRunID, storySlug); err != nil {
		t.Fatalf("insert mismatched legacy run: %v", err)
	}

	before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
	id, path, err := RecordHandoff(db, changesetDir, otherRunID, checkID, nil, true)
	assertLifecycleValidationError(t, err, "check_run_mismatch", "handoff record: check does not gate the supplied run")
	if id != "" || path != "" {
		t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
	}
	assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
}

func TestLifecycleGuardClosingHandoffRequiresBothAnchors(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "missing-close-anchors"
	runID := createLifecycleRun(t, db, changesetDir, storySlug)
	checkID := recordCleanLifecycleCheck(t, db, changesetDir, runID)

	for _, tc := range []struct {
		name    string
		runID   string
		checkID string
	}{
		{name: "both missing"},
		{name: "run missing", checkID: checkID},
		{name: "check missing", runID: runID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := takeLifecycleSnapshot(t, db, changesetDir, storySlug)
			id, path, err := RecordHandoff(db, changesetDir, tc.runID, tc.checkID, nil, true)
			assertLifecycleValidationError(t, err, "missing_required_field", "handoff record: --close-phase requires run_id and check_id")
			if id != "" || path != "" {
				t.Fatalf("rejected RecordHandoff returned id=%q path=%q, want empty values", id, path)
			}
			assertLifecycleUnchanged(t, before, takeLifecycleSnapshot(t, db, changesetDir, storySlug))
		})
	}
}

func TestLifecycleGuardLatestCleanCloseReachesDone(t *testing.T) {
	db, changesetDir := freshDB(t)
	storySlug := "clean-close"
	runID := createLifecycleRun(t, db, changesetDir, storySlug)
	checkID := recordCleanLifecycleCheck(t, db, changesetDir, runID)

	if _, _, err := RecordHandoff(db, changesetDir, runID, checkID, nil, true); err != nil {
		t.Fatalf("RecordHandoff: %v", err)
	}
	if got := queryStoryStatus(t, db, storySlug); got != domain.StoryDone {
		t.Fatalf("story status = %q, want done", got)
	}
}

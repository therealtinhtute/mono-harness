package interfaces

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

func TestExclusiveMutationCommandInventory(t *testing.T) {
	root := NewRootCmd("dev")
	var got []string
	var visit func(*cobra.Command, string)
	visit = func(parent *cobra.Command, prefix string) {
		for _, command := range parent.Commands() {
			path := command.Name()
			if prefix != "" {
				path = prefix + " " + path
			}
			if command.Annotations[repositoryLockAnnotation] == "exclusive" {
				got = append(got, path)
			}
			visit(command, path)
		}
	}
	visit(root, "")
	sort.Strings(got)

	want := make([]string, 0, len(exclusiveMutationCommandPaths))
	for path := range exclusiveMutationCommandPaths {
		want = append(want, path)
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("exclusive command paths = %v, want %v", got, want)
	}
	for _, path := range []string{"id", "scaffold", "preflight", "query", "db changeset status", "next", "resume", "validate", "audit"} {
		if containsString(got, path) {
			t.Fatalf("pure reader/file-only command %q is exclusively wrapped", path)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestLifecycleValidationRunsInsideExclusiveMutationBoundary(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := executeRepositoryLockCommand("init", "--json"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := executeRepositoryLockCommand("story", "--slug", "serialized", "--goal", "serialize check and run", "--json"); err != nil {
		t.Fatalf("story: %v", err)
	}
	runOutput, err := executeRepositoryLockCommand("run", "create", "--slug", "serialized", "--json")
	if err != nil {
		t.Fatalf("initial run: %v", err)
	}
	var runResponse struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(runOutput, &runResponse); err != nil {
		t.Fatalf("decode initial run: %v", err)
	}
	before, err := infrastructure.ListChangesets(changesetDir)
	if err != nil {
		t.Fatalf("ListChangesets before race: %v", err)
	}

	checkAcquired := make(chan struct{})
	releaseCheck := make(chan struct{})
	repositoryLockAcquiredHook = func(path string) {
		if path == "check record" {
			close(checkAcquired)
			<-releaseCheck
		}
	}
	defer func() { repositoryLockAcquiredHook = nil }()

	jsonOutput = false
	resumeFacts = ""
	checkCmd, _ := newRepositoryLockCommand("check", "record", "--verdict", "APPROVED", "--run-id", runResponse.ID, "--judge", "independent", "--judge-model", "test-model", "--proof-links", `[{"command":"true","output_ref":"pass"}]`)
	runCmd, _ := newRepositoryLockCommand("run", "create", "--slug", "serialized")
	checkDone := make(chan error, 1)
	go func() { checkDone <- checkCmd.Execute() }()
	select {
	case <-checkAcquired:
	case <-time.After(2 * time.Second):
		t.Fatal("check command did not acquire the mutation boundary")
	}

	runDone := make(chan error, 1)
	go func() { runDone <- runCmd.Execute() }()
	select {
	case err := <-runDone:
		close(releaseCheck)
		t.Fatalf("second run entered while check held the mutation boundary: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseCheck)
	if err := <-checkDone; err != nil {
		t.Fatalf("check command: %v", err)
	}
	if err := <-runDone; err == nil {
		t.Fatal("second run succeeded after clean check, want story_not_runnable")
	} else if cliErr, ok := err.(*cliError); !ok || cliErr.Code != "story_not_runnable" {
		t.Fatalf("second run error = %T %v, want story_not_runnable", err, err)
	}

	after, err := infrastructure.ListChangesets(changesetDir)
	if err != nil {
		t.Fatalf("ListChangesets after race: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Fatalf("changesets after race = %d, want exactly check append above %d", len(after), len(before))
	}
	readOnly, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	defer readOnly.Close()
	var status string
	var runs, checks int
	if err := readOnly.QueryRow(`SELECT status FROM stories WHERE slug = 'serialized'`).Scan(&status); err != nil {
		t.Fatalf("query story status: %v", err)
	}
	if err := readOnly.QueryRow(`SELECT COUNT(*) FROM runs WHERE story_slug = 'serialized'`).Scan(&runs); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if err := readOnly.QueryRow(`SELECT COUNT(*) FROM checks`).Scan(&checks); err != nil {
		t.Fatalf("count checks: %v", err)
	}
	if status != "checked" || runs != 1 || checks != 1 {
		t.Fatalf("race state = status=%q runs=%d checks=%d, want checked/1/1", status, runs, checks)
	}
}

func TestSerializedPublicRunsAndChecksMatchULIDReplayOrder(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := executeRepositoryLockCommand("init", "--json"); err != nil {
		t.Fatalf("init: %v", err)
	}
	storyID := executeIDCommand(t, "story", "--slug", "replay-order", "--goal", "preserve latest order", "--json")
	futureTime := time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC)
	futureFloor, err := ulid.New(ulid.Timestamp(futureTime), bytes.NewReader(make([]byte, 10)))
	if err != nil {
		t.Fatalf("create future changeset floor: %v", err)
	}
	floorPath, err := infrastructure.WriteChangesetWithID(changesetDir, futureFloor.String(), []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "preserve latest order"},
		At:     futureTime.Format(time.RFC3339),
	}})
	if err != nil {
		t.Fatalf("write future changeset floor: %v", err)
	}
	if output, err := executeRepositoryLockCommand("db", "changeset", "apply", floorPath, "--json"); err != nil {
		t.Fatalf("apply future changeset floor: %v (output=%s)", err, output)
	}

	runA := executeIDCommand(t, "run", "create", "--slug", "replay-order", "--json")
	runB := executeIDCommand(t, "run", "create", "--slug", "replay-order", "--json")
	proof := `[{"command":"go test ./...","output_ref":"same-time proof"}]`
	checkA := executeIDCommand(t, "check", "record", "--verdict", "REQUEST_CHANGES", "--run-id", runB, "--judge", "independent", "--judge-model", "test-model", "--proof-links", proof, "--json")
	checkB := executeIDCommand(t, "check", "record", "--verdict", "REQUEST_CHANGES", "--run-id", runB, "--judge", "independent", "--judge-model", "test-model", "--proof-links", proof, "--json")
	if runB <= runA || checkB <= checkA {
		t.Fatalf("ordered entity IDs = runs(%s,%s) checks(%s,%s), want command order", runA, runB, checkA, checkB)
	}
	for _, id := range []string{runA, runB, checkA, checkB} {
		if _, err := infrastructure.ReadChangeset(filepath.Join(changesetDir, id+".changeset.jsonl")); err != nil {
			t.Fatalf("entity %s does not share its ordered changeset id: %v", id, err)
		}
	}

	liveDB, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly live: %v", err)
	}
	var sharedTimestamp string
	for _, entity := range []struct {
		table string
		id    string
	}{
		{table: "runs", id: runA},
		{table: "runs", id: runB},
		{table: "checks", id: checkA},
		{table: "checks", id: checkB},
	} {
		var createdAt string
		if err := liveDB.QueryRow("SELECT created_at FROM "+entity.table+" WHERE id = ?", entity.id).Scan(&createdAt); err != nil {
			t.Fatalf("query %s timestamp for %s: %v", entity.table, entity.id, err)
		}
		expected := time.UnixMilli(int64(ulid.MustParse(entity.id).Time())).UTC().Format("2006-01-02T15:04:05.000Z07:00")
		if createdAt != expected {
			t.Fatalf("%s timestamp for %s = %q, want ordered ID time %q", entity.table, entity.id, createdAt, expected)
		}
		if sharedTimestamp == "" {
			sharedTimestamp = createdAt
		} else if createdAt != sharedTimestamp {
			t.Fatalf("serialized entities do not share the forced millisecond: got %q and %q", sharedTimestamp, createdAt)
		}
	}
	live := latestOrderSnapshot(t, liveDB.Raw(), "replay-order")
	if err := liveDB.Close(); err != nil {
		t.Fatalf("close live db: %v", err)
	}
	if live.latestRun != runB || live.metaRun != runB || live.latestCheck != checkB || live.metaCheck != checkB {
		t.Fatalf("live latest state = %+v, want run=%s check=%s", live, runB, checkB)
	}

	replayPath := filepath.Join(t.TempDir(), "replay.db")
	replayDB, err := infrastructure.Open(replayPath)
	if err != nil {
		t.Fatalf("Open replay db: %v", err)
	}
	defer replayDB.Close()
	if _, _, err := infrastructure.Migrate(replayDB); err != nil {
		t.Fatalf("Migrate replay db: %v", err)
	}
	if _, err := infrastructure.Replay(replayDB, changesetDir); err != nil {
		t.Fatalf("Replay: %v", err)
	}
	replayed := latestOrderSnapshot(t, replayDB, "replay-order")
	if replayed != live {
		t.Fatalf("latest state differs after ULID replay: live=%+v replay=%+v", live, replayed)
	}
}

type latestOrderTestSnapshot struct {
	latestRun   string
	metaRun     string
	latestCheck string
	metaCheck   string
	runCount    int
	checkCount  int
}

func latestOrderSnapshot(t *testing.T, db interface {
	QueryRow(query string, args ...any) *sql.Row
}, slug string) latestOrderTestSnapshot {
	t.Helper()
	var snapshot latestOrderTestSnapshot
	if err := db.QueryRow(`SELECT id FROM runs WHERE story_slug = ? ORDER BY created_at DESC, id DESC LIMIT 1`, slug).Scan(&snapshot.latestRun); err != nil {
		t.Fatalf("query latest run: %v", err)
	}
	if err := db.QueryRow(`SELECT id FROM checks ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&snapshot.latestCheck); err != nil {
		t.Fatalf("query latest check: %v", err)
	}
	if err := db.QueryRow(`SELECT latest_run_id, latest_check_id FROM meta`).Scan(&snapshot.metaRun, &snapshot.metaCheck); err != nil {
		t.Fatalf("query meta pointers: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM runs WHERE story_slug = ?`, slug).Scan(&snapshot.runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM checks`).Scan(&snapshot.checkCount); err != nil {
		t.Fatalf("count checks: %v", err)
	}
	return snapshot
}

func executeIDCommand(t *testing.T, args ...string) string {
	t.Helper()
	output, err := executeRepositoryLockCommand(args...)
	if err != nil {
		t.Fatalf("command %v: %v (output=%s)", args, err, output)
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(output, &response); err != nil || response.ID == "" {
		t.Fatalf("decode command %v output %q: id=%q err=%v", args, output, response.ID, err)
	}
	return response.ID
}

func executeRepositoryLockCommand(args ...string) ([]byte, error) {
	jsonOutput = false
	resumeFacts = ""
	cmd, output := newRepositoryLockCommand(args...)
	err := cmd.Execute()
	return output.Bytes(), err
}

func newRepositoryLockCommand(args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewRootCmd("dev")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)
	return cmd, &output
}

func TestPublicChangesetRecoveryRequiresEarliestPending(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := executeRepositoryLockCommand("init", "--json"); err != nil {
		t.Fatalf("init: %v", err)
	}
	storyID := executeIDCommand(t, "story", "--slug", "public-recovery", "--goal", "before recovery", "--json")
	first, err := infrastructure.WriteChangeset(changesetDir, []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "first"},
		At:     "2026-07-27T00:01:00Z",
	}})
	if err != nil {
		t.Fatalf("write first pending changeset: %v", err)
	}
	second, err := infrastructure.WriteChangeset(changesetDir, []infrastructure.ChangesetLine{{
		Op: "update", Entity: "story", ID: storyID,
		Fields: map[string]any{"goal": "second"},
		At:     "2026-07-27T00:02:00Z",
	}})
	if err != nil {
		t.Fatalf("write second pending changeset: %v", err)
	}
	before, err := infrastructure.ListChangesets(changesetDir)
	if err != nil {
		t.Fatalf("list changesets before recovery: %v", err)
	}

	if _, err := executeRepositoryLockCommand("intake", "--type", "maintenance", "--summary", "must recover first", "--lane", "tiny", "--json"); err == nil {
		t.Fatal("ordinary mutation succeeded with pending changesets")
	} else if cliErr, ok := err.(*cliError); !ok || cliErr.Code != "changeset_recovery_required" {
		t.Fatalf("ordinary mutation error = %T %v, want changeset_recovery_required", err, err)
	}
	if after, err := infrastructure.ListChangesets(changesetDir); err != nil {
		t.Fatalf("list changesets after blocked mutation: %v", err)
	} else if !reflect.DeepEqual(after, before) {
		t.Fatalf("blocked mutation changed changeset files: before=%v after=%v", before, after)
	}

	if _, err := executeRepositoryLockCommand("db", "changeset", "apply", second, "--json"); err == nil {
		t.Fatal("later pending changeset applied before the earliest")
	} else if cliErr, ok := err.(*cliError); !ok || cliErr.Code != "changeset_recovery_required" {
		t.Fatalf("later apply error = %T %v, want changeset_recovery_required", err, err)
	}
	if output, err := executeRepositoryLockCommand("db", "changeset", "apply", first, "--json"); err != nil {
		t.Fatalf("apply first pending changeset: %v (output=%s)", err, output)
	}
	if output, err := executeRepositoryLockCommand("db", "changeset", "apply", second, "--json"); err != nil {
		t.Fatalf("apply second pending changeset: %v (output=%s)", err, output)
	}

	readOnly, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly after recovery: %v", err)
	}
	defer readOnly.Close()
	var goal string
	if err := readOnly.QueryRow(`SELECT goal FROM stories WHERE id = ?`, storyID).Scan(&goal); err != nil {
		t.Fatalf("query recovered story: %v", err)
	}
	if goal != "second" {
		t.Fatalf("recovered goal = %q, want second", goal)
	}
}

func TestPublicMutationReportsStableRepositoryLockTimeout(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	held, err := infrastructure.AcquireRepositoryLock(context.Background(), root, infrastructure.RepositoryLockExclusive)
	if err != nil {
		t.Fatalf("acquire held lock: %v", err)
	}
	defer held.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	cmd := NewRootCmd("dev")
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"story", "--slug", "blocked", "--goal", "timeout"})
	err = cmd.Execute()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != "repository_lock_timeout" || cliErr.Exit != 2 {
		t.Fatalf("story lock error = %T %v, want repository_lock_timeout exit 2", err, err)
	}
}

func TestReadOnlyOpenWaitsForPublicWriterAndReadsLatestCommit(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := executeRepositoryLockCommand("init", "--json"); err != nil {
		t.Fatalf("init: %v", err)
	}

	writerAcquired := make(chan struct{})
	releaseWriter := make(chan struct{})
	repositoryLockAcquiredHook = func(path string) {
		if path == "story" {
			close(writerAcquired)
			<-releaseWriter
		}
	}
	defer func() { repositoryLockAcquiredHook = nil }()

	writerDone := make(chan error, 1)
	go func() {
		_, err := executeRepositoryLockCommand("story", "--slug", "writer-commit", "--goal", "visible after writer release", "--json")
		writerDone <- err
	}()
	select {
	case <-writerAcquired:
	case <-time.After(2 * time.Second):
		t.Fatal("writer did not acquire the exclusive repository lock")
	}

	type readResult struct {
		db  *infrastructure.ReadOnlyDB
		err error
	}
	readerDone := make(chan readResult, 1)
	go func() {
		db, err := infrastructure.OpenReadOnly(dbPath)
		readerDone <- readResult{db: db, err: err}
	}()
	select {
	case result := <-readerDone:
		close(releaseWriter)
		if result.db != nil {
			result.db.Close()
		}
		t.Fatalf("reader opened while writer held the exclusive lock: %v", result.err)
	case <-time.After(100 * time.Millisecond):
	}

	close(releaseWriter)
	if err := <-writerDone; err != nil {
		t.Fatalf("writer command: %v", err)
	}
	var result readResult
	select {
	case result = <-readerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not open after writer released the lock")
	}
	if result.err != nil {
		t.Fatalf("OpenReadOnly after writer release: %v", result.err)
	}
	defer result.db.Close()
	var goal string
	if err := result.db.QueryRow(`SELECT goal FROM stories WHERE slug = 'writer-commit'`).Scan(&goal); err != nil {
		t.Fatalf("read writer commit: %v", err)
	}
	if goal != "visible after writer release" {
		t.Fatalf("goal = %q, want latest committed writer state", goal)
	}
}

func TestReadOnlyHandleBlocksPublicWriterUntilClose(t *testing.T) {
	t.Chdir(t.TempDir())
	writable, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := infrastructure.Migrate(writable); err != nil {
		writable.Close()
		t.Fatalf("Migrate: %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("close writable database: %v", err)
	}

	readerA, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly A: %v", err)
	}
	readerB, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		readerA.Close()
		t.Fatalf("OpenReadOnly B: %v", err)
	}
	if err := readerB.Close(); err != nil {
		readerA.Close()
		t.Fatalf("close reader B: %v", err)
	}

	jsonOutput = false
	cmd := NewRootCmd("dev")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"story", "--slug", "after-reader", "--goal", "writer waits for reader"})
	done := make(chan error, 1)
	go func() { done <- cmd.Execute() }()

	select {
	case err := <-done:
		readerA.Close()
		t.Fatalf("writer returned while reader held the shared lock: %v (output=%s)", err, output.String())
	case <-time.After(100 * time.Millisecond):
	}
	if err := readerA.Close(); err != nil {
		t.Fatalf("close reader A: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("writer after reader close: %v (output=%s)", err, output.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("writer did not proceed after reader closed")
	}

	readBack, err := infrastructure.OpenReadOnly(dbPath)
	if err != nil {
		t.Fatalf("OpenReadOnly after writer: %v", err)
	}
	defer readBack.Close()
	var goal string
	if err := readBack.QueryRow(`SELECT goal FROM stories WHERE slug = 'after-reader'`).Scan(&goal); err != nil {
		t.Fatalf("read committed writer state: %v", err)
	}
	if goal != "writer waits for reader" {
		t.Fatalf("goal = %q", goal)
	}
}

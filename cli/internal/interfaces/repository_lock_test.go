package interfaces

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"time"

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
	for _, path := range []string{"id", "scaffold", "preflight", "query", "next", "resume", "validate", "audit"} {
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

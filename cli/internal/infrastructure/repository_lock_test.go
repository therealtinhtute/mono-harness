//go:build linux || darwin

package infrastructure

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestRepositoryLockCompatibilityAndRelease(t *testing.T) {
	root := t.TempDir()
	before := directoryEntryNames(t, root)

	sharedA, err := AcquireRepositoryLockWithTimeout(context.Background(), root, RepositoryLockShared, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire first shared lock: %v", err)
	}
	sharedB, err := AcquireRepositoryLockWithTimeout(context.Background(), root, RepositoryLockShared, 100*time.Millisecond)
	if err != nil {
		sharedA.Close()
		t.Fatalf("acquire second shared lock: %v", err)
	}
	assertRepositoryLockTimeout(t, root, RepositoryLockExclusive)
	if err := sharedA.Close(); err != nil {
		t.Fatalf("close first shared lock: %v", err)
	}
	assertRepositoryLockTimeout(t, root, RepositoryLockExclusive)
	if err := sharedB.Close(); err != nil {
		t.Fatalf("close second shared lock: %v", err)
	}

	exclusive, err := AcquireRepositoryLockWithTimeout(context.Background(), root, RepositoryLockExclusive, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire exclusive lock after readers close: %v", err)
	}
	assertRepositoryLockTimeout(t, root, RepositoryLockShared)
	assertRepositoryLockTimeout(t, root, RepositoryLockExclusive)
	if err := exclusive.Close(); err != nil {
		t.Fatalf("close exclusive lock: %v", err)
	}
	if err := exclusive.Close(); err != nil {
		t.Fatalf("close exclusive lock again: %v", err)
	}

	sharedAfterRelease, err := AcquireRepositoryLockWithTimeout(context.Background(), root, RepositoryLockShared, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire shared lock after release: %v", err)
	}
	if err := sharedAfterRelease.Close(); err != nil {
		t.Fatalf("close final shared lock: %v", err)
	}

	after := directoryEntryNames(t, root)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("directory entries changed: before=%v after=%v", before, after)
	}
}

func TestRepositoryLockCanonicalizesSymlinkedRoot(t *testing.T) {
	parent := t.TempDir()
	realRoot := filepath.Join(parent, "real")
	linkRoot := filepath.Join(parent, "link")
	if err := os.Mkdir(realRoot, 0o755); err != nil {
		t.Fatalf("Mkdir real root: %v", err)
	}
	if err := os.Symlink(realRoot, linkRoot); err != nil {
		t.Fatalf("Symlink root: %v", err)
	}

	exclusive, err := AcquireRepositoryLockWithTimeout(context.Background(), realRoot, RepositoryLockExclusive, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire real-root lock: %v", err)
	}
	defer exclusive.Close()

	_, err = AcquireRepositoryLockWithTimeout(context.Background(), linkRoot, RepositoryLockShared, 25*time.Millisecond)
	var timeoutErr *RepositoryLockTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("symlink-root lock error = %v, want RepositoryLockTimeoutError", err)
	}
	canonical, err := filepath.EvalSymlinks(realRoot)
	if err != nil {
		t.Fatalf("EvalSymlinks real root: %v", err)
	}
	if timeoutErr.Root != canonical {
		t.Fatalf("timeout root = %q, want %q", timeoutErr.Root, canonical)
	}
}

func TestRepositoryLockHonorsCanceledContextBeforeAcquiring(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lock, err := AcquireRepositoryLockWithTimeout(ctx, t.TempDir(), RepositoryLockShared, 100*time.Millisecond)
	if lock != nil {
		lock.Close()
		t.Fatal("canceled context acquired a repository lock")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled lock error = %v, want context.Canceled", err)
	}
}

func assertRepositoryLockTimeout(t *testing.T, root string, mode RepositoryLockMode) {
	t.Helper()
	started := time.Now()
	_, err := AcquireRepositoryLockWithTimeout(context.Background(), root, mode, 25*time.Millisecond)
	var timeoutErr *RepositoryLockTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("acquire conflicting %s lock error = %v, want RepositoryLockTimeoutError", mode, err)
	}
	if elapsed := time.Since(started); elapsed < 20*time.Millisecond || elapsed > 500*time.Millisecond {
		t.Fatalf("conflicting %s lock elapsed = %s, want bounded short timeout", mode, elapsed)
	}
}

func directoryEntryNames(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}

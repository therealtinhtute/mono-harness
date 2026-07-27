package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const DefaultRepositoryLockTimeout = 5 * time.Second

type RepositoryLockMode uint8

const (
	RepositoryLockShared RepositoryLockMode = iota + 1
	RepositoryLockExclusive
)

func (m RepositoryLockMode) String() string {
	switch m {
	case RepositoryLockShared:
		return "shared"
	case RepositoryLockExclusive:
		return "exclusive"
	default:
		return "unknown"
	}
}

type RepositoryLockTimeoutError struct {
	Root string
	Mode RepositoryLockMode
}

func (e *RepositoryLockTimeoutError) Error() string {
	return fmt.Sprintf("repository %s lock timed out for %s", e.Mode, e.Root)
}

type RepositoryLockUnsupportedError struct {
	Platform string
}

func (e *RepositoryLockUnsupportedError) Error() string {
	return fmt.Sprintf("repository locking is unsupported on %s", e.Platform)
}

type RepositoryLock struct {
	root string
	mode RepositoryLockMode
	file *os.File

	closeOnce sync.Once
	closeErr  error
}

func AcquireRepositoryLock(ctx context.Context, root string, mode RepositoryLockMode) (*RepositoryLock, error) {
	return AcquireRepositoryLockWithTimeout(ctx, root, mode, DefaultRepositoryLockTimeout)
}

func AcquireRepositoryLockWithTimeout(ctx context.Context, root string, mode RepositoryLockMode, timeout time.Duration) (*RepositoryLock, error) {
	if mode != RepositoryLockShared && mode != RepositoryLockExclusive {
		return nil, fmt.Errorf("repository lock: invalid mode %d", mode)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("repository lock: timeout must be positive")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("repository lock: resolve root: %w", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("repository lock: resolve root: %w", err)
	}
	info, err := os.Stat(canonicalRoot)
	if err != nil {
		return nil, fmt.Errorf("repository lock: stat root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repository lock: root is not a directory: %s", canonicalRoot)
	}

	file, err := os.Open(canonicalRoot)
	if err != nil {
		return nil, fmt.Errorf("repository lock: open root: %w", err)
	}
	if err := acquirePlatformDirectoryLock(ctx, file, mode, timeout, canonicalRoot); err != nil {
		_ = file.Close()
		return nil, err
	}
	return &RepositoryLock{root: canonicalRoot, mode: mode, file: file}, nil
}

func (l *RepositoryLock) Close() error {
	if l == nil {
		return nil
	}
	l.closeOnce.Do(func() {
		unlockErr := unlockPlatformDirectory(l.file)
		closeErr := l.file.Close()
		l.closeErr = errors.Join(unlockErr, closeErr)
	})
	return l.closeErr
}

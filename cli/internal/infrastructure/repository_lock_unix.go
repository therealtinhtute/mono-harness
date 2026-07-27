//go:build linux || darwin

package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const repositoryLockRetryInterval = 10 * time.Millisecond

func acquirePlatformDirectoryLock(ctx context.Context, file *os.File, mode RepositoryLockMode, timeout time.Duration, root string) error {
	operation := unix.LOCK_NB
	if mode == RepositoryLockShared {
		operation |= unix.LOCK_SH
	} else {
		operation |= unix.LOCK_EX
	}

	deadline := time.Now().Add(timeout)
	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return &RepositoryLockTimeoutError{Root: root, Mode: mode}
			}
			return fmt.Errorf("repository lock: acquire %s: %w", mode, err)
		}
		err := unix.Flock(int(file.Fd()), operation)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, syscall.EINTR):
			continue
		case !errors.Is(err, syscall.EAGAIN) && !errors.Is(err, syscall.EWOULDBLOCK):
			return fmt.Errorf("repository lock: acquire %s: %w", mode, err)
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return &RepositoryLockTimeoutError{Root: root, Mode: mode}
		}
		wait := repositoryLockRetryInterval
		if remaining < wait {
			wait = remaining
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return &RepositoryLockTimeoutError{Root: root, Mode: mode}
			}
			return fmt.Errorf("repository lock: acquire %s: %w", mode, ctx.Err())
		case <-timer.C:
		}
	}
}

func unlockPlatformDirectory(file *os.File) error {
	for {
		err := unix.Flock(int(file.Fd()), unix.LOCK_UN)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return fmt.Errorf("repository lock: unlock: %w", err)
		}
		return nil
	}
}

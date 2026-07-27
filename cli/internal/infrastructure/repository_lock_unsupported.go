//go:build !linux && !darwin

package infrastructure

import (
	"context"
	"os"
	"runtime"
	"time"
)

func acquirePlatformDirectoryLock(_ context.Context, _ *os.File, _ RepositoryLockMode, _ time.Duration, _ string) error {
	return &RepositoryLockUnsupportedError{Platform: runtime.GOOS}
}

func unlockPlatformDirectory(_ *os.File) error {
	return nil
}

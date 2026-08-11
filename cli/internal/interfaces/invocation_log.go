package interfaces

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	invocationLogDir     = ".kit/log"
	invocationLogName    = "zharness.jsonl"
	invocationLogMaxSize = 1 << 20 // 1 MB, per R8 (docs/audit/sdlc-gap-analysis.md G3)
)

// logInvocation appends one JSONL line recording this invocation — R8: a
// failed lifecycle otherwise leaves no forensic record outside agent
// transcripts. Best-effort by design: any failure here (missing repo,
// full disk, permissions) is silently swallowed and never changes the
// command's own exit code or stdout/stderr (CONTRACT.md's shapes are
// untouched — this writes only to .kit/log/, never to the command's own
// output streams). argv omits argv[0] (the binary path) — an
// install-path detail, not part of what was actually invoked; the
// contract guarantees argv itself carries no secrets, so nothing here is
// redacted.
func logInvocation(argv []string, exit int, elapsed time.Duration, errorCode string) {
	if err := os.MkdirAll(invocationLogDir, 0o755); err != nil {
		return
	}
	path := filepath.Join(invocationLogDir, invocationLogName)
	rotateInvocationLogIfLarge(path)

	line, err := json.Marshal(map[string]any{
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"argv":       argv,
		"exit":       exit,
		"ms":         elapsed.Milliseconds(),
		"error_code": errorCode,
	})
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}

// rotateInvocationLogIfLarge renames the log to `.1` once it reaches the
// 1 MB cap, discarding any prior `.1` generation — a single-generation
// rotation is enough for a best-effort forensic log, not a retained audit
// trail.
func rotateInvocationLogIfLarge(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < invocationLogMaxSize {
		return
	}
	_ = os.Rename(path, path+".1")
}

// invocationErrorCode extracts the machine-readable error code logInvocation
// records, mirroring handleError's own code resolution (errors.go) without
// duplicating its stdout/stderr writing — a silentExit's body was already
// written elsewhere, so it carries no code of its own.
func invocationErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if _, ok := err.(*silentExit); ok {
		return ""
	}
	if ce, ok := err.(*cliError); ok {
		return ce.Code
	}
	return "internal_error"
}

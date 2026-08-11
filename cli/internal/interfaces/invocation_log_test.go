package interfaces

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLogInvocationWritesJSONLine is R8's round trip
// (docs/audit/sdlc-gap-analysis.md G3): every invocation appends one JSONL
// line carrying ts/argv/exit/ms/error_code, without touching any command's
// own stdout/stderr.
func TestLogInvocationWritesJSONLine(t *testing.T) {
	t.Chdir(t.TempDir())

	logInvocation([]string{"preflight", "watzup", "--json"}, 0, 12*time.Millisecond, "")

	data, err := os.ReadFile(filepath.Join(".kit", "log", "zharness.jsonl"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("log lines = %d, want 1", len(lines))
	}

	var entry struct {
		TS        string   `json:"ts"`
		Argv      []string `json:"argv"`
		Exit      int      `json:"exit"`
		MS        int64    `json:"ms"`
		ErrorCode string   `json:"error_code"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("decode log line %q: %v", lines[0], err)
	}
	if entry.TS == "" {
		t.Fatal("entry.TS is empty, want an RFC3339 timestamp")
	}
	if len(entry.Argv) != 3 || entry.Argv[0] != "preflight" {
		t.Fatalf("entry.Argv = %v, want [preflight watzup --json]", entry.Argv)
	}
	if entry.Exit != 0 {
		t.Fatalf("entry.Exit = %d, want 0", entry.Exit)
	}
	if entry.MS != 12 {
		t.Fatalf("entry.MS = %d, want 12", entry.MS)
	}
	if entry.ErrorCode != "" {
		t.Fatalf("entry.ErrorCode = %q, want empty on success", entry.ErrorCode)
	}
}

// TestLogInvocationAppendsAcrossCalls proves successive invocations append
// rather than overwrite — the log is a history, not a snapshot.
func TestLogInvocationAppendsAcrossCalls(t *testing.T) {
	t.Chdir(t.TempDir())

	logInvocation([]string{"query", "phases", "--json"}, 0, time.Millisecond, "")
	logInvocation([]string{"trace", "add", "--wave", "1"}, 2, time.Millisecond, "db_unreadable")

	data, err := os.ReadFile(filepath.Join(".kit", "log", "zharness.jsonl"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("log lines = %d, want 2", len(lines))
	}

	var second struct {
		ErrorCode string `json:"error_code"`
		Exit      int    `json:"exit"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatalf("decode second line %q: %v", lines[1], err)
	}
	if second.Exit != 2 || second.ErrorCode != "db_unreadable" {
		t.Fatalf("second entry = %+v, want exit=2 error_code=db_unreadable", second)
	}
}

// TestLogInvocationRotatesAtSizeCap proves the log rotates to `.1` once it
// reaches the 1 MB cap, and that the new entry lands in a fresh file, not
// appended past the cap forever.
func TestLogInvocationRotatesAtSizeCap(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(invocationLogDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(invocationLogDir, invocationLogName)
	oversized := make([]byte, invocationLogMaxSize+1)
	for i := range oversized {
		oversized[i] = 'x'
	}
	if err := os.WriteFile(path, oversized, 0o644); err != nil {
		t.Fatalf("WriteFile oversized log: %v", err)
	}

	logInvocation([]string{"resume", "--json"}, 0, time.Millisecond, "")

	rotated, err := os.ReadFile(path + ".1")
	if err != nil {
		t.Fatalf("ReadFile rotated log: %v", err)
	}
	if len(rotated) != invocationLogMaxSize+1 {
		t.Fatalf("rotated log size = %d, want the original oversized content preserved (%d)", len(rotated), invocationLogMaxSize+1)
	}

	fresh, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile fresh log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(fresh)), "\n")
	if len(lines) != 1 {
		t.Fatalf("fresh log lines = %d, want 1 (only the post-rotation entry)", len(lines))
	}
}

// TestLogInvocationNeverPanicsWithoutRepo proves the best-effort contract:
// when .kit/log can't be created (a file sits where the directory should
// be), logInvocation swallows the failure instead of panicking or erroring
// out to the caller — R8's "never alters the exit code or output" clause.
func TestLogInvocationNeverPanicsWithoutRepo(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".kit", []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	logInvocation([]string{"preflight", "watzup", "--json"}, 0, time.Millisecond, "")
	// No assertion beyond "did not panic" — .kit is a file, so MkdirAll(".kit/log")
	// must fail, and logInvocation must have returned silently.
}

func TestInvocationErrorCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil", nil, ""},
		{"silent exit", newSilentExit(1), ""},
		{"cli error", newUserError("invalid_mode", "bad mode"), "invalid_mode"},
		{"unclassified error", os.ErrNotExist, "internal_error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := invocationErrorCode(c.err); got != c.want {
				t.Fatalf("invocationErrorCode(%v) = %q, want %q", c.err, got, c.want)
			}
		})
	}
}

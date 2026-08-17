package interfaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/therealtinhtute/skills/cli/internal/application"
	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// cliError is a command failure carrying CONTRACT.md's stable
// machine-readable code and exit status (1 user error, 2 system error).
type cliError struct {
	Code    string
	Message string
	Exit    int
}

func (e *cliError) Error() string { return e.Message }

// newUserError builds a validation/user error (exit 1).
func newUserError(code, message string) *cliError {
	return &cliError{Code: code, Message: message, Exit: 1}
}

// newSystemError builds a system error (exit 2).
func newSystemError(code, message string) *cliError {
	return &cliError{Code: code, Message: message, Exit: 2}
}

func mapRepositoryLockError(prefix string, err error) *cliError {
	var timeout *infrastructure.RepositoryLockTimeoutError
	if errors.As(err, &timeout) {
		return newSystemError("repository_lock_timeout", fmt.Sprintf("%s: %v", prefix, timeout))
	}
	var unsupported *infrastructure.RepositoryLockUnsupportedError
	if errors.As(err, &unsupported) {
		return newSystemError("repository_lock_unsupported", fmt.Sprintf("%s: %v", prefix, unsupported))
	}
	return newSystemError("repository_lock_failed", fmt.Sprintf("%s: %v", prefix, err))
}

func mapReadOnlyOpenError(prefix string, err error) *cliError {
	var timeout *infrastructure.RepositoryLockTimeoutError
	var unsupported *infrastructure.RepositoryLockUnsupportedError
	if errors.As(err, &timeout) || errors.As(err, &unsupported) {
		return mapRepositoryLockError(prefix, err)
	}
	return newSystemError("db_unreadable", fmt.Sprintf("%s: %v", prefix, err))
}

// mapValidationError converts a domain/application validation failure into
// a cliError. An empty Code means the rule is an internal invariant that
// CONTRACT.md does not enumerate (not reachable via a documented command
// argument in the normal case) — "invalid_input" is the fallback so it
// still renders a well-formed error envelope instead of internal_error.
func mapValidationError(ve *domain.ValidationError) *cliError {
	code := ve.Code
	if code == "" {
		code = "invalid_input"
	}
	return newUserError(code, ve.Message)
}

// mapStop converts a shared active-plan resolver Stop (R2,
// docs/audit/consumer-adoption-audit.md D1) into the same non-zero-exit
// JSON envelope a ValidationError produces, so a Stop-based caller and a
// ValidationError-based caller are indistinguishable at the process
// boundary.
func mapStop(stop *application.StopInfo) *cliError {
	return newUserError(stop.Code, stop.Message)
}

// errNotImplemented marks a command stub not yet wired to the
// application layer. Later cli-core waves replace these calls.
func errNotImplemented(cmdName string) error {
	return newUserError("not_implemented", fmt.Sprintf("%s: not yet implemented", cmdName))
}

// silentExit signals a non-zero exit code for a command whose own
// structured body (e.g. validate's {"valid": false, "findings": [...]})
// has already been written to stdout — unlike cliError, handleError must
// not print anything further for it.
type silentExit struct{ code int }

func (e *silentExit) Error() string { return "" }

func newSilentExit(code int) error { return &silentExit{code: code} }

// handleError renders a command failure per CONTRACT.md's error
// preamble: `{"error": {"code": "...", "message": "..."}}` under --json,
// plain text otherwise. Returns the process exit code.
func handleError(w io.Writer, err error) int {
	if se, ok := err.(*silentExit); ok {
		return se.code
	}
	code, msg, exit := "internal_error", err.Error(), 1
	if ce, ok := err.(*cliError); ok {
		code, msg, exit = ce.Code, ce.Message, ce.Exit
	}
	if jsonOutput {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": code, "message": msg},
		})
	} else {
		fmt.Fprintln(w, msg)
	}
	return exit
}

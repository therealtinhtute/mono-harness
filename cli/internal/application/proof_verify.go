package application

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// proofVerificationTimeout bounds one proof command's re-execution so a
// hung command cannot hang `check record` indefinitely.
const proofVerificationTimeout = 5 * time.Minute

// verifyProofLinks re-executes each proof link's command and requires exit
// code 0 before `check record` accepts the verdict — declared-string proof
// (validated for shape only, D5) was a deliberate scope cut the owner
// later reversed (D22, docs/plans/completed/harness-memory-ceremony-
// convergence.md). Only exit code is checked, not output text: a command
// like `go test ./...` is not guaranteed byte-identical between runs
// (timestamps, parallel test ordering), so comparing text would produce
// false verification failures for genuinely passing proof.
//
// Only APPROVED and APPROVE_WITH_REQUESTS proof is verified. Those
// verdicts vouch that something works, so their proof must actually pass.
// REQUEST_CHANGES proof commonly demonstrates the opposite — a failing
// test or a broken build, proving the problem — so requiring exit 0 there
// would reject exactly the evidence a REQUEST_CHANGES verdict needs to
// carry. Callers select verification by only invoking this for the two
// verdicts it applies to.
func verifyProofLinks(proofLinks []domain.ProofLink) error {
	for _, pl := range proofLinks {
		if err := runProofCommand(pl.Command); err != nil {
			return &domain.ValidationError{
				Code:    "proof_verification_failed",
				Message: fmt.Sprintf("check record: proof command failed verification: %v", err),
			}
		}
	}
	return nil
}

func runProofCommand(command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), proofVerificationTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%q: %v\n%s", command, err, lastLines(out.String(), 20))
	}
	return nil
}

// lastLines returns at most the last n lines of s, for a bounded error
// message instead of an unbounded command's full captured output.
func lastLines(s string, n int) string {
	trimmed := strings.TrimRight(s, "\n")
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

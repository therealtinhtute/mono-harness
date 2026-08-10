package domain

import (
	"fmt"
	"strings"
)

// Check verdict enum (CONTRACT.md `check record`).
const (
	VerdictApproved            = "APPROVED"
	VerdictApproveWithRequests = "APPROVE_WITH_REQUESTS"
	VerdictRequestChanges      = "REQUEST_CHANGES"
)

var checkVerdicts = map[string]bool{
	VerdictApproved:            true,
	VerdictApproveWithRequests: true,
	VerdictRequestChanges:      true,
}

func IsValidCheckVerdict(verdict string) bool {
	return checkVerdicts[verdict]
}

// Check judge enum (CONTRACT.md `check record`, eval-layer initiative):
// declares whether the verdict was produced by an independent judge or by
// the same session that authored the diff under review.
const (
	JudgeIndependent = "independent"
	JudgeSameSession = "same-session"
)

var checkJudges = map[string]bool{
	JudgeIndependent: true,
	JudgeSameSession: true,
}

func IsValidJudge(judge string) bool {
	return checkJudges[judge]
}

// ProofLink is one entry of a Check's proof_links JSON array.
type ProofLink struct {
	Command      string `json:"command"`
	OutputRef    string `json:"output_ref"`
	ArtifactPath string `json:"artifact_path"`
}

// Check is a persisted gate verdict.
type Check struct {
	ID           string
	RunID        string
	Verdict      string
	Judge        string
	JudgeModel   string
	ProofLinks   []ProofLink
	ArtifactPath *string
	CreatedAt    string
}

func (c Check) Validate() error {
	if c.RunID == "" {
		return &ValidationError{Code: "missing_required_field", Message: "check: run_id is required"}
	}
	if !IsValidCheckVerdict(c.Verdict) {
		return &ValidationError{Code: "invalid_verdict", Message: fmt.Sprintf("check: invalid verdict %q", c.Verdict)}
	}
	if c.Verdict != VerdictRequestChanges && len(c.ProofLinks) == 0 {
		return &ValidationError{Code: "empty_proof_links", Message: fmt.Sprintf("check: proof_links required for verdict %q", c.Verdict)}
	}
	for _, pl := range c.ProofLinks {
		if strings.TrimSpace(pl.Command) == "" {
			return &ValidationError{Code: "invalid_proof_links", Message: "check: proof_links[].command must not be blank"}
		}
	}
	if !IsValidJudge(c.Judge) {
		return &ValidationError{Code: "invalid_judge", Message: fmt.Sprintf("check: invalid judge %q", c.Judge)}
	}
	if strings.TrimSpace(c.JudgeModel) == "" {
		return &ValidationError{Code: "missing_required_field", Message: "check: judge_model is required"}
	}
	return nil
}

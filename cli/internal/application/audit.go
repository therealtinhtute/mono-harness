package application

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// AuditFinding is one entry in audit's contract_violations or
// unlinked_proofs arrays.
type AuditFinding struct {
	Link   string `json:"link"`
	Issue  string `json:"issue"`
	Detail string `json:"detail"`
}

// AuditReport mirrors CONTRACT.md's locked `audit --json` shape.
type AuditReport struct {
	PointerDrift       []DriftFinding `json:"pointer_drift"`
	ContractViolations []AuditFinding `json:"contract_violations"`
	UnlinkedProofs     []AuditFinding `json:"unlinked_proofs"`
	EntropyScore       int            `json:"entropy_score"`
}

// Audit composes Resume (pointer_drift) and Validate (contract_violations)
// unchanged — per validation-gate-PLAN.md's T2 "avoid: duplicating
// validate logic — audit composes it and adds scoring" — and adds one
// genuinely new check: unlinked_proofs, a sweep of every recorded check's
// proof_links for artifact_path entries that no longer resolve on disk.
// entropy_score is an upstream-style (HARNESS_AUDIT.md) weighted sum
// capped at 100; the three categories are zharness's own since
// CONTRACT.md's locked audit shape has no story/decision/backlog/tool
// categories to score upstream's exact formula against (see
// .kit/implementation-notes.md).
func Audit(db *sql.DB, root, cliVersion string) (AuditReport, error) {
	report := AuditReport{
		PointerDrift:       []DriftFinding{},
		ContractViolations: []AuditFinding{},
		UnlinkedProofs:     []AuditFinding{},
	}

	resumeView, err := Resume(db, cliVersion)
	if err != nil {
		return report, fmt.Errorf("audit: %w", err)
	}
	report.PointerDrift = resumeView.Drift

	validateResult, err := Validate(db, root)
	if err != nil {
		return report, fmt.Errorf("audit: %w", err)
	}
	for _, f := range validateResult.Findings {
		report.ContractViolations = append(report.ContractViolations, AuditFinding{Link: f.Link, Issue: f.Issue, Detail: f.Detail})
	}

	unlinked, err := unlinkedProofs(db)
	if err != nil {
		return report, fmt.Errorf("audit: %w", err)
	}
	report.UnlinkedProofs = unlinked

	report.EntropyScore = entropyScore(report.PointerDrift, report.ContractViolations, report.UnlinkedProofs)
	return report, nil
}

// unlinkedProofs sweeps every recorded check's proof_links for entries
// whose artifact_path is empty or no longer resolves on disk. Ordered by
// check id (ULID-sortable, so chronological) for determinism.
func unlinkedProofs(db *sql.DB) ([]AuditFinding, error) {
	findings := []AuditFinding{}
	if db == nil {
		return findings, nil
	}

	rows, err := db.Query(`SELECT id, proof_links FROM checks ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query checks: %w", err)
	}
	defer rows.Close()

	type checkRow struct{ id, proofLinksRaw string }
	var checkRows []checkRow
	for rows.Next() {
		var cr checkRow
		if err := rows.Scan(&cr.id, &cr.proofLinksRaw); err != nil {
			return nil, fmt.Errorf("scan check: %w", err)
		}
		checkRows = append(checkRows, cr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checks: %w", err)
	}

	for _, cr := range checkRows {
		var links []domain.ProofLink
		if err := json.Unmarshal([]byte(cr.proofLinksRaw), &links); err != nil {
			findings = append(findings, AuditFinding{
				Link:   "check",
				Issue:  "invalid_proof_links",
				Detail: fmt.Sprintf("check %s proof_links is not valid JSON: %v", cr.id, err),
			})
			continue
		}
		for _, l := range links {
			if l.ArtifactPath == "" {
				findings = append(findings, AuditFinding{
					Link:   "check",
					Issue:  "unlinked_proof",
					Detail: fmt.Sprintf("check %s proof link (command=%q) has no artifact_path", cr.id, l.Command),
				})
				continue
			}
			if !infrastructure.Exists(l.ArtifactPath) {
				findings = append(findings, AuditFinding{
					Link:   "check",
					Issue:  "unlinked_proof",
					Detail: fmt.Sprintf("check %s proof link artifact_path %q not found on disk", cr.id, l.ArtifactPath),
				})
			}
		}
	}
	return findings, nil
}

// entropyScore mirrors HARNESS_AUDIT.md's weighted-and-capped shape.
// Weights are zharness's own (see Audit's doc comment): pointer_drift is
// the most structural break (10), unlinked_proofs is a broken-evidence
// signal analogous to upstream's broken_tools (8), contract_violations is
// a lower-severity doc/format issue (5).
func entropyScore(drift []DriftFinding, violations []AuditFinding, unlinked []AuditFinding) int {
	score := 10*len(drift) + 5*len(violations) + 8*len(unlinked)
	if score > 100 {
		score = 100
	}
	return score
}

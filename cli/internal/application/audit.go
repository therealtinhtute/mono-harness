package application

import (
	"database/sql"
	"fmt"
)

// AuditFinding is one entry in audit's contract_violations or
// unlinked_proofs arrays.
type AuditFinding struct {
	Link   string `json:"link"`
	Issue  string `json:"issue"`
	Detail string `json:"detail"`
}

// AuditReport preserves the public `audit --json` shape.
type AuditReport struct {
	PointerDrift       []DriftFinding `json:"pointer_drift"`
	ContractViolations []AuditFinding `json:"contract_violations"`
	UnlinkedProofs     []AuditFinding `json:"unlinked_proofs"`
}

// Audit composes the DB-backed Resume and Validate readers. UnlinkedProofs is
// retained as an empty compatibility field: proof-link artifact paths are
// optional legacy metadata, not lifecycle integrity requirements.
func Audit(db *sql.DB, cliVersion string) (AuditReport, error) {
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

	validateResult, err := Validate(db)
	if err != nil {
		return report, fmt.Errorf("audit: %w", err)
	}
	for _, finding := range validateResult.Findings {
		report.ContractViolations = append(report.ContractViolations, AuditFinding{
			Link: finding.Link, Issue: finding.Issue, Detail: finding.Detail,
		})
	}

	return report, nil
}

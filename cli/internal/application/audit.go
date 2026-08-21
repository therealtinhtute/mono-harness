package application

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/therealtinhtute/skills/cli/internal/embedded"
)

// AuditFinding is one entry in audit's contract_violations or
// unlinked_proofs arrays.
type AuditFinding struct {
	Link       string `json:"link"`
	Issue      string `json:"issue"`
	Detail     string `json:"detail"`
	Identifier string `json:"identifier,omitempty"`
	Severity   string `json:"severity,omitempty"`
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
func Audit(db *sql.DB, cliVersion, repoRoot string) (AuditReport, error) {
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

	finding, present, err := authoredDocsFinding(repoRoot)
	if err != nil {
		return report, fmt.Errorf("audit: %w", err)
	}
	if present {
		report.ContractViolations = append(report.ContractViolations, finding)
	}

	return report, nil
}

func authoredDocsFinding(repoRoot string) (AuditFinding, bool, error) {
	managedPaths := map[string]struct{}{}
	managedPresent := false
	err := fs.WalkDir(embedded.FS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relative := filepath.ToSlash(path)
		target := filepath.Join(repoRoot, filepath.FromSlash(relative))
		if relative != "AGENTS.md" {
			docRelative := filepath.ToSlash(filepath.Join("docs", filepath.FromSlash(relative)))
			managedPaths[docRelative] = struct{}{}
			target = filepath.Join(repoRoot, filepath.FromSlash(docRelative))
		}
		if _, err := os.Stat(target); err == nil {
			managedPresent = true
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat managed document %s: %w", relative, err)
		}
		return nil
	})
	if err != nil {
		return AuditFinding{}, false, fmt.Errorf("walk embedded documents: %w", err)
	}
	if !managedPresent {
		return AuditFinding{}, false, nil
	}

	docsRoot := filepath.Join(repoRoot, "docs")
	authoredMarkdown := false
	err = filepath.WalkDir(docsRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		relative, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if _, managed := managedPaths[filepath.ToSlash(relative)]; !managed {
			authoredMarkdown = true
		}
		return nil
	})
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return AuditFinding{}, false, fmt.Errorf("walk authored documents: %w", err)
	}
	if authoredMarkdown {
		return AuditFinding{}, false, nil
	}

	return AuditFinding{
		Link:       "REPO->AUTHORED_DOCS",
		Issue:      "missing_authored_docs",
		Detail:     "managed documentation is present on disk, but docs/ contains no authored Markdown outside the managed path set; this reports documentation presence only",
		Identifier: "authored_docs_missing",
		Severity:   "warning",
	}, true, nil
}

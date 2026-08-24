package application

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

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

	pins, err := scanPinnedDocs(repoRoot)
	if err != nil {
		return report, fmt.Errorf("audit: %w", err)
	}
	for _, doc := range pins {
		finding, present, err := pinDriftFinding(repoRoot, doc)
		if err != nil {
			return report, fmt.Errorf("audit: %w", err)
		}
		if present {
			report.ContractViolations = append(report.ContractViolations, finding)
		}
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

// docPinPattern matches the opt-in pin declaration an authored document may
// carry — a single HTML comment naming the commit its source citations were
// verified against, e.g. `<!-- zharness:pin 1a2b3c4d -->`. A document with no
// pin is never reported (R5): pinning is opt-in and its absence is not a
// defect.
var docPinPattern = regexp.MustCompile(`(?m)<!--\s*zharness:pin\s+([0-9a-fA-F]{7,40})\s*-->`)

// docCitationPattern matches repository-relative source citations of the form
// `path/to/file.ext:NN` — the shape docs/ARCHITECTURE.md uses for every
// citation it carries. The `:NN` suffix is required so prose paths and glob
// placeholders never match.
var docCitationPattern = regexp.MustCompile(`(?:[A-Za-z0-9_][A-Za-z0-9_.\-]*/)*[A-Za-z0-9_][A-Za-z0-9_.\-]*\.[A-Za-z0-9]+:[0-9]+`)

// pinnedDoc is one eligible document that declares a pin.
type pinnedDoc struct {
	Name      string   // file name under docs/, e.g. "ARCHITECTURE.md"
	Pin       string   // abbreviated or full SHA exactly as written
	Citations []string // unique path:line tokens, sorted
}

// scanPinnedDocs reads only top-level docs/*.md files (R4: directory scoping
// by construction — anything under docs/plans/, docs/decisions/, docs/audit/,
// docs/memory/, docs/research/ lives in a subdirectory and is never read for
// pins). Documents without a pin are skipped silently (R5).
func scanPinnedDocs(repoRoot string) ([]pinnedDoc, error) {
	entries, err := os.ReadDir(filepath.Join(repoRoot, "docs"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read docs directory: %w", err)
	}
	var found []pinnedDoc
	for _, entry := range entries { // ReadDir sorts by name; iteration order is deterministic
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(repoRoot, "docs", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		match := docPinPattern.FindSubmatch(data)
		if match == nil {
			continue
		}
		seen := map[string]bool{}
		var citations []string
		for _, c := range docCitationPattern.FindAllString(string(data), -1) {
			if !seen[c] {
				seen[c] = true
				citations = append(citations, c)
			}
		}
		sort.Strings(citations)
		found = append(found, pinnedDoc{Name: entry.Name(), Pin: string(match[1]), Citations: citations})
	}
	return found, nil
}

// citationDrift is one cited path measured against its document's pin.
type citationDrift struct {
	Citation     string
	Moved        bool // at least one commit touched the path after the pin
	LinesAdded   int
	LinesRemoved int
	Missing      bool // the cited path does not exist in the working tree — distinct from moved
}

// measureCitation compares one cited path against the pin using read-only git
// commands (R7: audit performs no repin and writes nothing; GIT_OPTIONAL_LOCKS=0
// keeps even the index-refresh write git would otherwise opportunistically take).
func measureCitation(repoRoot, pin, citation string) (citationDrift, error) {
	drift := citationDrift{Citation: citation}
	path := citation[:strings.LastIndex(citation, ":")]
	target := filepath.Join(repoRoot, filepath.FromSlash(path))
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			drift.Missing = true
			return drift, nil
		}
		return drift, fmt.Errorf("stat cited path %s: %w", path, err)
	}

	env := append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	run := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoRoot
		cmd.Env = env
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	countOut, err := run("rev-list", "--count", pin+"..HEAD", "--", path)
	if err != nil {
		return drift, err
	}
	movedCount, err := strconv.Atoi(countOut)
	if err != nil {
		return drift, fmt.Errorf("parse rev-list count for %s: %w", path, err)
	}
	if movedCount == 0 {
		return drift, nil
	}
	drift.Moved = true

	numstatOut, err := run("diff", "--numstat", pin+"..HEAD", "--", path)
	if err != nil {
		return drift, err
	}
	for _, line := range strings.Split(numstatOut, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}
		// Binary entries report "-" instead of counts; they still count as movement.
		if n, err := strconv.Atoi(fields[0]); err == nil {
			drift.LinesAdded += n
		} else {
			drift.LinesAdded++
		}
		if n, err := strconv.Atoi(fields[1]); err == nil {
			drift.LinesRemoved += n
		} else {
			drift.LinesRemoved++
		}
	}
	return drift, nil
}

// pinDriftFinding composes R3's finding: it names the document, each moved
// citation, and the size of each change. Missing paths are listed distinctly
// from moved ones. Wording states freshness only (R9): a pinned, unmoved
// document is not claimed accurate anywhere in the text.
func pinDriftFinding(repoRoot string, doc pinnedDoc) (AuditFinding, bool, error) {
	var moved, missing []string
	for _, citation := range doc.Citations {
		drift, err := measureCitation(repoRoot, doc.Pin, citation)
		if err != nil {
			return AuditFinding{}, false, err
		}
		switch {
		case drift.Missing:
			missing = append(missing, drift.Citation)
		case drift.Moved:
			moved = append(moved, fmt.Sprintf("%s (+%d/-%d)", drift.Citation, drift.LinesAdded, drift.LinesRemoved))
		}
	}
	if len(moved) == 0 && len(missing) == 0 {
		return AuditFinding{}, false, nil
	}

	var detail strings.Builder
	fmt.Fprintf(&detail, "%s pins commit %s. Since that commit its cited sources have moved — a freshness signal, not a correctness judgment.", doc.Name, doc.Pin)
	if len(moved) > 0 {
		fmt.Fprintf(&detail, " Moved citations: %s.", strings.Join(moved, "; "))
	}
	if len(missing) > 0 {
		fmt.Fprintf(&detail, " Citations whose path no longer exists in the working tree (reported separately from moved): %s.", strings.Join(missing, "; "))
	}

	return AuditFinding{
		Link:       "DOCS->SOURCE",
		Issue:      "pinned_doc_drift",
		Detail:     detail.String(),
		Identifier: "authored_doc_pin_drift",
		Severity:   "info",
	}, true, nil
}

package application

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ValidateFinding and ValidateResult mirror CONTRACT.md's `validate`
// --json shape exactly: {"valid": bool, "findings": [{"link", "issue",
// "detail"}]}.
type ValidateFinding struct {
	Link   string `json:"link"`
	Issue  string `json:"issue"`
	Detail string `json:"detail"`
}

type ValidateResult struct {
	Valid    bool              `json:"valid"`
	Findings []ValidateFinding `json:"findings"`
}

type runRef struct {
	id, phase string
}

type checkRef struct {
	id, runID string
}

// Validate walks SPEC->PLAN->RUN->CHECK->HANDOFF by frontmatter
// cross-links under root (the .kit/-equivalent directory). db is
// optional: nil skips the freshness-vs-DB checks, since cross-checking
// the documents against each other doesn't require a live harness (same
// "no db is a valid state" precedent as Resume).
func Validate(db *sql.DB, root string) (ValidateResult, error) {
	var findings []ValidateFinding

	specFields, specExists, err := parseFrontmatter(filepath.Join(root, "planning", "SPEC.md"))
	if err != nil {
		return ValidateResult{}, err
	}
	switch {
	case !specExists:
		findings = append(findings, ValidateFinding{
			Link: "SPEC->PLAN", Issue: "missing_key",
			Detail: "planning/SPEC.md not found",
		})
	case specFields["id"] == "":
		findings = append(findings, ValidateFinding{
			Link: "SPEC->PLAN", Issue: "missing_key",
			Detail: "planning/SPEC.md missing required key \"id\"",
		})
	case !looksLikeULID(specFields["id"]):
		findings = append(findings, ValidateFinding{
			Link: "SPEC->PLAN", Issue: "missing_key",
			Detail: fmt.Sprintf("planning/SPEC.md key \"id\" value %q is not a valid ULID", specFields["id"]),
		})
	default:
		// Known gap (CONTRACT.md): PLAN artifacts carry no spec_id field
		// yet, so SPEC->PLAN can never be checked by ID-equality — report
		// it as not-yet-implemented rather than a hard failure.
		findings = append(findings, ValidateFinding{
			Link: "SPEC->PLAN", Issue: "not_yet_implemented",
			Detail: "PLAN artifacts don't carry a spec_id field yet; SPEC->PLAN cannot be cross-checked",
		})
	}

	runFiles, err := globMD(filepath.Join(root, "runs", "work"))
	if err != nil {
		return ValidateResult{}, err
	}
	sort.Strings(runFiles)

	var runs []runRef
	for _, rf := range runFiles {
		fields, _, err := parseFrontmatter(rf)
		if err != nil {
			return ValidateResult{}, err
		}
		ctx := fmt.Sprintf("RUN %s", rf)

		id, idOK := requireULID(&findings, "PLAN->RUN", fields, "id", ctx)
		phase, phaseOK := requireKey(&findings, "PLAN->RUN", fields, "phase", ctx)
		requireULID(&findings, "PLAN->RUN", fields, "plan_id", ctx)

		if phaseOK {
			planPath := filepath.Join(root, "planning", "phases", phase, phase+"-PLAN.md")
			if _, statErr := os.Stat(planPath); statErr != nil {
				findings = append(findings, ValidateFinding{
					Link: "PLAN->RUN", Issue: "broken_link",
					Detail: fmt.Sprintf("%s references phase %q but %s does not exist", ctx, phase, planPath),
				})
			}
		}
		if idOK {
			if db != nil {
				exists, err := rowExists(db, "runs", id)
				if err != nil {
					return ValidateResult{}, err
				}
				if !exists {
					findings = append(findings, ValidateFinding{
						Link: "PLAN->RUN", Issue: "stale_pointer",
						Detail: fmt.Sprintf("%s id %q has no matching row in the runs table", ctx, id),
					})
				}
			}
			runs = append(runs, runRef{id: id, phase: phase})
		}
	}

	checkFiles, err := globMD(filepath.Join(root, "reports", "check"))
	if err != nil {
		return ValidateResult{}, err
	}
	sort.Strings(checkFiles)

	var checks []checkRef
	for _, cf := range checkFiles {
		fields, _, err := parseFrontmatter(cf)
		if err != nil {
			return ValidateResult{}, err
		}
		ctx := fmt.Sprintf("CHECK %s", cf)

		id, idOK := requireULID(&findings, "RUN->CHECK", fields, "id", ctx)
		runID, runIDOK := requireULID(&findings, "RUN->CHECK", fields, "run_id", ctx)

		if runIDOK && !containsRunID(runs, runID) {
			findings = append(findings, ValidateFinding{
				Link: "RUN->CHECK", Issue: "broken_link",
				Detail: fmt.Sprintf("%s run_id %q does not match any known RUN", ctx, runID),
			})
		}
		if idOK {
			if db != nil {
				exists, err := rowExists(db, "checks", id)
				if err != nil {
					return ValidateResult{}, err
				}
				if !exists {
					findings = append(findings, ValidateFinding{
						Link: "RUN->CHECK", Issue: "stale_pointer",
						Detail: fmt.Sprintf("%s id %q has no matching row in the checks table", ctx, id),
					})
				}
			}
			checks = append(checks, checkRef{id: id, runID: runID})
		}
	}

	handoffFields, handoffExists, err := parseFrontmatter(filepath.Join(root, "HANDOFF.md"))
	if err != nil {
		return ValidateResult{}, err
	}
	if handoffExists {
		ctx := "HANDOFF.md"
		requireULID(&findings, "CHECK->HANDOFF", handoffFields, "id", ctx)
		runID, runIDOK := requireULID(&findings, "CHECK->HANDOFF", handoffFields, "run_id", ctx)
		checkID, checkIDOK := requireULID(&findings, "CHECK->HANDOFF", handoffFields, "check_id", ctx)

		if runIDOK && !containsRunID(runs, runID) {
			findings = append(findings, ValidateFinding{
				Link: "CHECK->HANDOFF", Issue: "broken_link",
				Detail: fmt.Sprintf("%s run_id %q does not match any known RUN", ctx, runID),
			})
		}
		if checkIDOK && !containsCheckID(checks, checkID) {
			findings = append(findings, ValidateFinding{
				Link: "CHECK->HANDOFF", Issue: "broken_link",
				Detail: fmt.Sprintf("%s check_id %q does not match any known CHECK", ctx, checkID),
			})
		}
	}

	valid := true
	for _, f := range findings {
		if f.Issue != "not_yet_implemented" {
			valid = false
			break
		}
	}
	if findings == nil {
		findings = []ValidateFinding{}
	}
	return ValidateResult{Valid: valid, Findings: findings}, nil
}

// requireKey appends a missing_key finding when fields[key] is absent or
// empty, returning ok=false so callers can skip link checks that need it.
func requireKey(findings *[]ValidateFinding, link string, fields map[string]string, key, context string) (string, bool) {
	v, ok := fields[key]
	if !ok || v == "" {
		*findings = append(*findings, ValidateFinding{
			Link: link, Issue: "missing_key",
			Detail: fmt.Sprintf("%s missing required key %q", context, key),
		})
		return "", false
	}
	return v, true
}

// requireULID is requireKey plus a ULID-shape check on the value, per
// cli-domain-CONTEXT.md's Locked Decision that validate checks "required
// keys, link targets exist, ULID formats, pointer freshness vs DB state".
// CONTRACT.md's issue enum has no dedicated slot for a malformed-but-present
// value, so a bad format is reported as missing_key too — from a
// chain-walking consumer's perspective a garbage string is exactly as
// unusable as an absent one.
func requireULID(findings *[]ValidateFinding, link string, fields map[string]string, key, context string) (string, bool) {
	v, ok := requireKey(findings, link, fields, key, context)
	if !ok {
		return "", false
	}
	if !looksLikeULID(v) {
		*findings = append(*findings, ValidateFinding{
			Link: link, Issue: "missing_key",
			Detail: fmt.Sprintf("%s key %q value %q is not a valid ULID", context, key, v),
		})
		return "", false
	}
	return v, true
}

// looksLikeULID reports whether s has a ULID's shape: 26 characters, all
// drawn from Crockford's base32 alphabet.
func looksLikeULID(s string) bool {
	if len(s) != 26 {
		return false
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789ABCDEFGHJKMNPQRSTVWXYZ", c) {
			return false
		}
	}
	return true
}

func containsRunID(runs []runRef, id string) bool {
	for _, r := range runs {
		if r.id == id {
			return true
		}
	}
	return false
}

func containsCheckID(checks []checkRef, id string) bool {
	for _, c := range checks {
		if c.id == id {
			return true
		}
	}
	return false
}

func rowExists(db *sql.DB, table, id string) (bool, error) {
	var found string
	err := db.QueryRow(fmt.Sprintf(`SELECT id FROM %s WHERE id = ?`, table), id).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query %s %q: %w", table, id, err)
	}
	return true, nil
}

// globMD lists *.md files directly under dir, sorted by caller. A missing
// dir is not an error — it just yields no files (e.g. no runs recorded yet).
func globMD(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}

// parseFrontmatter extracts the flat `key: value` block between the
// leading `---` fences of an artifact file. Following legacy.go's
// parseFlatYAML precedent: no nesting/lists needed here (only scalar
// id/phase/*_id fields are read), so a full YAML parser isn't warranted.
// exists=false only when the file itself is absent; a file with no
// frontmatter block (e.g. PLAN.md, which has none) returns an empty map
// with exists=true.
func parseFrontmatter(path string) (fields map[string]string, exists bool, err error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}

	fields = map[string]string{}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fields, true, nil
	}
	for _, line := range lines[1:] {
		trimmed := strings.TrimRight(line, "\r")
		if strings.TrimSpace(trimmed) == "---" {
			break
		}
		idx := strings.Index(trimmed, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		val := strings.TrimSpace(trimmed[idx+1:])
		fields[key] = val
	}
	return fields, true, nil
}

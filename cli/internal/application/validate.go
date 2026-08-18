package application

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// ValidateFinding and ValidateResult preserve the public `validate --json`
// shape: {"valid": bool, "findings": [{"link", "issue", "detail"}]}.
type ValidateFinding struct {
	Link   string `json:"link"`
	Issue  string `json:"issue"`
	Detail string `json:"detail"`
}

type ValidateResult struct {
	Valid    bool              `json:"valid"`
	Findings []ValidateFinding `json:"findings"`
}

// Validate checks the durable lifecycle graph stored in the harness database.
// Legacy artifact paths are metadata only and are not part of lifecycle
// validity. A nil database produces a deterministic invalid result rather than
// claiming that an unchecked lifecycle is valid.
func Validate(db *sql.DB) (ValidateResult, error) {
	findings := []ValidateFinding{}
	if db == nil {
		findings = append(findings, ValidateFinding{
			Link:   "DB->LIFECYCLE",
			Issue:  "missing_key",
			Detail: "harness database is unavailable; lifecycle links cannot be validated",
		})
		return ValidateResult{Valid: false, Findings: findings}, nil
	}

	if err := validateEntityIDs(db, &findings); err != nil {
		return ValidateResult{}, err
	}
	if err := validateEntityEnums(db, &findings); err != nil {
		return ValidateResult{}, err
	}
	if err := validateStoryLinks(db, &findings); err != nil {
		return ValidateResult{}, err
	}
	if err := validateRunLinks(db, &findings); err != nil {
		return ValidateResult{}, err
	}
	if err := validateCheckLinks(db, &findings); err != nil {
		return ValidateResult{}, err
	}
	if err := validateHandoffLinks(db, &findings); err != nil {
		return ValidateResult{}, err
	}
	if err := validateMetaLinks(db, &findings); err != nil {
		return ValidateResult{}, err
	}

	return ValidateResult{Valid: len(findings) == 0, Findings: findings}, nil
}

func validateEntityIDs(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`
		SELECT entity, id FROM (
			SELECT 'STORY' AS entity, id FROM stories
			UNION ALL SELECT 'RUN', id FROM runs
			UNION ALL SELECT 'CHECK', id FROM checks
			UNION ALL SELECT 'HANDOFF', id FROM handoffs
			UNION ALL SELECT 'INTAKE', id FROM intakes
			UNION ALL SELECT 'TRACE', id FROM traces
		)
		ORDER BY entity, id
	`)
	if err != nil {
		return fmt.Errorf("validate entity ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entity, id string
		if err := rows.Scan(&entity, &id); err != nil {
			return fmt.Errorf("validate entity ids: %w", err)
		}
		if !looksLikeULID(id) {
			*findings = append(*findings, ValidateFinding{
				Link:   "DB->" + entity,
				Issue:  "missing_key",
				Detail: fmt.Sprintf("%s id %q is not a valid ULID", strings.ToLower(entity), id),
			})
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate entity ids: %w", err)
	}
	return nil
}

func validateEntityEnums(db *sql.DB, findings *[]ValidateFinding) error {
	if err := validateStoryStatuses(db, findings); err != nil {
		return err
	}
	return validateCheckVerdicts(db, findings)
}

func validateStoryStatuses(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`SELECT slug, status FROM stories ORDER BY created_at, id`)
	if err != nil {
		return fmt.Errorf("validate story statuses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slug, status string
		if err := rows.Scan(&slug, &status); err != nil {
			return fmt.Errorf("validate story statuses: %w", err)
		}
		if !domain.IsValidStoryStatus(status) {
			*findings = append(*findings, ValidateFinding{
				Link:   "DB->STORY",
				Issue:  "invalid_value",
				Detail: fmt.Sprintf("story %q has invalid status %q", slug, status),
			})
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate story statuses: %w", err)
	}
	return nil
}

func validateCheckVerdicts(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`SELECT id, verdict FROM checks ORDER BY created_at, id`)
	if err != nil {
		return fmt.Errorf("validate check verdicts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, verdict string
		if err := rows.Scan(&id, &verdict); err != nil {
			return fmt.Errorf("validate check verdicts: %w", err)
		}
		if !domain.IsValidCheckVerdict(verdict) {
			*findings = append(*findings, ValidateFinding{
				Link:   "DB->CHECK",
				Issue:  "invalid_value",
				Detail: fmt.Sprintf("check %s has invalid verdict %q", id, verdict),
			})
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate check verdicts: %w", err)
	}
	return nil
}

func validateStoryLinks(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`
		SELECT story.slug, story.depends_on
		FROM stories AS story
		LEFT JOIN stories AS dependency ON dependency.slug = story.depends_on
		WHERE story.depends_on IS NOT NULL AND dependency.slug IS NULL
		ORDER BY story.created_at, story.slug
	`)
	if err != nil {
		return fmt.Errorf("validate story links: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slug, dependency string
		if err := rows.Scan(&slug, &dependency); err != nil {
			return fmt.Errorf("validate story links: %w", err)
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "STORY->STORY",
			Issue:  "broken_link",
			Detail: fmt.Sprintf("story %q depends_on %q, but that story does not exist", slug, dependency),
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate story links: %w", err)
	}
	return nil
}

func validateRunLinks(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`
		SELECT runs.id, runs.story_slug
		FROM runs
		LEFT JOIN stories ON stories.slug = runs.story_slug
		WHERE stories.slug IS NULL
		ORDER BY runs.created_at, runs.id
	`)
	if err != nil {
		return fmt.Errorf("validate run links: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, storySlug string
		if err := rows.Scan(&id, &storySlug); err != nil {
			return fmt.Errorf("validate run links: %w", err)
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "STORY->RUN",
			Issue:  "broken_link",
			Detail: fmt.Sprintf("run %s references story %q, but that story does not exist", id, storySlug),
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate run links: %w", err)
	}
	return nil
}

func validateCheckLinks(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`
		SELECT checks.id, checks.run_id
		FROM checks
		LEFT JOIN runs ON runs.id = checks.run_id
		WHERE runs.id IS NULL
		ORDER BY checks.created_at, checks.id
	`)
	if err != nil {
		return fmt.Errorf("validate check links: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, runID string
		if err := rows.Scan(&id, &runID); err != nil {
			return fmt.Errorf("validate check links: %w", err)
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "RUN->CHECK",
			Issue:  "broken_link",
			Detail: fmt.Sprintf("check %s references run %s, but that run does not exist", id, runID),
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate check links: %w", err)
	}
	return nil
}

func validateHandoffLinks(db *sql.DB, findings *[]ValidateFinding) error {
	rows, err := db.Query(`
		SELECT handoffs.id, handoffs.run_id
		FROM handoffs
		LEFT JOIN runs ON runs.id = handoffs.run_id
		WHERE handoffs.run_id IS NOT NULL AND runs.id IS NULL
		ORDER BY handoffs.created_at, handoffs.id
	`)
	if err != nil {
		return fmt.Errorf("validate handoff run links: %w", err)
	}
	for rows.Next() {
		var id, runID string
		if err := rows.Scan(&id, &runID); err != nil {
			rows.Close()
			return fmt.Errorf("validate handoff run links: %w", err)
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "RUN->HANDOFF",
			Issue:  "broken_link",
			Detail: fmt.Sprintf("handoff %s references run %s, but that run does not exist", id, runID),
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("validate handoff run links: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("validate handoff run links: %w", err)
	}

	rows, err = db.Query(`
		SELECT handoffs.id, handoffs.check_id
		FROM handoffs
		LEFT JOIN checks ON checks.id = handoffs.check_id
		WHERE handoffs.check_id IS NOT NULL AND checks.id IS NULL
		ORDER BY handoffs.created_at, handoffs.id
	`)
	if err != nil {
		return fmt.Errorf("validate handoff check links: %w", err)
	}
	for rows.Next() {
		var id, checkID string
		if err := rows.Scan(&id, &checkID); err != nil {
			rows.Close()
			return fmt.Errorf("validate handoff check links: %w", err)
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "CHECK->HANDOFF",
			Issue:  "broken_link",
			Detail: fmt.Sprintf("handoff %s references check %s, but that check does not exist", id, checkID),
		})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("validate handoff check links: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("validate handoff check links: %w", err)
	}

	rows, err = db.Query(`
		SELECT handoffs.id, handoffs.run_id, handoffs.check_id, checks.run_id
		FROM handoffs
		JOIN runs ON runs.id = handoffs.run_id
		JOIN checks ON checks.id = handoffs.check_id
		WHERE checks.run_id != handoffs.run_id
		ORDER BY handoffs.created_at, handoffs.id
	`)
	if err != nil {
		return fmt.Errorf("validate handoff lifecycle: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, handoffRunID, checkID, checkRunID string
		if err := rows.Scan(&id, &handoffRunID, &checkID, &checkRunID); err != nil {
			return fmt.Errorf("validate handoff lifecycle: %w", err)
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "CHECK->HANDOFF",
			Issue:  "broken_link",
			Detail: fmt.Sprintf("handoff %s anchors run %s, but check %s belongs to run %s", id, handoffRunID, checkID, checkRunID),
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("validate handoff lifecycle: %w", err)
	}
	return nil
}

func validateMetaLinks(db *sql.DB, findings *[]ValidateFinding) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM meta`).Scan(&count); err != nil {
		return fmt.Errorf("validate meta: %w", err)
	}
	if count != 1 {
		issue := "broken_link"
		if count == 0 {
			issue = "missing_key"
		}
		*findings = append(*findings, ValidateFinding{
			Link:   "DB->META",
			Issue:  issue,
			Detail: fmt.Sprintf("meta contains %d rows; expected exactly one", count),
		})
		if count == 0 {
			return nil
		}
	}

	links := []struct {
		column string
		table  string
		key    string
		link   string
	}{
		{column: "current_phase", table: "stories", key: "slug", link: "META->STORY"},
		{column: "entry_phase", table: "stories", key: "slug", link: "META->STORY"},
		{column: "latest_run_id", table: "runs", key: "id", link: "META->RUN"},
		{column: "latest_check_id", table: "checks", key: "id", link: "META->CHECK"},
	}
	for _, link := range links {
		query := fmt.Sprintf(`
			SELECT meta.%s
			FROM meta
			LEFT JOIN %s ON %s.%s = meta.%s
			WHERE meta.%s IS NOT NULL AND %s.%s IS NULL
			ORDER BY meta.%s
		`, link.column, link.table, link.table, link.key, link.column, link.column, link.table, link.key, link.column)
		rows, err := db.Query(query)
		if err != nil {
			return fmt.Errorf("validate meta %s: %w", link.column, err)
		}
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				rows.Close()
				return fmt.Errorf("validate meta %s: %w", link.column, err)
			}
			*findings = append(*findings, ValidateFinding{
				Link:   link.link,
				Issue:  "stale_pointer",
				Detail: fmt.Sprintf("meta.%s references %q, but that row does not exist", link.column, value),
			})
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return fmt.Errorf("validate meta %s: %w", link.column, err)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("validate meta %s: %w", link.column, err)
		}
	}
	return nil
}

func looksLikeULID(s string) bool {
	_, err := ulid.ParseStrict(s)
	return err == nil
}

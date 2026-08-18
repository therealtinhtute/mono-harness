package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// CreateStory validates and records a new story (phase) entity
// (CONTRACT.md `story`). duplicate_slug and unknown_dependency are both
// DB-lookup-dependent, so they're enforced here rather than in
// domain.Story.Validate() (which stays dependency-free per the domain
// layering rule).
func CreateStory(db *sql.DB, slug, goal, dependsOn string) (id string, err error) {
	entity := domain.Story{Slug: slug, Goal: goal, Status: domain.StoryPlanned}
	if err := entity.Validate(); err != nil {
		return "", err
	}

	_, _, exists, err := storyByExactSlug(db, slug)
	if err != nil {
		return "", err
	}
	if exists {
		return "", &domain.ValidationError{Code: "duplicate_slug", Message: "story: slug " + slug + " already exists"}
	}

	if dependsOn != "" {
		_, _, depExists, err := storyByExactSlug(db, dependsOn)
		if err != nil {
			return "", err
		}
		if !depExists {
			return "", &domain.ValidationError{Code: "unknown_dependency", Message: "story: depends-on slug " + dependsOn + " not found"}
		}
	}

	// Markdown is the write target for the story's phase-block status the
	// same way trace/decision/check/handoff already are (R8, P3 wave 1,
	// docs/plans/active/harness-markdown-truth.md): a slug with no matching
	// phase block (an ad hoc story, not part of `## Phases and
	// Verification`) is not an error — writePlan is then a no-op.
	writePlan, err := preparePlanPhaseStatus(db, slug, domain.StoryPlanned)
	if err != nil {
		return "", err
	}
	if err := writePlan(); err != nil {
		return "", fmt.Errorf("plan write failed: %w", err)
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	var dependsOnArg any
	if dependsOn != "" {
		dependsOnArg = dependsOn
	}
	if _, err := db.Exec(
		`INSERT INTO stories (id, slug, goal, status, depends_on, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, slug, goal, domain.StoryPlanned, dependsOnArg, at,
	); err != nil {
		return "", fmt.Errorf("insert story: %w", err)
	}
	return id, nil
}

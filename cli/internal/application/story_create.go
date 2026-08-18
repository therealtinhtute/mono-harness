package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// CreateStory validates and records a new story (phase) entity
// (CONTRACT.md `story`), changeset-first. duplicate_slug and
// unknown_dependency are both DB-lookup-dependent, so they're enforced
// here rather than in domain.Story.Validate() (which stays dependency-free
// per the domain layering rule).
func CreateStory(db *sql.DB, changesetDir, slug, goal, dependsOn string) (id, path string, err error) {
	entity := domain.Story{Slug: slug, Goal: goal, Status: domain.StoryPlanned}
	if err := entity.Validate(); err != nil {
		return "", "", err
	}

	_, _, exists, err := storyByExactSlug(db, slug)
	if err != nil {
		return "", "", err
	}
	if exists {
		return "", "", &domain.ValidationError{Code: "duplicate_slug", Message: "story: slug " + slug + " already exists"}
	}

	if dependsOn != "" {
		_, _, depExists, err := storyByExactSlug(db, dependsOn)
		if err != nil {
			return "", "", err
		}
		if !depExists {
			return "", "", &domain.ValidationError{Code: "unknown_dependency", Message: "story: depends-on slug " + dependsOn + " not found"}
		}
	}

	// Markdown is the write target for the story's phase-block status the
	// same way trace/decision/check/handoff already are (R8, P3 wave 1,
	// docs/plans/active/harness-markdown-truth.md): a slug with no matching
	// phase block (an ad hoc story, not part of `## Phases and
	// Verification`) is not an error — writePlan is then a no-op.
	writePlan, err := preparePlanPhaseStatus(db, slug, domain.StoryPlanned)
	if err != nil {
		return "", "", err
	}
	if err := writePlan(); err != nil {
		return "", "", fmt.Errorf("plan write failed: %w", err)
	}

	at := time.Now().UTC().Format(time.RFC3339)
	id = ulid.Make().String()
	fields := map[string]any{
		"slug":       slug,
		"goal":       goal,
		"status":     domain.StoryPlanned,
		"created_at": at,
	}
	if dependsOn != "" {
		fields["depends_on"] = dependsOn
	}
	path, _, err = AppendAndApply(db, changesetDir, []infrastructure.ChangesetLine{
		{Op: "create", Entity: "story", ID: id, Fields: fields, At: at},
	})
	if err != nil {
		return "", "", err
	}
	return id, path, nil
}

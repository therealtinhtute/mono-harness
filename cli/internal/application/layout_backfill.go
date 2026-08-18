package application

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

type legacyStory struct {
	ID        string
	Slug      string
	Goal      string
	Status    string
	DependsOn sql.NullString
	CreatedAt string
}

func layoutBackfillLines(db *sql.DB) ([]infrastructure.ChangesetLine, error) {
	at := time.Now().UTC().Format(time.RFC3339)
	var lines []infrastructure.ChangesetLine

	intakeQuery := `SELECT id, type, summary, lane, created_at FROM intakes ORDER BY created_at, id`
	intakeFields := []string{"type", "summary", "lane", "created_at"}
	hasPlanPath, err := tableHasColumn(db, "intakes", "plan_path")
	if err != nil {
		return nil, err
	}
	if hasPlanPath {
		intakeQuery = `SELECT id, type, summary, lane, plan_path, created_at FROM intakes ORDER BY created_at, id`
		intakeFields = []string{"type", "summary", "lane", "plan_path", "created_at"}
	}
	intakeLines, err := queryBackfillRows(db, intakeQuery, "intake", at, intakeFields)
	if err != nil {
		return nil, err
	}
	lines = append(lines, intakeLines...)

	stories, err := queryLegacyStories(db)
	if err != nil {
		return nil, err
	}
	orderedStories, err := orderLegacyStories(stories)
	if err != nil {
		return nil, err
	}
	for _, story := range orderedStories {
		fields := map[string]any{
			"slug": story.Slug, "goal": story.Goal, "status": story.Status, "created_at": story.CreatedAt,
		}
		if story.DependsOn.Valid {
			fields["depends_on"] = story.DependsOn.String
		}
		lines = appendCreateAndUpdate(lines, "story", story.ID, fields, at)
	}

	for _, table := range []struct {
		query  string
		entity string
		fields []string
	}{
		{`SELECT id, story_slug, plan_id, trace_ids, artifact_path, created_at FROM runs ORDER BY created_at, id`, "run", []string{"story_slug", "plan_id", "trace_ids", "artifact_path", "created_at"}},
		{`SELECT id, run_id, verdict, proof_links, artifact_path, created_at FROM checks ORDER BY created_at, id`, "check", []string{"run_id", "verdict", "proof_links", "artifact_path", "created_at"}},
		{`SELECT id, run_id, check_id, anchors, created_at FROM handoffs ORDER BY created_at, id`, "handoff", []string{"run_id", "check_id", "anchors", "created_at"}},
		{`SELECT id, run_id, wave, summary, created_at FROM traces ORDER BY created_at, id`, "trace", []string{"run_id", "wave", "summary", "created_at"}},
	} {
		rowLines, err := queryBackfillRows(db, table.query, table.entity, at, table.fields)
		if err != nil {
			return nil, err
		}
		lines = append(lines, rowLines...)
	}

	var currentPhase, entryPhase, latestRunID, latestCheckID, docsVersion sql.NullString
	if err := db.QueryRow(`SELECT current_phase, entry_phase, latest_run_id, latest_check_id, docs_version FROM meta LIMIT 1`).Scan(&currentPhase, &entryPhase, &latestRunID, &latestCheckID, &docsVersion); err != nil {
		return nil, fmt.Errorf("read legacy meta for backfill: %w", err)
	}
	meta := map[string]any{}
	addNullableString(meta, "current_phase", currentPhase)
	addNullableString(meta, "entry_phase", entryPhase)
	addNullableString(meta, "latest_run_id", latestRunID)
	addNullableString(meta, "latest_check_id", latestCheckID)
	addNullableString(meta, "docs_version", docsVersion)
	lines = append(lines, infrastructure.ChangesetLine{Op: "update", Entity: "meta", ID: "meta", Fields: meta, At: at})
	return lines, nil
}

func queryBackfillRows(db *sql.DB, query, entity, at string, fieldNames []string) ([]infrastructure.ChangesetLine, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query legacy %s rows: %w", entity, err)
	}
	defer rows.Close()

	var lines []infrastructure.ChangesetLine
	for rows.Next() {
		var id string
		values := make([]any, len(fieldNames))
		dest := make([]any, 0, len(fieldNames)+1)
		dest = append(dest, &id)
		for i := range values {
			dest = append(dest, &values[i])
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scan legacy %s row: %w", entity, err)
		}
		fields := make(map[string]any, len(fieldNames))
		for i, name := range fieldNames {
			fields[name] = normalizeSQLValue(values[i])
		}
		lines = appendCreateAndUpdate(lines, entity, id, fields, at)
	}
	return lines, rows.Err()
}

func queryLegacyStories(db *sql.DB) ([]legacyStory, error) {
	rows, err := db.Query(`SELECT id, slug, goal, status, depends_on, created_at FROM stories ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("query legacy stories: %w", err)
	}
	defer rows.Close()

	var stories []legacyStory
	for rows.Next() {
		var story legacyStory
		if err := rows.Scan(&story.ID, &story.Slug, &story.Goal, &story.Status, &story.DependsOn, &story.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan legacy story: %w", err)
		}
		stories = append(stories, story)
	}
	return stories, rows.Err()
}

func orderLegacyStories(stories []legacyStory) ([]legacyStory, error) {
	bySlug := make(map[string]legacyStory, len(stories))
	for _, story := range stories {
		bySlug[story.Slug] = story
	}
	state := map[string]int{}
	ordered := make([]legacyStory, 0, len(stories))
	var visit func(string) error
	visit = func(slug string) error {
		switch state[slug] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("legacy story dependency cycle at %s", slug)
		}
		story, ok := bySlug[slug]
		if !ok {
			return nil
		}
		state[slug] = 1
		if story.DependsOn.Valid {
			if _, exists := bySlug[story.DependsOn.String]; !exists {
				return fmt.Errorf("legacy story %s depends on missing story %s", slug, story.DependsOn.String)
			}
			if err := visit(story.DependsOn.String); err != nil {
				return err
			}
		}
		state[slug] = 2
		ordered = append(ordered, story)
		return nil
	}
	for _, story := range stories {
		if err := visit(story.Slug); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func appendCreateAndUpdate(lines []infrastructure.ChangesetLine, entity, id string, fields map[string]any, at string) []infrastructure.ChangesetLine {
	lines = append(lines, infrastructure.ChangesetLine{Op: "create", Entity: entity, ID: id, Fields: fields, At: at})
	lines = append(lines, infrastructure.ChangesetLine{Op: "update", Entity: entity, ID: id, Fields: fields, At: at})
	return lines
}

func normalizeSQLValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return string(typed)
	default:
		return typed
	}
}

func tableHasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func addNullableString(fields map[string]any, key string, value sql.NullString) {
	if value.Valid {
		fields[key] = value.String
	} else {
		fields[key] = nil
	}
}

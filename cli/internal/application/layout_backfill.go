package application

import (
	"database/sql"
	"fmt"
	"strings"
)

type legacyStory struct {
	ID        string
	Slug      string
	Goal      string
	Status    string
	DependsOn sql.NullString
	CreatedAt string
}

// copyLifecycleRows copies every lifecycle row directly from legacyDB into
// tempDB (freshly migrated, empty). It replaces the old changeset-replay
// path: the legacy database's own rows are the source of truth now, not a
// changeset log (P3 wave 2, docs/plans/active/harness-markdown-truth.md).
func copyLifecycleRows(legacyDB, tempDB *sql.DB) (copied int, err error) {
	intakeQuery := `SELECT id, type, summary, lane, created_at FROM intakes ORDER BY created_at, id`
	intakeCols := []string{"id", "type", "summary", "lane", "created_at"}
	hasPlanPath, err := tableHasColumn(legacyDB, "intakes", "plan_path")
	if err != nil {
		return 0, err
	}
	if hasPlanPath {
		intakeQuery = `SELECT id, type, summary, lane, plan_path, created_at FROM intakes ORDER BY created_at, id`
		intakeCols = []string{"id", "type", "summary", "lane", "plan_path", "created_at"}
	}
	n, err := copyRows(legacyDB, tempDB, intakeQuery, "intakes", intakeCols)
	if err != nil {
		return copied, err
	}
	copied += n

	stories, err := queryLegacyStories(legacyDB)
	if err != nil {
		return copied, err
	}
	ordered, err := orderLegacyStories(stories)
	if err != nil {
		return copied, err
	}
	for _, story := range ordered {
		var dependsOn any
		if story.DependsOn.Valid {
			dependsOn = story.DependsOn.String
		}
		if _, err := tempDB.Exec(
			`INSERT INTO stories (id, slug, goal, status, depends_on, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			story.ID, story.Slug, story.Goal, story.Status, dependsOn, story.CreatedAt,
		); err != nil {
			return copied, fmt.Errorf("copy story %s: %w", story.ID, err)
		}
		copied++
	}

	for _, table := range []struct {
		query string
		table string
		cols  []string
	}{
		{`SELECT id, story_slug, plan_id, trace_ids, artifact_path, created_at FROM runs ORDER BY created_at, id`, "runs", []string{"id", "story_slug", "plan_id", "trace_ids", "artifact_path", "created_at"}},
		{`SELECT id, run_id, verdict, judge, judge_model, proof_links, artifact_path, created_at FROM checks ORDER BY created_at, id`, "checks", []string{"id", "run_id", "verdict", "judge", "judge_model", "proof_links", "artifact_path", "created_at"}},
		{`SELECT id, run_id, check_id, anchors, created_at FROM handoffs ORDER BY created_at, id`, "handoffs", []string{"id", "run_id", "check_id", "anchors", "created_at"}},
		{`SELECT id, run_id, wave, summary, task, task_status, created_at FROM traces ORDER BY created_at, id`, "traces", []string{"id", "run_id", "wave", "summary", "task", "task_status", "created_at"}},
		{`SELECT id, run_id, phase, task, decision, rationale, created_at FROM decisions ORDER BY created_at, id`, "decisions", []string{"id", "run_id", "phase", "task", "decision", "rationale", "created_at"}},
	} {
		n, err := copyRows(legacyDB, tempDB, table.query, table.table, table.cols)
		if err != nil {
			return copied, err
		}
		copied += n
	}

	var currentPhase, entryPhase, latestRunID, latestCheckID, docsVersion sql.NullString
	if err := legacyDB.QueryRow(`SELECT current_phase, entry_phase, latest_run_id, latest_check_id, docs_version FROM meta LIMIT 1`).Scan(&currentPhase, &entryPhase, &latestRunID, &latestCheckID, &docsVersion); err != nil {
		return copied, fmt.Errorf("read legacy meta: %w", err)
	}
	if _, err := tempDB.Exec(
		`UPDATE meta SET current_phase = ?, entry_phase = ?, latest_run_id = ?, latest_check_id = ?, docs_version = ?`,
		currentPhase, entryPhase, latestRunID, latestCheckID, docsVersion,
	); err != nil {
		return copied, fmt.Errorf("copy meta: %w", err)
	}
	return copied, nil
}

// copyRows reads every row of query from legacyDB and inserts it into
// tempDB's table verbatim, in cols order (cols[0] must be "id").
func copyRows(legacyDB, tempDB *sql.DB, query, table string, cols []string) (int, error) {
	rows, err := legacyDB.Query(query)
	if err != nil {
		return 0, fmt.Errorf("read legacy %s: %w", table, err)
	}
	defer rows.Close()

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",")
	insert := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), placeholders) //nolint:gosec // table/cols are literal call-site constants, not user input

	count := 0
	for rows.Next() {
		values := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return count, fmt.Errorf("scan legacy %s row: %w", table, err)
		}
		args := make([]any, len(values))
		for i, v := range values {
			args[i] = normalizeSQLValue(v)
		}
		if _, err := tempDB.Exec(insert, args...); err != nil {
			return count, fmt.Errorf("copy %s row: %w", table, err)
		}
		count++
	}
	return count, rows.Err()
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

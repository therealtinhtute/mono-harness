package application

import (
	"database/sql"
	"fmt"
)

func storyForRun(db *sql.DB, runID string) (storyID, storySlug string, exists bool, err error) {
	err = db.QueryRow(`
		SELECT stories.id, stories.slug
		FROM runs
		JOIN stories ON stories.slug = runs.story_slug
		WHERE runs.id = ?
	`, runID).Scan(&storyID, &storySlug)
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("query run story %q: %w", runID, err)
	}
	return storyID, storySlug, true, nil
}

func checkForPhaseClose(db *sql.DB, checkID string) (runID, verdict string, exists bool, err error) {
	err = db.QueryRow(`SELECT run_id, verdict FROM checks WHERE id = ?`, checkID).Scan(&runID, &verdict)
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("query check %q: %w", checkID, err)
	}
	return runID, verdict, true, nil
}

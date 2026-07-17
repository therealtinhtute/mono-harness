package application

import (
	"database/sql"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// AppendAndApply writes a changeset file to changesetDir, then applies it
// to db. Per cli-core's locked write-order decision, the write happens
// first: if apply then fails, the changeset is still on disk and a later
// `db changeset apply`/replay heals the db from it.
func AppendAndApply(db *sql.DB, changesetDir string, lines []infrastructure.ChangesetLine) (path string, applied int, err error) {
	path, err = infrastructure.WriteChangeset(changesetDir, lines)
	if err != nil {
		return "", 0, err
	}

	applied, _, err = infrastructure.ApplyChangeset(db, path)
	return path, applied, err
}

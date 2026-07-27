package application

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// AppendAndApply writes a changeset file to changesetDir, then applies it
// to db. Per cli-core's locked write-order decision, the write happens
// first: if apply then fails, the changeset is still on disk and a later
// `db changeset apply`/replay heals the db from it.
func AppendAndApply(db *sql.DB, changesetDir string, lines []infrastructure.ChangesetLine) (path string, applied int, err error) {
	id, err := prepareChangesetAppend(db, changesetDir)
	if err != nil {
		return "", 0, err
	}
	return writeAndApplyPreparedChangeset(db, changesetDir, id, lines)
}

// AppendNewEntityAndApply uses the ordered changeset ULID as the new entity's
// ID. Callers derive timestamps from that ID so latest selection stays aligned
// with replay order through same-time generation and clock rollback.
func AppendNewEntityAndApply(db *sql.DB, changesetDir string, build func(id string) []infrastructure.ChangesetLine) (id, path string, applied int, err error) {
	id, err = prepareChangesetAppend(db, changesetDir)
	if err != nil {
		return "", "", 0, err
	}
	path, applied, err = writeAndApplyPreparedChangeset(db, changesetDir, id, build(id))
	return id, path, applied, err
}

func orderedChangesetTime(id string) string {
	return time.UnixMilli(int64(ulid.MustParse(id).Time())).UTC().Format("2006-01-02T15:04:05.000Z07:00")
}

func ApplyChangesetForRecovery(db *sql.DB, changesetDir, path string) (applied, skipped int, err error) {
	pending, _, err := infrastructure.PendingCanonicalChangesets(db, changesetDir)
	if err != nil {
		return 0, 0, err
	}
	if len(pending) == 0 {
		return infrastructure.ApplyChangeset(db, path)
	}

	trackedPath := filepath.Join(changesetDir, pending[0])
	requestedInfo, err := os.Stat(path)
	if err != nil {
		return 0, 0, recoveryRequired(pending[0])
	}
	trackedInfo, err := os.Stat(trackedPath)
	if err != nil {
		return 0, 0, err
	}
	if !os.SameFile(requestedInfo, trackedInfo) {
		return 0, 0, recoveryRequired(pending[0])
	}
	return infrastructure.ApplyChangeset(db, trackedPath)
}

func prepareChangesetAppend(db *sql.DB, changesetDir string) (string, error) {
	pending, fence, err := infrastructure.PendingCanonicalChangesets(db, changesetDir)
	if err != nil {
		return "", err
	}
	if len(pending) > 0 {
		return "", recoveryRequired(pending[0])
	}
	return infrastructure.NextChangesetID(changesetDir, fence)
}

func writeAndApplyPreparedChangeset(db *sql.DB, changesetDir, id string, lines []infrastructure.ChangesetLine) (path string, applied int, err error) {
	path, err = infrastructure.WriteChangesetWithID(changesetDir, id, lines)
	if err != nil {
		return "", 0, err
	}
	applied, _, err = infrastructure.ApplyChangeset(db, path)
	return path, applied, err
}

func recoveryRequired(earliest string) *domain.ValidationError {
	return &domain.ValidationError{
		Code:    "changeset_recovery_required",
		Message: "changeset recovery required: apply " + earliest + " first with `zharness db changeset apply " + filepath.ToSlash(filepath.Join(".kit", "changesets", earliest)) + "`",
	}
}

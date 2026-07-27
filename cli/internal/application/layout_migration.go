package application

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

type LayoutMigrationResult struct {
	Status        string `json:"status"`
	SourceDB      string `json:"source_db"`
	TargetDB      string `json:"target_db"`
	Replayed      int    `json:"replayed"`
	Backfilled    int    `json:"backfilled"`
	Parity        bool   `json:"parity"`
	DocsWritten   bool   `json:"docs_written"`
	DryRun        bool   `json:"dry_run"`
	SchemaVersion int    `json:"schema_version"`
}

type layoutSnapshot struct {
	State         StateView
	Phases        []PhaseView
	Resume        ResumeView
	LatestCheck   CheckView
	HasCheck      bool
	LifecycleRows map[string][][]any
}

type fileSnapshot struct {
	Path   string
	Exists bool
	Data   []byte
	Mode   fs.FileMode
}

func MigrateLayout(root, legacyPath, targetPath, changesetDir, kitDir string, docsFS fs.FS, version string, dryRun bool) (LayoutMigrationResult, error) {
	legacyPath = rootedPath(root, legacyPath)
	targetPath = rootedPath(root, targetPath)
	changesetDir = rootedPath(root, changesetDir)

	result := LayoutMigrationResult{
		SourceDB: filepath.ToSlash(legacyPath),
		TargetDB: filepath.ToSlash(targetPath),
		DryRun:   dryRun,
	}
	legacyExists := infrastructure.Exists(legacyPath)
	targetExists := infrastructure.Exists(targetPath)
	switch {
	case targetExists && !legacyExists:
		result.Status = "already-v2"
		result.Parity = true
		return result, nil
	case targetExists && legacyExists:
		return result, fmt.Errorf("layout migration: both %s and %s exist", legacyPath, targetPath)
	case !legacyExists:
		return result, fmt.Errorf("layout migration: legacy database not found at %s", legacyPath)
	}

	legacyDB, err := infrastructure.OpenReadOnlyUnderExistingLock(legacyPath)
	if err != nil {
		return result, err
	}
	before, err := captureLayoutSnapshot(legacyDB.Raw(), version)
	if err != nil {
		legacyDB.Close()
		return result, err
	}

	tempFile, err := os.CreateTemp(root, ".harness.db.migrate-*")
	if err != nil {
		legacyDB.Close()
		return result, fmt.Errorf("create migration db: %w", err)
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		legacyDB.Close()
		return result, err
	}
	if err := os.Remove(tempPath); err != nil {
		legacyDB.Close()
		return result, err
	}
	defer cleanupSQLiteFiles(tempPath)

	tempDB, err := infrastructure.Open(tempPath)
	if err != nil {
		legacyDB.Close()
		return result, err
	}
	_, schemaVersion, err := infrastructure.Migrate(tempDB)
	if err != nil {
		tempDB.Close()
		legacyDB.Close()
		return result, err
	}
	result.SchemaVersion = schemaVersion
	result.Replayed, err = infrastructure.Replay(tempDB, changesetDir)
	if err != nil {
		tempDB.Close()
		legacyDB.Close()
		return result, fmt.Errorf("layout migration replay: %w", err)
	}

	stageParent := ""
	if !dryRun {
		stageParent = rootedPath(root, filepath.Join(kitDir, "cache"))
		if err := os.MkdirAll(stageParent, 0o755); err != nil {
			tempDB.Close()
			legacyDB.Close()
			return result, fmt.Errorf("create staged changeset parent: %w", err)
		}
	}
	stageChangesets, err := os.MkdirTemp(stageParent, "zharness-layout-changesets-")
	if err != nil {
		tempDB.Close()
		legacyDB.Close()
		return result, fmt.Errorf("create staged changeset dir: %w", err)
	}
	defer os.RemoveAll(stageChangesets)

	after, err := captureLayoutSnapshot(tempDB, version)
	if err != nil {
		tempDB.Close()
		legacyDB.Close()
		return result, err
	}
	before.State.SchemaVersion = 0
	after.State.SchemaVersion = 0
	parityErr := compareLayoutSnapshots(before, after)
	if parityErr != nil {
		backfill, backfillErr := layoutBackfillLines(legacyDB.Raw())
		if backfillErr != nil {
			tempDB.Close()
			legacyDB.Close()
			return result, backfillErr
		}
		_, result.Backfilled, backfillErr = AppendAndApply(tempDB, stageChangesets, backfill)
		if backfillErr != nil {
			tempDB.Close()
			legacyDB.Close()
			return result, fmt.Errorf("layout migration backfill: %w", backfillErr)
		}
		after, backfillErr = captureLayoutSnapshot(tempDB, version)
		if backfillErr != nil {
			tempDB.Close()
			legacyDB.Close()
			return result, backfillErr
		}
		after.State.SchemaVersion = 0
		parityErr = compareLayoutSnapshots(before, after)
	}
	result.Parity = parityErr == nil
	if parityErr != nil {
		tempDB.Close()
		legacyDB.Close()
		return result, fmt.Errorf("layout migration: replayed state does not match legacy state after semantic backfill: %w", parityErr)
	}
	if dryRun {
		result.Status = "dry-run"
		tempDB.Close()
		legacyDB.Close()
		return result, nil
	}

	snapshots, err := snapshotManagedTargets(root, docsFS)
	if err != nil {
		tempDB.Close()
		legacyDB.Close()
		return result, err
	}
	scaffold, err := ScaffoldDocs(tempDB, stageChangesets, root, kitDir, docsFS, version, true, false)
	if err != nil {
		tempDB.Close()
		legacyDB.Close()
		_ = restoreFileSnapshots(snapshots)
		return result, err
	}
	result.DocsWritten = scaffold.DocsWritten
	if _, err := tempDB.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		tempDB.Close()
		legacyDB.Close()
		_ = restoreFileSnapshots(snapshots)
		return result, fmt.Errorf("checkpoint migration db: %w", err)
	}
	if err := tempDB.Close(); err != nil {
		legacyDB.Close()
		_ = restoreFileSnapshots(snapshots)
		return result, err
	}
	if err := legacyDB.Close(); err != nil {
		_ = restoreFileSnapshots(snapshots)
		return result, err
	}

	movedChangesets, err := activateStagedChangesets(stageChangesets, changesetDir)
	if err != nil {
		_ = restoreFileSnapshots(snapshots)
		return result, err
	}
	rollback := func() {
		_ = rollbackActivatedChangesets(movedChangesets, stageChangesets)
		_ = restoreFileSnapshots(snapshots)
	}

	legacyWritable, err := infrastructure.Open(legacyPath)
	if err != nil {
		rollback()
		return result, err
	}
	if _, err := legacyWritable.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		legacyWritable.Close()
		rollback()
		return result, fmt.Errorf("checkpoint legacy db: %w", err)
	}
	if err := legacyWritable.Close(); err != nil {
		rollback()
		return result, err
	}

	backupPath := legacyPath + ".layout-v1-backup"
	if err := os.Rename(legacyPath, backupPath); err != nil {
		rollback()
		return result, fmt.Errorf("backup legacy db: %w", err)
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Rename(backupPath, legacyPath)
		rollback()
		return result, fmt.Errorf("activate root db: %w", err)
	}
	if err := os.Remove(backupPath); err != nil {
		cleanupSQLiteFiles(targetPath)
		_ = os.Rename(backupPath, legacyPath)
		rollback()
		return result, fmt.Errorf("remove legacy db backup: %w", err)
	}
	cleanupSQLiteFiles(legacyPath)
	result.Status = "migrated"
	return result, nil
}

func captureLayoutSnapshot(db *sql.DB, version string) (layoutSnapshot, error) {
	state, err := QueryState(db)
	if err != nil {
		return layoutSnapshot{}, err
	}
	phases, err := QueryPhases(db)
	if err != nil {
		return layoutSnapshot{}, err
	}
	resume, err := Resume(db, version)
	if err != nil {
		return layoutSnapshot{}, err
	}
	check, hasCheck, err := QueryLatestCheck(db)
	if err != nil {
		return layoutSnapshot{}, err
	}
	lifecycleRows, err := captureLifecycleRows(db)
	if err != nil {
		return layoutSnapshot{}, err
	}
	return layoutSnapshot{State: state, Phases: phases, Resume: resume, LatestCheck: check, HasCheck: hasCheck, LifecycleRows: lifecycleRows}, nil
}

func captureLifecycleRows(db *sql.DB) (map[string][][]any, error) {
	queries := map[string]string{
		"intakes":       `SELECT id, type, summary, lane, created_at FROM intakes ORDER BY id`,
		"stories":       `SELECT id, slug, goal, status, depends_on, created_at FROM stories ORDER BY id`,
		"runs":          `SELECT id, story_slug, plan_id, trace_ids, artifact_path, created_at FROM runs ORDER BY id`,
		"checks":        `SELECT id, run_id, verdict, proof_links, artifact_path, created_at FROM checks ORDER BY id`,
		"handoffs":      `SELECT id, run_id, check_id, anchors, created_at FROM handoffs ORDER BY id`,
		"interventions": `SELECT id, verdict_id, reason, created_at FROM interventions ORDER BY id`,
		"traces":        `SELECT id, run_id, wave, summary, created_at FROM traces ORDER BY id`,
		"meta":          `SELECT current_phase, entry_phase, latest_run_id, latest_check_id, docs_version FROM meta`,
	}
	captured := make(map[string][][]any, len(queries))
	for name, query := range queries {
		rows, err := db.Query(query)
		if err != nil {
			return nil, fmt.Errorf("capture %s rows: %w", name, err)
		}
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			return nil, err
		}
		for rows.Next() {
			values := make([]any, len(columns))
			dest := make([]any, len(columns))
			for i := range values {
				dest[i] = &values[i]
			}
			if err := rows.Scan(dest...); err != nil {
				rows.Close()
				return nil, err
			}
			for i := range values {
				values[i] = normalizeSQLValue(values[i])
			}
			captured[name] = append(captured[name], values)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return captured, nil
}

func compareLayoutSnapshots(before, after layoutSnapshot) error {
	if !reflect.DeepEqual(before.State, after.State) {
		return fmt.Errorf("state differs: legacy=%+v replay=%+v", before.State, after.State)
	}
	if !reflect.DeepEqual(before.Phases, after.Phases) {
		return fmt.Errorf("phases differ: legacy=%d replay=%d", len(before.Phases), len(after.Phases))
	}
	if !reflect.DeepEqual(before.Resume, after.Resume) {
		return fmt.Errorf("resume differs: legacy=%+v replay=%+v", before.Resume, after.Resume)
	}
	if before.HasCheck != after.HasCheck || !reflect.DeepEqual(before.LatestCheck, after.LatestCheck) {
		return fmt.Errorf("latest check differs: legacy=%+v/%v replay=%+v/%v", before.LatestCheck, before.HasCheck, after.LatestCheck, after.HasCheck)
	}
	if !reflect.DeepEqual(before.LifecycleRows, after.LifecycleRows) {
		for table, legacyRows := range before.LifecycleRows {
			if !reflect.DeepEqual(legacyRows, after.LifecycleRows[table]) {
				return fmt.Errorf("%s rows differ: legacy=%d replay=%d", table, len(legacyRows), len(after.LifecycleRows[table]))
			}
		}
		return fmt.Errorf("lifecycle rows differ")
	}
	return nil
}

func rootedPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func snapshotManagedTargets(root string, docsFS fs.FS) ([]fileSnapshot, error) {
	paths := []string{filepath.Join(root, "AGENTS.md"), filepath.Join(root, ".gitignore")}
	err := fs.WalkDir(docsFS, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && path != "AGENTS.md" {
			paths = append(paths, filepath.Join(root, "docs", filepath.FromSlash(path)))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	snapshots := make([]fileSnapshot, 0, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			snapshots = append(snapshots, fileSnapshot{Path: path})
			continue
		}
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, fileSnapshot{Path: path, Exists: true, Data: data, Mode: info.Mode()})
	}
	return snapshots, nil
}

func restoreFileSnapshots(snapshots []fileSnapshot) error {
	for _, snapshot := range snapshots {
		if !snapshot.Exists {
			if err := os.Remove(snapshot.Path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(snapshot.Path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(snapshot.Path, snapshot.Data, snapshot.Mode.Perm()); err != nil {
			return err
		}
	}
	return nil
}

type movedChangeset struct {
	From string
	To   string
}

func activateStagedChangesets(stageDir, targetDir string) ([]movedChangeset, error) {
	names, err := infrastructure.ListChangesets(stageDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, err
	}
	moved := make([]movedChangeset, 0, len(names))
	for _, name := range names {
		from := filepath.Join(stageDir, name)
		to := filepath.Join(targetDir, name)
		if infrastructure.Exists(to) {
			_ = rollbackActivatedChangesets(moved, stageDir)
			return nil, fmt.Errorf("activate changeset: %s already exists", to)
		}
		if err := os.Rename(from, to); err != nil {
			_ = rollbackActivatedChangesets(moved, stageDir)
			return nil, err
		}
		moved = append(moved, movedChangeset{From: from, To: to})
	}
	return moved, nil
}

func rollbackActivatedChangesets(moved []movedChangeset, stageDir string) error {
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return err
	}
	for i := len(moved) - 1; i >= 0; i-- {
		if infrastructure.Exists(moved[i].To) {
			if err := os.Rename(moved[i].To, moved[i].From); err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanupSQLiteFiles(path string) {
	_ = os.Remove(path)
	_ = os.Remove(path + "-wal")
	_ = os.Remove(path + "-shm")
}

package infrastructure

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/domain"
)

// ChangesetLine is one line of a `{ulid}.changeset.jsonl` file
// (SCHEMA.md Changeset Format).
type ChangesetLine struct {
	Op     string         `json:"op"`
	Entity string         `json:"entity"`
	ID     string         `json:"id"`
	Fields map[string]any `json:"fields"`
	At     string         `json:"at"`
}

// entityTables maps a changeset `entity` string to its SQL table name
// (SCHEMA.md's Table ↔ Changeset Entity Type).
var entityTables = map[string]string{
	"story":        "stories",
	"run":          "runs",
	"check":        "checks",
	"handoff":      "handoffs",
	"intake":       "intakes",
	"intervention": "interventions",
	"trace":        "traces",
	"managed_doc":  "managed_docs",
}

// entityColumns allowlists the writable, non-id columns per table (SCHEMA.md
// column lists, `migrations.go`'s CREATE TABLE statements). Changeset
// `fields` keys become raw SQL identifiers in applyCreate/applyUpdate, so
// every key must be checked against this set before it reaches query text —
// a changeset is untrusted external input (`db changeset apply <path>`
// accepts any file on disk), not something the CLI itself always produced.
var entityColumns = map[string]map[string]bool{
	"stories":       {"slug": true, "goal": true, "status": true, "depends_on": true, "created_at": true},
	"runs":          {"story_slug": true, "plan_id": true, "trace_ids": true, "artifact_path": true, "created_at": true},
	"checks":        {"run_id": true, "verdict": true, "judge": true, "judge_model": true, "proof_links": true, "artifact_path": true, "created_at": true},
	"handoffs":      {"run_id": true, "check_id": true, "anchors": true, "created_at": true},
	"intakes":       {"type": true, "summary": true, "lane": true, "plan_path": true, "created_at": true},
	"interventions": {"verdict_id": true, "reason": true, "created_at": true},
	"traces":        {"run_id": true, "wave": true, "summary": true, "created_at": true},
	"managed_docs":  {"path": true, "installed_sha256": true, "docs_version": true, "updated_at": true},
}

// metaColumns allowlists the meta columns a changeset line may set.
// schema_version and last_applied_changeset are deliberately excluded:
// those are infrastructure-internal bookkeeping owned by Migrate/setFence,
// not something a changeset's fields should ever assign directly.
var metaColumns = map[string]bool{
	"current_phase":   true,
	"entry_phase":     true,
	"latest_run_id":   true,
	"latest_check_id": true,
	"docs_version":    true,
}

// validateFieldNames rejects any key not allowlisted for table, before it
// can be spliced into SQL identifier position.
func validateFieldNames(table string, allowed map[string]bool, keys []string) error {
	for _, k := range keys {
		if !allowed[k] {
			return fmt.Errorf("changeset_malformed: unknown field %q for %s", k, table)
		}
	}
	return nil
}

func validateFieldValues(table string, fields map[string]any) error {
	switch table {
	case "stories":
		if value, ok := fields["status"]; ok {
			return validateEnumValue(table, "status", value, domain.IsValidStoryStatus)
		}
	case "checks":
		if value, ok := fields["verdict"]; ok {
			if err := validateEnumValue(table, "verdict", value, domain.IsValidCheckVerdict); err != nil {
				return err
			}
		}
		if value, ok := fields["judge"]; ok {
			return validateEnumValue(table, "judge", value, domain.IsValidJudge)
		}
	}
	return nil
}

func validateEnumValue(table, column string, value any, valid func(string) bool) error {
	if text, ok := value.(string); ok && valid(text) {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		encoded = []byte(fmt.Sprintf("%q", fmt.Sprint(value)))
	}
	return fmt.Errorf("changeset_malformed: invalid %s.%s value %s", table, column, encoded)
}

// ErrOutOfOrder is returned when a changeset's ULID predates the fence —
// not a hard block (the file is still safely skipped), but flagged per
// CONTRACT.md's `changeset_out_of_order`.
type ErrOutOfOrder struct {
	ULID  string
	Fence string
}

func (e *ErrOutOfOrder) Error() string {
	return fmt.Sprintf("changeset %s predates last-applied %s", e.ULID, e.Fence)
}

type ChangesetIDOverflowError struct {
	Floor string
}

func (e *ChangesetIDOverflowError) Error() string {
	return fmt.Sprintf("changeset ULID overflow above %s", e.Floor)
}

// WriteChangeset mints a new ULID-named file above every canonical filename
// already in dir. Call WriteChangesetAbove when a database fence is available.
func WriteChangeset(dir string, lines []ChangesetLine) (path string, err error) {
	return WriteChangesetAbove(dir, "", lines)
}

// WriteChangesetAbove writes an append-only changeset whose filename is
// strictly greater than both fence and the greatest canonical existing name.
func WriteChangesetAbove(dir, fence string, lines []ChangesetLine) (path string, err error) {
	id, err := NextChangesetID(dir, fence)
	if err != nil {
		return "", err
	}
	return WriteChangesetWithID(dir, id, lines)
}

func NextChangesetID(dir, fence string) (string, error) {
	var floor ulid.ULID
	hasFloor := false
	if fence != "" {
		parsed, err := ulid.ParseStrict(fence)
		if err != nil || parsed.String() != fence {
			return "", fmt.Errorf("changeset fence is not a canonical ULID: %q", fence)
		}
		floor = parsed
		hasFloor = true
	}

	names, err := ListChangesets(dir)
	if err != nil {
		return "", err
	}
	for _, name := range names {
		id, ok := canonicalChangesetID(name)
		if ok && (!hasFloor || id.Compare(floor) > 0) {
			floor = id
			hasFloor = true
		}
	}

	candidate := ulid.Make()
	if !hasFloor || candidate.Compare(floor) > 0 {
		return candidate.String(), nil
	}
	next, ok := incrementULID(floor)
	if !ok {
		return "", &ChangesetIDOverflowError{Floor: floor.String()}
	}
	return next.String(), nil
}

func WriteChangesetWithID(dir, id string, lines []ChangesetLine) (path string, err error) {
	parsed, err := ulid.ParseStrict(id)
	if err != nil || parsed.String() != id {
		return "", fmt.Errorf("changeset id is not a canonical ULID: %q", id)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("changeset dir: %w", err)
	}

	path = filepath.Join(dir, id+".changeset.jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("create changeset file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, line := range lines {
		if err := enc.Encode(line); err != nil {
			return "", fmt.Errorf("write changeset line: %w", err)
		}
	}
	return path, nil
}

func incrementULID(id ulid.ULID) (ulid.ULID, bool) {
	next := id
	for i := len(next) - 1; i >= 0; i-- {
		if next[i] != 0xff {
			next[i]++
			return next, true
		}
		next[i] = 0
	}
	return ulid.ULID{}, false
}

// ReadChangeset parses a `.changeset.jsonl` file into its lines.
func ReadChangeset(path string) ([]ChangesetLine, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open changeset: %w", err)
	}
	defer f.Close()

	var lines []ChangesetLine
	dec := json.NewDecoder(f)
	for dec.More() {
		var line ChangesetLine
		if err := dec.Decode(&line); err != nil {
			return nil, fmt.Errorf("decode changeset line: %w", err)
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// ListChangesets returns `.changeset.jsonl` filenames under dir in ULID
// (chronological) order. A missing dir is reported as an empty list.
func ListChangesets(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list changesets: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".changeset.jsonl") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func changesetULID(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".changeset.jsonl")
}

func canonicalChangesetID(name string) (ulid.ULID, bool) {
	idText := changesetULID(name)
	if idText == filepath.Base(name) {
		return ulid.ULID{}, false
	}
	id, err := ulid.ParseStrict(idText)
	return id, err == nil && id.String() == idText
}

func PendingCanonicalChangesets(db *sql.DB, dir string) (pending []string, fence string, err error) {
	fence, err = readFence(db)
	if err != nil {
		return nil, "", err
	}
	names, err := ListChangesets(dir)
	if err != nil {
		return nil, "", err
	}
	for _, name := range names {
		id, ok := canonicalChangesetID(name)
		if ok && (fence == "" || id.String() > fence) {
			pending = append(pending, name)
		}
	}
	return pending, fence, nil
}

// ApplyChangeset applies a single changeset file to db, honoring the
// file-level idempotency fence (`meta.last_applied_changeset`). A file
// whose ULID is <= the fence is skipped entirely; one strictly less than
// the fence additionally returns *ErrOutOfOrder (flagged, not fatal —
// the file is still safely skipped).
func ApplyChangeset(db *sql.DB, path string) (applied int, skippedAlreadyApplied int, err error) {
	parsedID, ok := canonicalChangesetID(filepath.Base(path))
	if !ok {
		return 0, 0, fmt.Errorf("changeset_malformed: filename must be a canonical ULID changeset: %s", filepath.Base(path))
	}
	id := parsedID.String()

	lines, err := ReadChangeset(path)
	if err != nil {
		return 0, 0, err
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("begin apply tx: %w", err)
	}
	defer tx.Rollback()

	// Acquire SQLite's write reservation before reading the fence. Public
	// callers also hold the repository lock, but this prevents bypass callers
	// from making skip/order decisions from the same pre-transaction fence.
	if _, err := tx.Exec(`UPDATE meta SET last_applied_changeset = last_applied_changeset`); err != nil {
		return 0, 0, fmt.Errorf("lock apply tx: %w", err)
	}
	fence, err := readFence(tx)
	if err != nil {
		return 0, 0, err
	}
	if fence != "" && id <= fence {
		if id < fence {
			return 0, len(lines), &ErrOutOfOrder{ULID: id, Fence: fence}
		}
		return 0, len(lines), nil
	}

	for _, line := range lines {
		if err := applyLine(tx, line); err != nil {
			return 0, 0, err
		}
		applied++
	}

	if err := setFence(tx, id); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit apply tx: %w", err)
	}
	return applied, 0, nil
}

// Replay drop-and-rebuilds materialized state: it applies every
// changeset under dir, in ULID order, against db (expected to be freshly
// migrated / empty of fence state).
func Replay(db *sql.DB, dir string) (totalApplied int, err error) {
	names, err := ListChangesets(dir)
	if err != nil {
		return 0, err
	}
	for _, name := range names {
		applied, _, err := ApplyChangeset(db, filepath.Join(dir, name))
		if err != nil {
			return totalApplied, err
		}
		totalApplied += applied
	}
	return totalApplied, nil
}

// ChangesetStatus reports pending vs. applied changesets against the
// current fence (CONTRACT.md `db changeset status`). "Applied" for a
// changeset at or below the fence is only ever an assumption from
// filename order — nothing individually records that each one actually
// ran. UnverifiedBelowFence names those where that assumption does not
// hold: a create-op entity id the changeset should have inserted is
// missing from its table, which happens when a changeset from another
// machine lands below a fence that already advanced past its ULID
// region (see docs/audit/workflow-harness-ceremony-audit.md, F5 — the
// false all-clear this replaces). Update-only changesets cannot be
// verified this way and are not flagged.
func ChangesetStatus(db *sql.DB, dir string) (pending []string, appliedCount int, lastApplied string, unverifiedBelowFence []string, err error) {
	fence, err := readFence(db)
	if err != nil {
		return nil, 0, "", nil, err
	}

	names, err := ListChangesets(dir)
	if err != nil {
		return nil, 0, "", nil, err
	}

	for _, name := range names {
		id := changesetULID(name)
		if fence == "" || id > fence {
			pending = append(pending, name)
			continue
		}
		appliedCount++

		verified, verr := changesetCreatesPresent(db, filepath.Join(dir, name))
		if verr != nil {
			return nil, 0, "", nil, verr
		}
		if !verified {
			unverifiedBelowFence = append(unverifiedBelowFence, name)
		}
	}
	return pending, appliedCount, fence, unverifiedBelowFence, nil
}

// changesetCreatesPresent spot-checks a changeset's create-op entity ids
// against their tables. It reports false the moment one is missing —
// proof the changeset was never actually applied despite being at or
// below the fence.
func changesetCreatesPresent(db *sql.DB, path string) (bool, error) {
	lines, err := ReadChangeset(path)
	if err != nil {
		return false, fmt.Errorf("verify %s: %w", filepath.Base(path), err)
	}
	for _, line := range lines {
		if line.Op != "create" || line.Entity == "meta" {
			continue
		}
		table, ok := entityTables[line.Entity]
		if !ok {
			continue
		}
		var found string
		q := fmt.Sprintf("SELECT id FROM %s WHERE id = ?", table) //nolint:gosec // table is allowlist-derived, not user input
		switch err := db.QueryRow(q, line.ID).Scan(&found); {
		case err == sql.ErrNoRows:
			return false, nil
		case err != nil:
			return false, fmt.Errorf("verify %s in %s: %w", line.ID, table, err)
		}
	}
	return true, nil
}

type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

func readFence(db queryRower) (string, error) {
	var fence sql.NullString
	err := db.QueryRow(`SELECT last_applied_changeset FROM meta LIMIT 1`).Scan(&fence)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read fence: %w", err)
	}
	return fence.String, nil
}

func setFence(tx *sql.Tx, id string) error {
	if _, err := tx.Exec(`UPDATE meta SET last_applied_changeset = ?`, id); err != nil {
		return fmt.Errorf("update fence: %w", err)
	}
	return nil
}

func applyLine(tx *sql.Tx, line ChangesetLine) error {
	if line.Entity == "meta" {
		return applyMetaLine(tx, line)
	}

	table, ok := entityTables[line.Entity]
	if !ok {
		return fmt.Errorf("changeset_malformed: unknown entity %q", line.Entity)
	}

	switch line.Op {
	case "create":
		return applyCreate(tx, table, line)
	case "update":
		return applyUpdate(tx, table, line)
	default:
		return fmt.Errorf("changeset_malformed: unknown op %q", line.Op)
	}
}

// applyCreate is idempotent via INSERT OR IGNORE keyed on the PK `id`
// (Idempotency Key Rules, layer 2) — id is minted once and never
// re-minted on replay, so a duplicate create line is a guaranteed no-op.
func applyCreate(tx *sql.Tx, table string, line ChangesetLine) error {
	if line.ID == "" {
		return fmt.Errorf("changeset_malformed: create %s missing id", table)
	}

	keys := sortedKeys(line.Fields)
	if err := validateFieldNames(table, entityColumns[table], keys); err != nil {
		return err
	}
	if err := validateFieldValues(table, line.Fields); err != nil {
		return err
	}

	cols := []string{"id"}
	vals := []any{line.ID}
	for _, k := range keys {
		cols = append(cols, k)
		vals = append(vals, encodeValue(line.Fields[k]))
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",")
	q := fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ", "), placeholders)
	if _, err := tx.Exec(q, vals...); err != nil {
		return fmt.Errorf("changeset_malformed: insert %s: %w", table, err)
	}
	return nil
}

// applyUpdate is idempotent via a plain UPDATE ... WHERE id = ? — setting
// a column to a value it already holds changes nothing.
func applyUpdate(tx *sql.Tx, table string, line ChangesetLine) error {
	if line.ID == "" {
		return fmt.Errorf("changeset_malformed: update %s missing id", table)
	}
	keys := sortedKeys(line.Fields)
	if len(keys) == 0 {
		return nil
	}
	if err := validateFieldNames(table, entityColumns[table], keys); err != nil {
		return err
	}
	if err := validateFieldValues(table, line.Fields); err != nil {
		return err
	}

	setClauses := make([]string, len(keys))
	vals := make([]any, 0, len(keys)+1)
	for i, k := range keys {
		setClauses[i] = k + " = ?"
		vals = append(vals, encodeValue(line.Fields[k]))
	}
	vals = append(vals, line.ID)

	q := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(setClauses, ", "))
	if _, err := tx.Exec(q, vals...); err != nil {
		return fmt.Errorf("changeset_malformed: update %s: %w", table, err)
	}
	return nil
}

// applyMetaLine updates the single meta row directly (no id/WHERE — meta
// has no primary key, see SCHEMA.md).
func applyMetaLine(tx *sql.Tx, line ChangesetLine) error {
	if line.Op != "update" {
		return fmt.Errorf("changeset_malformed: meta only supports update, got %q", line.Op)
	}
	keys := sortedKeys(line.Fields)
	if len(keys) == 0 {
		return nil
	}
	if err := validateFieldNames("meta", metaColumns, keys); err != nil {
		return err
	}

	setClauses := make([]string, len(keys))
	vals := make([]any, len(keys))
	for i, k := range keys {
		setClauses[i] = k + " = ?"
		vals[i] = encodeValue(line.Fields[k])
	}

	q := fmt.Sprintf("UPDATE meta SET %s", strings.Join(setClauses, ", "))
	if _, err := tx.Exec(q, vals...); err != nil {
		return fmt.Errorf("changeset_malformed: update meta: %w", err)
	}
	return nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// encodeValue converts a decoded-JSON field value into a database/sql
// driver value: composite JSON values (arrays/objects — e.g. trace_ids,
// proof_links, anchors) are re-marshaled to their TEXT-column JSON form.
func encodeValue(v any) any {
	switch v.(type) {
	case []any, map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return string(b)
	default:
		return v
	}
}

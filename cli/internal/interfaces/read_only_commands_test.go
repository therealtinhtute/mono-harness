package interfaces

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

type readOnlyFileSnapshot struct {
	exists  bool
	sha256  [sha256.Size]byte
	size    int64
	modTime time.Time
}

type readOnlyCommandSnapshot struct {
	database      readOnlyFileSnapshot
	journalHeader [2]byte
	changesets    map[string]readOnlyFileSnapshot
	wal           readOnlyFileSnapshot
	shm           readOnlyFileSnapshot
}

func TestInspectionCommandsDoNotCreateWALSidecars(t *testing.T) {
	commands := []struct {
		name         string
		args         []string
		wantFragment string
	}{
		{name: "preflight full", args: []string{"preflight", "check", "--mode", "full", "--json"}, wantFragment: `"db":"ready"`},
		{name: "query state", args: []string{"query", "state", "--json"}, wantFragment: `"current_phase":"beta"`},
		{name: "query phases", args: []string{"query", "phases", "--json"}, wantFragment: `"slug":"beta"`},
		{name: "query traces", args: []string{"query", "traces", "--json"}, wantFragment: `"summary":"wave one done"`},
		{name: "query decisions", args: []string{"query", "decisions", "--json"}, wantFragment: `"decision":"used the beta phase"`},
		{name: "resume", args: []string{"resume", "--json"}, wantFragment: `"current_phase":"beta"`},
		{name: "validate", args: []string{"validate", "--json"}, wantFragment: `"valid":true`},
		{name: "audit", args: []string{"audit", "--json"}, wantFragment: `"contract_violations":[]`},
	}

	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.MkdirAll(filepath.Join("docs", "playbooks"), 0o755); err != nil {
				t.Fatalf("MkdirAll playbooks: %v", err)
			}
			if err := os.WriteFile(filepath.Join("docs", "playbooks", "check.md"), []byte("# Check\n"), 0o644); err != nil {
				t.Fatalf("WriteFile check playbook: %v", err)
			}
			before := prepareReadOnlyCommandDatabase(t, func(db *sql.DB) {
				if _, err := db.Exec(`INSERT INTO stories (id, slug, goal, status, created_at)
					VALUES (?, 'beta', 'persisted goal', 'planned', '2026-07-27T00:00:00Z')`, ulid.Make().String()); err != nil {
					t.Fatalf("seed beta story: %v", err)
				}
				if _, err := db.Exec(`UPDATE meta SET current_phase = 'beta'`); err != nil {
					t.Fatalf("seed current phase: %v", err)
				}
				runID := ulid.Make().String()
				if _, err := db.Exec(`INSERT INTO runs (id, story_slug, trace_ids, artifact_path, created_at)
					VALUES (?, 'beta', '[]', '', '2026-07-27T00:00:00Z')`, runID); err != nil {
					t.Fatalf("seed run for traces: %v", err)
				}
				if _, err := db.Exec(`INSERT INTO traces (id, run_id, wave, summary, created_at)
					VALUES (?, ?, 1, 'wave one done', '2026-07-27T00:00:01Z')`, ulid.Make().String(), runID); err != nil {
					t.Fatalf("seed trace: %v", err)
				}
				if _, err := db.Exec(`INSERT INTO decisions (id, run_id, phase, decision, rationale, created_at)
					VALUES (?, ?, 'beta', 'used the beta phase', 'seeded for read-only inspection test', '2026-07-27T00:00:02Z')`,
					ulid.Make().String(), runID); err != nil {
					t.Fatalf("seed decision: %v", err)
				}
			})

			output := executeReadOnlyJSONCommand(t, command.args...)
			if !bytes.Contains(output, []byte(command.wantFragment)) {
				t.Fatalf("%s output = %s, want fragment %s", command.name, output, command.wantFragment)
			}
			assertReadOnlyCommandState(t, before, captureReadOnlyCommandState(t))
		})
	}
}

func TestNextOpensDatabaseReadOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(filepath.Join("docs", "plans", "active"), 0o755); err != nil {
		t.Fatalf("MkdirAll active plans: %v", err)
	}
	plan := "# Active plan\n\n## Phases and Verification\n" +
		"### Phase 1: Alpha\n- phase_slug: alpha\n- goal: goal\n\n" +
		"### Phase 2: Beta\n- phase_slug: beta\n- goal: goal\n"
	if err := os.WriteFile(filepath.Join("docs", "plans", "active", "initiative.md"), []byte(plan), 0o644); err != nil {
		t.Fatalf("WriteFile active plan: %v", err)
	}

	before := prepareReadOnlyCommandDatabase(t, func(db *sql.DB) {
		if _, err := db.Exec(`INSERT INTO stories (id, slug, goal, status, created_at) VALUES
			(?, 'alpha', 'goal', 'done', '2026-07-27T00:00:00Z'),
			(?, 'beta', 'goal', 'planned', '2026-07-27T00:00:00Z')`, ulid.Make().String(), ulid.Make().String()); err != nil {
			t.Fatalf("seed next stories: %v", err)
		}
	})

	output := executeReadOnlyJSONCommand(t, "next", "full", "--json")
	var got struct {
		Mode        string          `json:"mode"`
		ActivePhase *string         `json:"active_phase"`
		Stop        json.RawMessage `json:"stop"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode next output %q: %v", output, err)
	}
	if got.Mode != "full" || got.ActivePhase == nil || *got.ActivePhase != "beta" || len(got.Stop) != 0 {
		t.Fatalf("next output = %s, want DB-backed active_phase beta without stop", output)
	}

	assertReadOnlyCommandState(t, before, captureReadOnlyCommandState(t))
}

func TestChangesetStatusOpensDatabaseReadOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	var appliedPath, pendingPath string
	before := prepareReadOnlyCommandDatabase(t, func(db *sql.DB) {
		storyID := ulid.Make().String()
		var err error
		appliedPath, err = infrastructure.WriteChangeset(changesetDir, []infrastructure.ChangesetLine{{
			Op: "create", Entity: "story", ID: storyID,
			Fields: map[string]any{"slug": "status-fixture", "goal": "goal", "status": "planned", "created_at": "2026-07-27T00:00:00Z"},
			At:     "2026-07-27T00:00:00Z",
		}})
		if err != nil {
			t.Fatalf("write applied changeset: %v", err)
		}
		if _, _, err := infrastructure.ApplyChangeset(db, appliedPath); err != nil {
			t.Fatalf("apply changeset: %v", err)
		}
		pendingPath, err = infrastructure.WriteChangeset(changesetDir, []infrastructure.ChangesetLine{{
			Op: "update", Entity: "story", ID: storyID,
			Fields: map[string]any{"status": "done"},
			At:     "2026-07-27T00:01:00Z",
		}})
		if err != nil {
			t.Fatalf("write pending changeset: %v", err)
		}
	})

	output := executeReadOnlyJSONCommand(t, "db", "changeset", "status", "--json")
	var got struct {
		Pending      []string `json:"pending"`
		AppliedCount int      `json:"applied_count"`
		LastApplied  string   `json:"last_applied"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode changeset status output %q: %v", output, err)
	}
	wantPending := []string{filepath.Base(pendingPath)}
	wantLastApplied := strings.TrimSuffix(filepath.Base(appliedPath), ".changeset.jsonl")
	if !reflect.DeepEqual(got.Pending, wantPending) || got.AppliedCount != 1 || got.LastApplied != wantLastApplied {
		t.Fatalf("changeset status output = %s, want pending=%v applied_count=1 last_applied=%q", output, wantPending, wantLastApplied)
	}

	assertReadOnlyCommandState(t, before, captureReadOnlyCommandState(t))
}

func prepareReadOnlyCommandDatabase(t *testing.T, setup func(*sql.DB)) readOnlyCommandSnapshot {
	t.Helper()
	db, err := infrastructure.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, err := infrastructure.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Migrate: %v", err)
	}
	setup(db)
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	snapshot := captureReadOnlyCommandState(t)
	if snapshot.journalHeader != [2]byte{2, 2} {
		t.Fatalf("database journal header = %v, want WAL", snapshot.journalHeader)
	}
	if snapshot.wal.exists || snapshot.shm.exists {
		t.Fatalf("sidecars before command: wal=%t shm=%t, want both absent", snapshot.wal.exists, snapshot.shm.exists)
	}
	return snapshot
}

func executeReadOnlyJSONCommand(t *testing.T, args ...string) []byte {
	t.Helper()
	jsonOutput = false
	resumeFacts = ""
	cmd := NewRootCmd("dev")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%v command: %v (output=%s)", args, err, output.String())
	}
	return output.Bytes()
}

func captureReadOnlyCommandState(t *testing.T) readOnlyCommandSnapshot {
	t.Helper()
	database := captureReadOnlyFileSnapshot(t, dbPath)
	if !database.exists {
		t.Fatalf("database %s does not exist", dbPath)
	}
	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("ReadFile database: %v", err)
	}
	if len(data) < 20 {
		t.Fatalf("database header is %d bytes, want at least 20", len(data))
	}
	changesetNames, err := infrastructure.ListChangesets(changesetDir)
	if err != nil {
		t.Fatalf("ListChangesets: %v", err)
	}
	changesets := make(map[string]readOnlyFileSnapshot, len(changesetNames))
	for _, name := range changesetNames {
		changesets[name] = captureReadOnlyFileSnapshot(t, filepath.Join(changesetDir, name))
	}
	return readOnlyCommandSnapshot{
		database:      database,
		journalHeader: [2]byte{data[18], data[19]},
		changesets:    changesets,
		wal:           captureReadOnlyFileSnapshot(t, dbPath+"-wal"),
		shm:           captureReadOnlyFileSnapshot(t, dbPath+"-shm"),
	}
}

func captureReadOnlyFileSnapshot(t *testing.T, path string) readOnlyFileSnapshot {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return readOnlyFileSnapshot{}
	}
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat %s: %v", path, err)
	}
	return readOnlyFileSnapshot{
		exists:  true,
		sha256:  sha256.Sum256(data),
		size:    info.Size(),
		modTime: info.ModTime(),
	}
}

func assertReadOnlyCommandState(t *testing.T, before, after readOnlyCommandSnapshot) {
	t.Helper()
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("inspection command mutated captured state:\nbefore=%+v\nafter=%+v", before, after)
	}
}

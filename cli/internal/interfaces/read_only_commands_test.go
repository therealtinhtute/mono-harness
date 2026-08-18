package interfaces

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/therealtinhtute/skills/cli/internal/infrastructure"
)

// readOnlyMemoryFixtureID is a fixed (non-ULID-minted) id for the memory
// fixture seeded by TestInspectionCommandsDoNotCreateWALSidecars, so the
// static command-args table below can reference it directly.
const readOnlyMemoryFixtureID = "01FIXTUREMEMORYENTRYIDXXXX"

type readOnlyFileSnapshot struct {
	exists  bool
	sha256  [sha256.Size]byte
	size    int64
	modTime time.Time
}

type readOnlyCommandSnapshot struct {
	database      readOnlyFileSnapshot
	journalHeader [2]byte
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
		{name: "memory get", args: []string{"memory", "get", "--id", readOnlyMemoryFixtureID, "--json"}, wantFragment: `"type":"gotcha"`},
		{name: "memory query", args: []string{"memory", "query", "--type", "gotcha", "--json"}, wantFragment: `"type":"gotcha"`},
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
				seedReadOnlyMemoryFixture(t, db)
			})

			output := executeReadOnlyJSONCommand(t, command.args...)
			if !bytes.Contains(output, []byte(command.wantFragment)) {
				t.Fatalf("%s output = %s, want fragment %s", command.name, output, command.wantFragment)
			}
			assertReadOnlyCommandState(t, before, captureReadOnlyCommandState(t))
		})
	}
}

// seedReadOnlyMemoryFixture writes a docs/memory/{id}.md entry and its
// matching memories index row, mirroring CreateMemory's own markdown-first
// output shape, so `memory get`/`memory query` have a real entry to read.
func seedReadOnlyMemoryFixture(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := os.MkdirAll("docs/memory", 0o755); err != nil {
		t.Fatalf("MkdirAll docs/memory: %v", err)
	}
	content := "---\nid: " + readOnlyMemoryFixtureID + "\ntype: gotcha\nscope: global\ncreated: 2026-07-27T00:00:03Z\n---\n\nseeded memory for read-only inspection test\n"
	path := "docs/memory/" + readOnlyMemoryFixtureID + ".md"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
	sha := sha256.Sum256([]byte(content))
	if _, err := db.Exec(
		`INSERT INTO memories (id, path, type, scope, plan_id, sha256, created_at) VALUES (?, ?, 'gotcha', 'global', NULL, ?, '2026-07-27T00:00:03Z')`,
		readOnlyMemoryFixtureID, path, hex.EncodeToString(sha[:]),
	); err != nil {
		t.Fatalf("seed memory fixture: %v", err)
	}
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
	return readOnlyCommandSnapshot{
		database:      database,
		journalHeader: [2]byte{data[18], data[19]},
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

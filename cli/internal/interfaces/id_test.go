package interfaces

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"
)

func runIDCommand(t *testing.T, args ...string) string {
	t.Helper()
	jsonOutput = false
	cmd := NewRootCmd("dev")
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("zharness %v: %v\nstderr: %s", args, err, stderr.String())
	}
	return stdout.String()
}

func TestIDCommandPlain(t *testing.T) {
	out := runIDCommand(t, "id")
	id := strings.TrimSpace(out)
	if _, err := ulid.ParseStrict(id); err != nil {
		t.Fatalf("id output %q is not a valid ULID: %v", id, err)
	}
	if out != id+"\n" {
		t.Fatalf("plain output = %q, want ULID plus newline", out)
	}
}

func TestIDCommandJSON(t *testing.T) {
	out := runIDCommand(t, "id", "--json")
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(out), &response); err != nil {
		t.Fatalf("unmarshal output %q: %v", out, err)
	}
	if _, err := ulid.ParseStrict(response.ID); err != nil {
		t.Fatalf("json id %q is not a valid ULID: %v", response.ID, err)
	}
}

func TestIDCommandMintsUniqueIDs(t *testing.T) {
	first := strings.TrimSpace(runIDCommand(t, "id"))
	second := strings.TrimSpace(runIDCommand(t, "id"))
	if first == second {
		t.Fatalf("consecutive id calls returned the same ULID %q", first)
	}
}

func TestIDCommandRejectsArguments(t *testing.T) {
	jsonOutput = false
	cmd := NewRootCmd("dev")
	cmd.SetArgs([]string{"id", "unexpected"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("id with positional argument succeeded, want an argument error")
	}
}

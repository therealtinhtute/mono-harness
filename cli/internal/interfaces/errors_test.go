package interfaces

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// missingDBError branches its recovery hint on committed plan markdown:
// `db rebuild` restores lifecycle state from docs/plans/*.md (P3 markdown
// truth), while a bare init would create an empty DB contradicting those
// plans; init is only right for a repo with nothing to restore.
func TestMissingDBErrorHintBranchesOnCommittedPlans(t *testing.T) {
	t.Chdir(t.TempDir())

	fresh := missingDBError("query")
	if fresh.Code != "db_unreadable" || fresh.Exit != 2 {
		t.Fatalf("missingDBError() code/exit = %q/%d, want db_unreadable/2", fresh.Code, fresh.Exit)
	}
	if !strings.Contains(fresh.Message, "; run `zharness init` first") {
		t.Fatalf("missingDBError() with no committed plans = %q, want the init hint", fresh.Message)
	}

	if err := os.MkdirAll(filepath.Join("docs", "plans", "active"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("docs", "plans", "active", "pattern-library.md"), []byte("# plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withPlans := missingDBError("query")
	if !strings.Contains(withPlans.Message, "; run `zharness db rebuild --yes` to restore lifecycle state from committed plan markdown") {
		t.Fatalf("missingDBError() with a committed plan = %q, want the rebuild hint", withPlans.Message)
	}
	if !strings.HasPrefix(withPlans.Message, "query: no db at ") {
		t.Fatalf("missingDBError() message = %q, want the query scope prefix", withPlans.Message)
	}
}

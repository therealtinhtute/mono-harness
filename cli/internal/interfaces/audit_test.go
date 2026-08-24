package interfaces

import (
	"errors"
	"strings"
	"testing"
)

// TestMapAuditErrorOwnCodeAndSinglePrefix locks R25's rendering contract: an
// application.Audit failure is a repository-inspection problem, not a
// database problem, so it renders under audit_failed — never db_unreadable —
// and carries exactly one "audit:" prefix, applied here and nowhere else.
func TestMapAuditErrorOwnCodeAndSinglePrefix(t *testing.T) {
	forced := errors.New("read docs directory: not a directory")
	rendered := mapAuditError(forced)

	if rendered.Code == "db_unreadable" {
		t.Fatalf("code = %q, want anything but db_unreadable for a content failure (R25)", rendered.Code)
	}
	if rendered.Code != "audit_failed" {
		t.Fatalf("code = %q, want audit_failed", rendered.Code)
	}
	if n := strings.Count(rendered.Message, "audit:"); n != 1 {
		t.Fatalf("message = %q, want exactly one audit: prefix, got %d", rendered.Message, n)
	}
	if !strings.Contains(rendered.Message, "read docs directory") {
		t.Fatalf("message = %q, want the underlying cause preserved", rendered.Message)
	}
}

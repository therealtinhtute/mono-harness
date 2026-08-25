package application

import "testing"

func TestPreflightMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stage         string
		requestedMode string
		db            string
		docs          string
		wantMode      string
		wantReadiness string
		wantStop      string
	}{
		{name: "ready reduced stage", stage: "watzup", db: PreflightDBReady, docs: PreflightDocsReady, wantMode: "reduced", wantReadiness: PreflightReady},
		{name: "missing db reduced stage", stage: "work", requestedMode: "simple", db: PreflightDBMissing, docs: PreflightDocsReady, wantMode: "reduced", wantReadiness: PreflightReduced},
		{name: "missing docs reduced stage", stage: "interview", db: PreflightDBReady, docs: PreflightDocsMissing, wantMode: "reduced", wantReadiness: PreflightReduced},
		{name: "stale docs reduced stage", stage: "check", requestedMode: "review", db: PreflightDBReady, docs: PreflightDocsStale, wantMode: "reduced", wantReadiness: PreflightReduced},
		{name: "missing db durable stage", stage: "to-plan", requestedMode: "full", db: PreflightDBMissing, docs: PreflightDocsReady, wantMode: "durable", wantReadiness: PreflightBlocked, wantStop: "harness_required"},
		{name: "missing docs durable stage", stage: "work", requestedMode: "full", db: PreflightDBReady, docs: PreflightDocsMissing, wantMode: "durable", wantReadiness: PreflightBlocked, wantStop: "docs_missing"},
		{name: "stale docs durable stage", stage: "check", requestedMode: "full", db: PreflightDBReady, docs: PreflightDocsStale, wantMode: "durable", wantReadiness: PreflightBlocked, wantStop: "stale_docs"},
		{name: "unreadable db always blocks", stage: "git", db: PreflightDBUnreadable, docs: PreflightDocsReady, wantMode: "reduced", wantReadiness: PreflightBlocked, wantStop: "db_unreadable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Preflight(tt.stage, tt.requestedMode, tt.db, tt.docs, ".kit/docs/playbooks/test.md", "dev", false)
			if err != nil {
				t.Fatalf("Preflight() error = %v", err)
			}
			if got.Mode != tt.wantMode || got.Readiness != tt.wantReadiness {
				t.Fatalf("Preflight() mode/readiness = %q/%q, want %q/%q", got.Mode, got.Readiness, tt.wantMode, tt.wantReadiness)
			}
			if tt.wantStop == "" {
				if got.Stop != nil {
					t.Fatalf("Preflight() stop = %+v, want nil", got.Stop)
				}
			} else if got.Stop == nil || got.Stop.Code != tt.wantStop {
				t.Fatalf("Preflight() stop = %+v, want code %q", got.Stop, tt.wantStop)
			}
		})
	}
}

func TestPreflightNormalizesStage(t *testing.T) {
	t.Parallel()

	view, err := Preflight(" WATZUP ", "", PreflightDBReady, PreflightDocsReady, ".kit/docs/playbooks/watzup.md", "dev", false)
	if err != nil {
		t.Fatalf("Preflight() error = %v", err)
	}
	if view.Stage != "watzup" {
		t.Fatalf("Preflight() stage = %q, want watzup", view.Stage)
	}
}

func TestPreflightRejectsInvalidObservedStatus(t *testing.T) {
	t.Parallel()

	if _, err := Preflight("watzup", "", "corrupt", PreflightDocsReady, "", "dev", false); err == nil {
		t.Fatal("Preflight() invalid db status error = nil")
	}
	if _, err := Preflight("watzup", "", PreflightDBReady, "old", "", "dev", false); err == nil {
		t.Fatal("Preflight() invalid docs status error = nil")
	}
}

// A missing database on a durable stage recovers from committed plan
// markdown when it exists (P3 markdown truth) and falls back to init only
// for a repo with nothing to restore.
func TestPreflightMissingDBRecoveryBranchesOnCommittedPlans(t *testing.T) {
	t.Parallel()

	withPlans, err := Preflight("to-plan", "full", PreflightDBMissing, PreflightDocsReady, "", "dev", true)
	if err != nil {
		t.Fatalf("Preflight() error = %v", err)
	}
	if withPlans.Stop == nil || withPlans.Stop.Code != "harness_required" || withPlans.Stop.Recovery != "zharness db rebuild --yes" {
		t.Fatalf("Preflight() stop = %+v, want harness_required recovering via zharness db rebuild --yes when committed plans exist", withPlans.Stop)
	}

	fresh, err := Preflight("to-plan", "full", PreflightDBMissing, PreflightDocsReady, "", "dev", false)
	if err != nil {
		t.Fatalf("Preflight() error = %v", err)
	}
	if fresh.Stop == nil || fresh.Stop.Recovery != "zharness init" {
		t.Fatalf("Preflight() stop = %+v, want harness_required recovering via zharness init with no committed plans", fresh.Stop)
	}
}

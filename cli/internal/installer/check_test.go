package installer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck_CleanInstallHasNoDrift(t *testing.T) {
	isolatedRegistry(t)
	root := tempRepo(t, true)
	mustInstall(t, root)
	got, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("drift on a fresh install: %v", got)
	}
	var sb strings.Builder
	if err := RunCheck(root, false, &sb); err != nil {
		t.Fatalf("RunCheck clean = %v\n%s", err, sb.String())
	}
}

func TestCheck_ReportsEachDriftKind(t *testing.T) {
	isolatedRegistry(t)
	root := tempRepo(t, true)
	mustInstall(t, root)

	names, err := playbookTargets()
	if err != nil || len(names) == 0 {
		t.Fatalf("playbook targets: %v", err)
	}
	stale := names[0].Dst
	pf(t, root, stale, "old playbook\n")
	if err := os.Remove(filepath.Join(root, workflowTarget)); err != nil {
		t.Fatal(err)
	}
	pf(t, root, agentsTarget, "<!-- ZHARNESS:BEGIN -->\nold block\n<!-- ZHARNESS:END -->\n")
	pf(t, root, projectTarget, "# Project\n\n## What is this project?\nx\n")

	got, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, "\n")
	for _, want := range []string{
		"stale    " + stale,
		"missing  " + workflowTarget,
		"stale    " + agentsTarget + " block",
		`heading  docs/PROJECT.md lacks "## Who is it for?"`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, `lacks "## What is this project?"`) {
		t.Errorf("present heading reported as missing:\n%s", joined)
	}
}

func TestCheck_IsReadOnly(t *testing.T) {
	isolatedRegistry(t)
	root := tempRepo(t, true)
	mustInstall(t, root)
	pf(t, root, workflowTarget, "edited\n")
	before := treeSnapshot(t, root)
	var sb strings.Builder
	_ = RunCheck(root, false, &sb)
	if after := treeSnapshot(t, root); after != before {
		t.Fatal("check modified the repository")
	}
}

func TestRunCheck_AllReportsMissingAndFailsOnDrift(t *testing.T) {
	isolatedRegistry(t)
	clean := tempRepo(t, true)
	drifted := tempRepo(t, true)
	mustInstall(t, clean)
	mustInstall(t, drifted)
	pf(t, drifted, workflowTarget, "edited\n")
	gone := filepath.Join(t.TempDir(), "deleted-repo")
	roots, _ := Registered()
	if err := writeRegistry(append(roots, gone)); err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	err := RunCheck("", true, &sb)
	if !errors.Is(err, ErrDrift) {
		t.Fatalf("RunCheck --all = %v, want ErrDrift\n%s", err, sb.String())
	}
	out := sb.String()
	for _, want := range []string{"current  " + clean, "drift    " + drifted, "missing: " + gone, "1 repository(ies) drifted"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRunCheck_AllMissingOnlyIsClean(t *testing.T) {
	isolatedRegistry(t)
	if err := writeRegistry([]string{filepath.Join(t.TempDir(), "gone")}); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	if err := RunCheck("", true, &sb); err != nil {
		t.Fatalf("missing roots alone must not fail: %v\n%s", err, sb.String())
	}
}

func treeSnapshot(t *testing.T, root string) string {
	t.Helper()
	var sb strings.Builder
	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		sb.WriteString(p + "\x00" + sha(b) + "\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return sb.String()
}

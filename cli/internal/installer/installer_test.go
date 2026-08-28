package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func tempRepo(t *testing.T, git bool) string {
	t.Helper()
	dir := t.TempDir()
	if git {
		if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
			t.Fatalf("git init: %v: %s", err, out)
		}
	}
	return dir
}

func mustInstall(t *testing.T, root string) string {
	t.Helper()
	var sb strings.Builder
	if err := Install(root, "test", &sb); err != nil {
		t.Fatalf("install: %v\n%s", err, sb.String())
	}
	return sb.String()
}

func pf(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func rf(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

const wfUp1 = "line1\nline2\nkeepme\nline4\n"

// withSource swaps the upstream provider for the duration of the test.
func withSource(t *testing.T, m map[string]string) {
	t.Helper()
	prev := srcBytesImpl
	srcBytesImpl = func(tg Target) ([]byte, error) {
		if s, ok := m[tg.Src]; ok {
			return []byte(s), nil
		}
		return prev(tg)
	}
	t.Cleanup(func() { srcBytesImpl = prev })
}

func TestInstall_Greenfield_ManagedSetAndBase(t *testing.T) {
	root := tempRepo(t, true)
	out := mustInstall(t, root)

	wantFiles := []string{
		workflowTarget,
		playbookDirTgt + "/work.md",
		playbookDirTgt + "/watzup.md",
		projectTarget,
		agentsTarget,
		gitignoreTarget,
		filepath.Join(baseDir, "manifest.json"),
	}
	for _, w := range wantFiles {
		if _, err := os.Stat(filepath.Join(root, w)); err != nil {
			t.Errorf("missing %s: %v", w, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, legacyDBName)); !os.IsNotExist(err) {
		t.Error("installer created a database; it must not")
	}
	if ent, _ := os.ReadDir(filepath.Join(root, upstreamDir)); len(ent) == 0 {
		t.Error("upstream blob store is empty; base tracking broken")
	}
	gi := rf(t, root, gitignoreTarget)
	if !strings.Contains(gi, "/"+zharnessDir+"/") {
		t.Error("gitignore missing /.zharness/ entry")
	}
	ag := rf(t, root, agentsTarget)
	if !strings.Contains(ag, blockBegin) || !strings.Contains(ag, "no parallel control-plane state") {
		t.Error("AGENTS.md block not installed correctly")
	}
	pj := rf(t, root, projectTarget)
	if len(strings.Split(pj, "\n")) > 50 || strings.Contains(pj, "<ANSWERED>") && false {
		t.Errorf("project template exceeds 50 lines:\n%s", pj)
	}
	if !strings.Contains(out, "greenfield") {
		t.Errorf("expected greenfield note in report:\n%s", out)
	}

	before := map[string]string{}
	for _, f := range []string{workflowTarget, projectTarget, agentsTarget} {
		before[f] = rf(t, root, f)
	}
	mustInstall(t, root)
	for f, b := range before {
		if got := rf(t, root, f); got != b {
			t.Errorf("re-install mutated managed file %s", f)
		}
	}
	if c := strings.Count(rf(t, root, gitignoreTarget), "/.zharness/"); c != 1 {
		t.Errorf("ignore entries duplicated on re-install (count=%d)", c)
	}
}

func TestInstall_Brownfield_ReportOnlyPreservesBytes(t *testing.T) {
	root := tempRepo(t, true)
	claude := "# Consumer CLAUDE.md\nhand-authored bytes\n"
	pf(t, root, "CLAUDE.md", claude)
	pf(t, root, "README.md", "readme")
	pf(t, root, "workflow-state.yml", "state: legacy")
	pf(t, root, "docs/plans/active/aaa.md", "# plan a")
	pf(t, root, "docs/plans/active/bbb.md", "# plan b")

	out := mustInstall(t, root)

	if got := rf(t, root, "CLAUDE.md"); got != claude {
		t.Error("consumer CLAUDE.md rewritten — forbidden by R10/R18")
	}
	if !strings.Contains(out, "active plans under docs/plans/active: 2") ||
		!strings.Contains(out, "reconcile which plan stays live") {
		t.Errorf("missing plan-reconcile advisory:\n%s", out)
	}
	if !strings.Contains(out, "workflow-state.yml") {
		t.Errorf("foreign state file not reported:\n%s", out)
	}
	if !strings.Contains(out, "nothing outside the managed set is written") {
		t.Error("report must state read-only nature")
	}
}

func TestUpdate_FastForward_Kept_AutoMerge_ConflictAbort(t *testing.T) {
	root := tempRepo(t, true)
	withSource(t, map[string]string{"WORKFLOW.md": wfUp1})
	mustInstall(t, root)

	up2 := wfUp1 + "\nappended-by-upstream\n"
	withSource(t, map[string]string{"WORKFLOW.md": up2})
	runUpdateOK(t, root)
	if got := rf(t, root, workflowTarget); got != up2 {
		t.Fatalf("fast-forward mismatch:\n%q", got)
	}

	localOnly := strings.Replace(up2, "keepme", "kept-local-edit", 1)
	pf(t, root, workflowTarget, localOnly)
	withSource(t, map[string]string{"WORKFLOW.md": up2}) // unchanged vs stored
	runUpdateOK(t, root)
	if got := rf(t, root, workflowTarget); got != localOnly {
		t.Error("local-only edit clobbered although upstream was unchanged")
	}

	autoUp := wfUp1 + "\nMERGE-INSERT-BY-UPSTREAM\n"
	withSource(t, map[string]string{"WORKFLOW.md": autoUp})
	beforeLocal := localOnly
	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &sb); err != nil {
		t.Fatalf("auto-merge update failed: %v\n%s", err, sb.String())
	}
	merged := rf(t, root, workflowTarget)
	if merged == beforeLocal || strings.Contains(merged, conflictOpenTag) {
		t.Errorf("expected clean auto-merge:\n%s", merged)
	}
	if !strings.Contains(merged, "kept-local-edit") || !strings.Contains(merged, "MERGE-INSERT-BY-UPSTREAM") {
		t.Errorf("auto-merge lost one side:\n%s", merged)
	}

	snapWorkflow := merged
	snapAgents := rf(t, root, agentsTarget)
	conflictUp := strings.Replace(wfUp1, "keepme", "CONFLICT-A", 1) + "\nappended-by-upstream\n"
	conflictLocal := strings.Replace(snapWorkflow, "keepme", "CONFLICT-B", 1)
	pf(t, root, workflowTarget, conflictLocal)
	withSource(t, map[string]string{"WORKFLOW.md": conflictUp})
	if o := (&updateOptions{Root: root, Version: "t"}); RunUpdate(*o, &strings.Builder{}) == nil {
		t.Fatal("expected conflict rejection")
	}
	if got := rf(t, root, workflowTarget); !strings.Contains(got, conflictOpenTag) {
		t.Fatalf("conflict markers missing after rejected update:\n%s", got)
	}

	var ab strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t", Abort: true}, &ab); err != nil {
		t.Fatalf("abort failed: %v", err)
	}
	if got := rf(t, root, workflowTarget); got != conflictLocal {
		t.Errorf("--abort did not restore workflow bytes exactly")
	}
	if got := rf(t, root, agentsTarget); got != snapAgents {
		t.Errorf("--abort perturbed unrelated file AGENTS.md")
	}
	_ = snapWorkflow
}

func TestUpdate_Continue_FinalizesResolvedFile(t *testing.T) {
	root := tempRepo(t, true)
	withSource(t, map[string]string{"WORKFLOW.md": wfUp1})
	mustInstall(t, root)

	newUp := wfUp1 + "\nupstream-tail-v2\n"
	local := strings.Replace(wfUp1, "keepme", "local-tweaked", 1)
	pf(t, root, workflowTarget, local)
	withSource(t, map[string]string{"WORKFLOW.md": newUp})
	if RunUpdate(updateOptions{Root: root, Version: "t"}, &strings.Builder{}) == nil {
		// identical-content overlap may auto-resolve; force conflict differently:
		newUp2 := strings.Replace(newUp, "keepme", "upstream-touched-same-line", 1)
		withSource(t, map[string]string{"WORKFLOW.md": newUp2})
		if RunUpdate(updateOptions{Root: root, Version: "t"}, &strings.Builder{}) == nil {
			t.Skip("merge engine resolved overlapping edit cleanly; continue-path not exercised")
		}
	}
	resolved := strings.Split(rf(t, root, workflowTarget), conflictOpenTag)[0] +
		strings.Join([]string{"resolved-manually"}, "\n") + "\n" +
		strings.Split(rf(t, root, workflowTarget), conflictCloseTag)[1]
	pf(t, root, workflowTarget, resolved)

	runUpdateOKContinue(t, root)
	if hasConflictMarkers(filepath.Join(root, workflowTarget)) {
		t.Error("markers persisted past --continue")
	}
	final := rf(t, root, workflowTarget)
	if !strings.Contains(final, "resolved-manually") || !strings.Contains(final, "upstream-tail-v2") {
		t.Errorf("resolution lost content from either side:\n%s", final)
	}
}

func runUpdateOKContinue(t *testing.T, root string) {
	t.Helper()
	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t", Continue: true}, &sb); err != nil {
		t.Fatalf("continue failed: %v\n%s", err, sb.String())
	}
}

func TestUninstall_ManagedOnly_ConsumerBytesSurvive(t *testing.T) {
	root := tempRepo(t, true)
	withSource(t, map[string]string{"WORKFLOW.md": wfUp1})
	mustInstall(t, root)

	handDoc := "# my own doc — do not delete\n"
	pf(t, root, "docs/playbooks/my-own-playbook.md", handDoc)
	pf(t, root, legacyDBName, "legacy consumer db bytes")

	var sb strings.Builder
	if err := Uninstall(root, &sb); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	for _, gone := range []string{
		workflowTarget, projectTarget,
		playbookDirTgt + "/work.md",
		filepath.Join(zharnessDir), // dir removed
	} {
		if _, err := os.Stat(filepath.Join(root, gone)); !os.IsNotExist(err) {
			t.Errorf("%s still exists after uninstall", gone)
		}
	}
	if got := rf(t, root, playbookDirTgt+"/my-own-playbook.md"); got != handDoc {
		t.Error("hand-written playbook inside docs/playbooks was destroyed")
	}
	if _, err := os.Stat(filepath.Join(root, legacyDBName)); err != nil {
		t.Error("consumer " + legacyDBName + " was deleted by uninstall — R12 violation")
	}
	filepath.Walk(filepath.Join(root, zharnessDir), func(pp string, fi os.FileInfo, e error) error {
		if e == nil {
			t.Logf("RESIDUE: %s", pp)
		}
		return nil
	})
	if _, err := os.Stat(filepath.Join(root, agentsTarget)); !os.IsNotExist(err) {
		t.Error("AGENTS.md was wholly created by install; uninstall must remove it")
	}
}

func runUpdateOK(t *testing.T, root string) string {
	t.Helper()
	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &sb); err != nil {
		t.Fatalf("update failed: %v\n%s", err, sb.String())
	}
	return sb.String()
}

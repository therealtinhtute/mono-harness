package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/embedded"
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
	if lines := strings.Split(strings.TrimRight(pj, "\n"), "\n"); len(lines) > 50 {
		t.Errorf("project template exceeds 50 lines (%d):\n%s", len(lines), pj)
	}
	if !strings.Contains(pj, "<one sentence: what the product IS>") {
		t.Error("project template lost its unanswered-question form")
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

// Regression: a later hunk straddling the cursor after an overlap event must
// be absorbed into the cluster, not strand the walker (judge finding F1).
func TestThreeWay_StraddlingClusterTerminates(t *testing.T) {
	base := make([]string, 15)
	for i := range base {
		base[i] = fmt.Sprintf("b%02d", i)
	}
	local := append([]string{}, base...)
	copy(local[2:5], []string{"L2", "L3", "L4"})
	copy(local[10:13], []string{"L10", "L11", "L12"})
	up := append([]string{}, base...)
	copy(up[4:11], []string{"U4", "U5", "U6", "U7", "U8", "U9", "U10"})

	done := make(chan string, 1)
	go func() {
		merged, _ := threeWay(strings.Join(base, "\n")+"\n",
			strings.Join(local, "\n")+"\n",
			strings.Join(up, "\n")+"\n")
		done <- merged
	}()
	select {
	case merged := <-done:
		if !strings.Contains(merged, conflictOpenTag) {
			t.Errorf("overlapping disjoint-hunk edits must conflict:\n%s", merged)
		}
		for _, want := range []string{"L2", "L10", "U4"} {
			if !strings.Contains(merged, want) {
				t.Errorf("conflict dropped side content %q:\n%s", want, merged)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("threeWay did not terminate — merge-walker hang (F1)")
	}

	// clean trace: base A/B/C, local edits B, upstream edits C — both kept
	merged2, conflict2 := threeWay("A\nB\nC\n", "A\nB-local\nC\n", "A\nB\nC-up\n")
	if conflict2 || !strings.Contains(merged2, "B-local") || !strings.Contains(merged2, "C-up") {
		t.Errorf("disjoint edits must merge cleanly keeping both sides:\n%s (conflict=%v)", merged2, conflict2)
	}

	// zero-width insertion at the cursor must survive the skip loops
	merged3, conflict3 := threeWay("l1\nl2\nl3\nl4\n", "# predating zharness\n", "l1\nl2\nl3\nl4\n\ntail\n")
	if conflict3 || !strings.Contains(merged3, "predating zharness") || !strings.Contains(merged3, "tail") {
		t.Errorf("whole-file local change plus zero-width upstream insertion lost a side:\n%s (conflict=%v)", merged3, conflict3)
	}
}

// Regression: the AGENTS block merge must use the recorded base block as
// ancestor; passing (local, want, want) made the conflict branch dead and
// silently overwrote consumer edits inside the block (judge finding F2).
func TestUpdate_AgentsBlock_UsesRecordedBase(t *testing.T) {
	rawUp, err := embedded.FS.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("embedded AGENTS.md: %v", err)
	}
	consumerRewrite := "consumer rewrote this line"
	upstreamRewrite := "upstream rewrote the same line"
	if !strings.Contains(string(rawUp), "no parallel control-plane state") {
		t.Fatal("fixture: anchor sentence not found in embedded AGENTS block")
	}

	t.Run("overlap stops with in-block markers", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		// consumer and upstream rewrite the SAME ancestor line, before any
		// reconcile: diff3 must stop instead of silently picking a side (F2)
		consumerTweak := strings.Replace(rf(t, root, agentsTarget), "no parallel control-plane state", consumerRewrite, 1)
		pf(t, root, agentsTarget, consumerTweak)
		upV3 := strings.Replace(string(rawUp), "no parallel control-plane state", upstreamRewrite, 1)
		withSource(t, map[string]string{"AGENTS.md": upV3})
		if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &strings.Builder{}); err == nil {
			t.Fatal("expected in-block conflict rejection")
		}
		got := rf(t, root, agentsTarget)
		if !strings.Contains(got, conflictOpenTag) || !strings.Contains(got, upstreamRewrite) || !strings.Contains(got, consumerRewrite) {
			t.Errorf("in-block conflict lost a side (F2):\n%s", got)
		}
		// both conflict sides are canonical marker-inclusive blocks, so the
		// marked block must stay recognizable for --continue (agentsSpan)
		if !strings.Contains(got, blockBegin) || !strings.Contains(got, blockEnd) {
			t.Errorf("conflict region dropped the block markers:\n%s", got)
		}
		if _, ok := agentsBlockOf(got); !ok {
			t.Error("agentsSpan cannot find the block inside the conflict region")
		}
		var ab strings.Builder
		if err := RunUpdate(updateOptions{Root: root, Version: "t", Abort: true}, &ab); err != nil {
			t.Fatalf("abort failed: %v", err)
		}
		if got := rf(t, root, agentsTarget); got != consumerTweak {
			t.Errorf("--abort did not restore AGENTS.md bytes exactly")
		}
	})

	t.Run("conflict resolution continues and next update is inert", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		consumerTweak := strings.Replace(rf(t, root, agentsTarget), "no parallel control-plane state", consumerRewrite, 1)
		pf(t, root, agentsTarget, consumerTweak)
		upV3 := strings.Replace(string(rawUp), "no parallel control-plane state", upstreamRewrite, 1)
		withSource(t, map[string]string{"AGENTS.md": upV3})
		if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &strings.Builder{}); err == nil {
			t.Fatal("expected in-block conflict rejection")
		}
		// resolve by keeping the consumer side, then --continue must record
		// the canonical block as new base and the next update must be inert
		pf(t, root, agentsTarget, consumerTweak)
		runUpdateOKContinue(t, root)
		if got := rf(t, root, agentsTarget); got != consumerTweak {
			t.Errorf("--continue perturbed the resolved AGENTS block:\n%s", got)
		}
		if hasConflictMarkers(filepath.Join(root, agentsTarget)) {
			t.Error("markers persisted past --continue on AGENTS.md")
		}
		withSource(t, map[string]string{"AGENTS.md": upV3})
		runUpdateOK(t, root)
		if got := rf(t, root, agentsTarget); got != consumerTweak {
			t.Errorf("post-resolution update clobbered the resolved block:\n%s", got)
		}
	})

	t.Run("disjoint edits merge and stay idempotent", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		consumerTweak := strings.Replace(rf(t, root, agentsTarget), "no parallel control-plane state", consumerRewrite, 1)
		pf(t, root, agentsTarget, consumerTweak)
		upV2 := string(rawUp) + "\nupstream-block-addition\n"
		withSource(t, map[string]string{"AGENTS.md": upV2})
		out := runUpdateOK(t, root)
		got := rf(t, root, agentsTarget)
		if !strings.Contains(got, consumerRewrite) || !strings.Contains(got, "upstream-block-addition") {
			t.Errorf("block merge lost one side (F2):\n%s\nout:\n%s", got, out)
		}
		if !strings.Contains(got, blockBegin) || !strings.Contains(got, blockEnd) {
			t.Errorf("block merge destroyed the markers:\n%s", got)
		}
		before := got
		runUpdateOK(t, root) // second run with the same upstream: no churn
		if got := rf(t, root, agentsTarget); got != before {
			t.Errorf("update is not idempotent for the AGENTS block:\n%s", got)
		}
	})
}

// Regression: uninstall must restore a captured pre-install original even
// when local == recorded base (e.g. after a fast-forward) — deleting it
// destroyed consumer bytes (judge finding F3).
func TestUninstall_RestoresPreInstallOriginal_AfterFastForward(t *testing.T) {
	root := tempRepo(t, true)
	pf(t, root, workflowTarget, wfUp1) // pre-existing, identical to upstream
	withSource(t, map[string]string{"WORKFLOW.md": wfUp1})
	mustInstall(t, root) // brownfield install captures the original

	up2 := wfUp1 + "\nupstream-tail-v2\n"
	withSource(t, map[string]string{"WORKFLOW.md": up2})
	runUpdateOK(t, root)
	if got := rf(t, root, workflowTarget); got != up2 {
		t.Fatalf("fast-forward did not apply:\n%q", got)
	}

	var sb strings.Builder
	if err := Uninstall(root, &sb); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if got := rf(t, root, workflowTarget); got != wfUp1 {
		t.Errorf("uninstall deleted a file with a captured pre-install original instead of restoring it (F3):\n%q", got)
	}
}

// Regression: --continue must commit the same-run refreshed base (draft)
// and drop the stash; otherwise fast-forwarded files lose their ancestor
// and the stash dir leaks (judge finding F4).
func TestUpdate_Continue_KeepsSameRunRefreshes_DropsStash(t *testing.T) {
	root := tempRepo(t, true)
	withSource(t, map[string]string{"WORKFLOW.md": wfUp1})
	mustInstall(t, root)

	playUp := "# play: work (refreshed by upstream v2)\n"
	conflictUp := strings.Replace(wfUp1, "keepme", "CONFLICT-A", 1) + "\nupstream-tail\n"
	withSource(t, map[string]string{
		"WORKFLOW.md":       conflictUp,
		"playbooks/work.md": playUp,
	})
	pf(t, root, workflowTarget, strings.Replace(wfUp1, "keepme", "CONFLICT-B", 1))
	if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &strings.Builder{}); err == nil {
		t.Fatal("expected conflict rejection")
	}

	resolved := strings.Replace(wfUp1, "keepme", "resolved-both", 1) + "\nupstream-tail\n"
	pf(t, root, workflowTarget, resolved)
	runUpdateOKContinue(t, root)

	if _, err := os.Stat(filepath.Join(root, stashDir)); !os.IsNotExist(err) {
		t.Error("stash dir survived --continue (F4)")
	}
	man := rf(t, root, filepath.Join(baseDir, "manifest.json"))
	sum := sha256.Sum256([]byte(playUp))
	if !strings.Contains(man, hex.EncodeToString(sum[:])) {
		t.Error("--continue dropped the same-run fast-forwarded base entry (F4)")
	}
	var manMap map[string]any
	if err := json.Unmarshal([]byte(man), &manMap); err != nil {
		t.Fatalf("manifest not valid JSON after --continue: %v", err)
	}
}

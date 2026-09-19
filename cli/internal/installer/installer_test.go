package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
	if _, err := os.Stat(filepath.Join(root, legacyUpstreamDir)); !os.IsNotExist(err) {
		t.Error("install created the pre-ADR 0011 blob store")
	}
	if base, _ := loadBase(root); base[agentsTarget] == "" || base[projectTarget] == "" {
		t.Errorf("manifest lacks recorded hashes: %v", base)
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

// Install() on a repo where a fresh-overwrite target (playbook or
// WORKFLOW.md) already exists with drifted content — e.g. hand-edited before
// zharness ever ran — must overwrite it with upstream bytes unconditionally.
// Only the write-once docs/PROJECT.md is left as found.
func TestInstall_FreshOverwriteTargets_DriftedLocalFileOverwritten(t *testing.T) {
	root := tempRepo(t, true)
	pf(t, root, workflowTarget, "pre-existing hand-edited WORKFLOW.md\n")
	pf(t, root, playbookDirTgt+"/work.md", "pre-existing hand-edited playbook\n")

	out := mustInstall(t, root)

	if got := rf(t, root, workflowTarget); got == "pre-existing hand-edited WORKFLOW.md\n" {
		t.Error("WORKFLOW.md drift left untouched; fresh-overwrite target must always install upstream bytes")
	}
	if got := rf(t, root, playbookDirTgt+"/work.md"); got == "pre-existing hand-edited playbook\n" {
		t.Error("playbook drift left untouched; fresh-overwrite target must always install upstream bytes")
	}
	if strings.Contains(out, "drifted") {
		t.Errorf("fresh-overwrite targets must never report drifted:\n%s", out)
	}
	if !strings.Contains(out, "installed  "+workflowTarget) {
		t.Errorf("expected installed report for %s:\n%s", workflowTarget, out)
	}
}

// docs/PROJECT.md is write-once (ADR 0011): update scaffolds it when absent
// and never touches an existing one, whatever the template does.
func TestUpdate_Project_WriteOnce(t *testing.T) {
	root := tempRepo(t, true)
	withSource(t, map[string]string{projectTemplate: wfUp1})
	mustInstall(t, root)

	custom := strings.Replace(wfUp1, "keepme", "project-owned answer", 1)
	pf(t, root, projectTarget, custom)
	up2 := wfUp1 + "\n## New question?\n"
	withSource(t, map[string]string{projectTemplate: up2})
	runUpdateOK(t, root)
	if got := rf(t, root, projectTarget); got != custom {
		t.Fatalf("update changed an existing PROJECT.md:\n%q", got)
	}

	if err := os.Remove(filepath.Join(root, projectTarget)); err != nil {
		t.Fatal(err)
	}
	out := runUpdateOK(t, root)
	if got := rf(t, root, projectTarget); got != up2 {
		t.Errorf("absent PROJECT.md not scaffolded from the template:\n%q", got)
	}
	if !strings.Contains(out, "installed") {
		t.Errorf("expected an installed line:\n%s", out)
	}
}

// Playbooks and WORKFLOW.md are pure upstream mirrors (Target.Merge ==
// false): update always overwrites them with upstream bytes, ignoring any
// local edit and never producing a conflict.
func TestUpdate_FreshOverwrite_PlaybooksAndWorkflow_IgnoreLocalEdits(t *testing.T) {
	root := tempRepo(t, true)
	withSource(t, map[string]string{"WORKFLOW.md": wfUp1})
	mustInstall(t, root)

	pf(t, root, workflowTarget, "totally different hand-edited content\n")
	up2 := wfUp1 + "\nupstream-v2\n"
	withSource(t, map[string]string{"WORKFLOW.md": up2})
	runUpdateOK(t, root)
	if got := rf(t, root, workflowTarget); got != up2 {
		t.Errorf("expected fresh overwrite to win over local edit:\n%q", got)
	}

	pf(t, root, playbookDirTgt+"/work.md", "hand-edited playbook\n")
	playUp := "# play: work (v2)\n"
	withSource(t, map[string]string{
		"WORKFLOW.md":       up2,
		"playbooks/work.md": playUp,
	})
	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &sb); err != nil {
		t.Fatalf("update failed: %v\n%s", err, sb.String())
	}
	if got := rf(t, root, playbookDirTgt+"/work.md"); got != playUp {
		t.Errorf("playbook local edit should be silently discarded, got:\n%q", got)
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

// The AGENTS block is replaced between its markers unless it was edited since
// zharness last wrote it; then update refuses before any write and prints the
// diff, and --force replaces it (ADR 0011).
func TestUpdate_AgentsBlock_HashGuard(t *testing.T) {
	rawUp, err := embedded.FS.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("embedded AGENTS.md: %v", err)
	}
	const anchor = "no parallel control-plane state"
	const edit = "consumer rewrote this line"
	if !strings.Contains(string(rawUp), anchor) {
		t.Fatal("fixture: anchor sentence not found in embedded AGENTS block")
	}
	upV2 := string(rawUp) + "\nupstream-block-addition\n"

	t.Run("untouched block is replaced and prose kept", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		pf(t, root, agentsTarget, "consumer prose above\n\n"+rf(t, root, agentsTarget)+"\nconsumer prose below\n")
		withSource(t, map[string]string{"AGENTS.md": upV2})
		runUpdateOK(t, root)
		got := rf(t, root, agentsTarget)
		if !strings.Contains(got, "upstream-block-addition") {
			t.Errorf("block not refreshed:\n%s", got)
		}
		if !strings.HasPrefix(got, "consumer prose above\n") || !strings.HasSuffix(got, "consumer prose below\n") {
			t.Errorf("prose outside the markers changed:\n%s", got)
		}
		runUpdateOK(t, root)
		if again := rf(t, root, agentsTarget); again != got {
			t.Errorf("update is not idempotent for the AGENTS block:\n%s", again)
		}
	})

	t.Run("hand-edited block refuses, writes nothing, force replaces", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		pf(t, root, agentsTarget, strings.Replace(rf(t, root, agentsTarget), anchor, edit, 1))
		pf(t, root, workflowTarget, "stale workflow\n")
		withSource(t, map[string]string{"AGENTS.md": upV2})
		before := treeSnapshot(t, root)
		var sb strings.Builder
		if err := RunUpdate(updateOptions{Root: root, Version: "t"}, &sb); err == nil {
			t.Fatal("expected refusal for a hand-edited block")
		}
		if treeSnapshot(t, root) != before {
			t.Fatal("refused update wrote files")
		}
		out := sb.String()
		if !strings.Contains(out, "+upstream-block-addition") || !strings.Contains(out, "--force") ||
			!strings.Contains(out, "\n-") || !strings.Contains(out, edit) {
			t.Errorf("refusal must print the diff and the --force hint:\n%s", out)
		}

		var fb strings.Builder
		if err := RunUpdate(updateOptions{Root: root, Version: "t", Force: true}, &fb); err != nil {
			t.Fatalf("--force: %v\n%s", err, fb.String())
		}
		if got := rf(t, root, agentsTarget); strings.Contains(got, edit) || !strings.Contains(got, "upstream-block-addition") {
			t.Errorf("--force did not replace the block:\n%s", got)
		}
		runUpdateOK(t, root)
	})

	t.Run("no recorded hash is accepted and recorded", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		pf(t, root, agentsTarget, strings.Replace(rf(t, root, agentsTarget), anchor, edit, 1))
		base, err := loadBase(root)
		if err != nil {
			t.Fatal(err)
		}
		delete(base, agentsTarget)
		if err := saveBase(root, "legacy", base); err != nil {
			t.Fatal(err)
		}
		withSource(t, map[string]string{"AGENTS.md": upV2})
		runUpdateOK(t, root)
		base, _ = loadBase(root)
		if base[agentsTarget] != sha([]byte(canonicalAgentsBlock(upV2))) {
			t.Errorf("block hash not recorded: %q", base[agentsTarget])
		}
	})

	t.Run("CRLF checkout of an untouched block is not a hand edit", func(t *testing.T) {
		root := tempRepo(t, true)
		mustInstall(t, root)
		pf(t, root, agentsTarget, strings.ReplaceAll(rf(t, root, agentsTarget), "\n", "\r\n"))
		withSource(t, map[string]string{"AGENTS.md": upV2})
		runUpdateOK(t, root)
		if got := rf(t, root, agentsTarget); !strings.Contains(got, "upstream-block-addition") {
			t.Errorf("block not refreshed:\n%s", got)
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

// A pre-ADR 0011 installation carries content-addressed blobs under
// .zharness/base/upstream/ and may carry an update stash. Its manifest already
// holds the hashes update needs, so update deletes the blobs, keeps the
// ledger, and leaves the stash for uninstall.
func TestUpdate_LegacyConflicts_RefusesAndCheckReports(t *testing.T) {
	root := tempRepo(t, true)
	mustInstall(t, root)
	pf(t, root, legacyConflictsFile, `["docs/PROJECT.md"]`+"\n")
	pf(t, root, workflowTarget, "stale workflow\n")
	before := treeSnapshot(t, root)
	var sb strings.Builder
	if err := RunUpdate(updateOptions{Root: root, Version: "t", Force: true}, &sb); err == nil {
		t.Fatal("expected refusal while a pre-0011 conflict is unresolved")
	}
	if treeSnapshot(t, root) != before {
		t.Fatal("refused update wrote files")
	}
	if !strings.Contains(sb.String(), legacyConflictsFile) {
		t.Errorf("refusal must name the conflict list:\n%s", sb.String())
	}
	got, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(got, "\n"), "conflict "+legacyConflictsFile) {
		t.Errorf("check must report the unresolved conflict: %v", got)
	}
}

func TestUpdate_LegacyArtifacts_BlobsDroppedLedgerKept(t *testing.T) {
	root := tempRepo(t, true)
	mustInstall(t, root)
	pf(t, root, legacyUpstreamDir+"/deadbeef.bin", "old blob\n")
	pf(t, root, legacyStashDir+"/stash.tsv", "docs/PROJECT.md\tx.bin\t1\n")
	ledger := rf(t, root, ownershipFile)

	runUpdateOK(t, root)
	if _, err := os.Stat(filepath.Join(root, legacyUpstreamDir)); !os.IsNotExist(err) {
		t.Error("update kept the legacy blob store")
	}
	if got := rf(t, root, ownershipFile); got != ledger {
		t.Errorf("update changed the ownership ledger:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(root, legacyStashDir)); err != nil {
		t.Error("update deleted a legacy stash; only uninstall may")
	}

	var sb strings.Builder
	if err := Uninstall(root, &sb); err != nil {
		t.Fatalf("uninstall: %v\n%s", err, sb.String())
	}
	if _, err := os.Stat(filepath.Join(root, zharnessDir)); !os.IsNotExist(err) {
		t.Error("uninstall left .zharness behind")
	}
}

// R3 (guard-v3): the '_' -> "__" + '/' -> "_2F" mapping is injective —
// distinct managed paths can never share an original-file name — and the
// legacy v0.15.0 mapping ('/' -> "__") is still found on upgrade.
func TestSafePath_Injective_AndLegacyFallback(t *testing.T) {
	paths := []string{
		"a/b.md", "a__b.md", "a_2Fb.md", "a/b__c.md", "a__b_2Fc.md",
		"a/b/c.md", "a__b__c.md", workflowTarget, "docs/WORKFLOW_2.md",
	}
	seen := map[string]string{}
	for _, p := range paths {
		s := safePath(p)
		if prev, dup := seen[s]; dup {
			t.Fatalf("safePath collision: %q and %q both map to %q", prev, p, s)
		}
		seen[s] = p
	}
	if got, want := safePath("a/b.md"), "a_2Fb.md"; got != want {
		t.Fatalf("safePath(a/b.md) = %q, want %q", got, want)
	}
	if legacySafePath("a/b.md") != legacySafePath("a__b.md") {
		t.Fatal("fixture: legacy mapping must collide on these two paths")
	}
	if safePath("a/b.md") == safePath("a__b.md") {
		t.Fatal("new mapping must separate the historically colliding paths")
	}

	// end-to-end: an original recorded under the LEGACY name is still
	// found by readOriginal, and captureOriginal never overwrites it.
	root := tempRepo(t, true)
	origDir := filepath.Join(root, originalDir)
	if err := os.MkdirAll(origDir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacyBytes := []byte("legacy-recorded original\n")
	legacyName := filepath.Join(origDir, legacySafePath("docs/x.md")+".orig")
	if err := os.WriteFile(legacyName, legacyBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := readOriginal(root, "docs/x.md")
	if !ok || string(got) != string(legacyBytes) {
		t.Fatalf("legacy original not found: ok=%v got=%q", ok, got)
	}
	pf(t, root, "docs/x.md", "current bytes\n")
	if err := captureOriginal(root, "docs/x.md"); err != nil {
		t.Fatalf("captureOriginal: %v", err)
	}
	if _, err := os.Stat(filepath.Join(origDir, safePath("docs/x.md")+".orig")); !os.IsNotExist(err) {
		t.Error("captureOriginal wrote a second original despite the legacy one")
	}
	if b, _ := os.ReadFile(legacyName); string(b) != string(legacyBytes) {
		t.Error("captureOriginal perturbed the legacy original")
	}
}

// identityPreEdit is the shipped templates/project.identity.md exactly as it
// stood before the gate-slot rewrite (commit aba7057). It is the base a
// consumer installed against, so the update under test replays the real
// upgrade path rather than a synthetic one.
const identityPreEdit = `# PROJECT — identity (answer inline; this is the single forced write step at
# brainstorm lock; keep the whole file at or under 50 lines)

## What is this project?
- <one sentence: what the product IS>

## Who is it for?
- <primary users/teams>

## Non-goals
- <explicitly excluded scope>

## How do we run the tests?
- ` + "`<exact verification command(s)>`" + `

## Architecture in one breath
- runtime shape: <...>
- where state lives: <...>
- entrypoints: <...>

## What are we working on right now?
- plan: docs/plans/active/<slug>.md (<status>)
`

// A consumer who answered the pre-gate-slot identity template keeps the file
// byte for byte on update; the new question surfaces through `update --check`
// as a missing heading instead of a merge conflict (ADR 0011).
func TestUpdate_IdentityTemplateChange_KeepsProject_CheckNamesHeading(t *testing.T) {
	// Capture the real shipped template before any withSource call: two
	// withSource calls layer, and the first override would become prev.
	postEdit, err := srcBytesImpl(Target{Src: projectTemplate})
	if err != nil {
		t.Fatalf("read shipped %s: %v", projectTemplate, err)
	}
	if !strings.Contains(string(postEdit), "## What are the gate commands?") {
		t.Fatalf("shipped template lacks the gate section under test:\n%s", postEdit)
	}

	root := tempRepo(t, true)
	withSource(t, map[string]string{projectTemplate: identityPreEdit})
	mustInstall(t, root)
	filled := strings.Replace(
		identityPreEdit,
		"- `<exact verification command(s)>`",
		"- `pnpm test`",
		1,
	)
	if filled == identityPreEdit {
		t.Fatal("fixture did not fill the tests answer")
	}
	pf(t, root, projectTarget, filled)

	withSource(t, map[string]string{projectTemplate: string(postEdit)})
	runUpdateOK(t, root)
	if got := rf(t, root, projectTarget); got != filled {
		t.Errorf("update changed the filled identity file:\n%q", got)
	}
	drift, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(drift, "\n"), `lacks "## What are the gate commands?"`) {
		t.Errorf("check does not name the new heading:\n%s", strings.Join(drift, "\n"))
	}
}

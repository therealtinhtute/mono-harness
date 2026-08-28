// Package installer implements the zharness v0.15 managed-set verbs:
// install, update, and uninstall. It owns exactly the managed doc set
// (docs/WORKFLOW.md, docs/playbooks/*.md, the marked AGENTS block,
// docs/PROJECT.md scaffold, .zharness bookkeeping, .gitignore entries);
// nothing else in a repository is created, modified, or deleted.
package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealtinhtute/skills/cli/internal/embedded"
)

// legacyDBName is folded at compile time so the S4 tree scan cannot match
// its own foreign-state probe list.
const legacyDBName = "harness" + ".db"

const (
	zharnessDir      = ".zharness"
	baseDir          = ".zharness/base"
	upstreamDir      = ".zharness/base/upstream"
	originalDir      = ".zharness/base/original"
	stashDir         = ".zharness/update-stash"
	conflictsFile    = ".zharness/conflicts.json"
	manifestFile     = ".zharness/base/manifest.json"
	projectTemplate  = "templates/project.identity.md"
	blockBegin       = "<!-- ZHARNESS:BEGIN -->"
	blockEnd         = "<!-- ZHARNESS:END -->"
	gitignoreMarker  = "# zharness v0.15 managed set"
	playbookDirTgt   = "docs/playbooks"
	workflowTarget   = "docs/WORKFLOW.md"
	projectTarget    = "docs/PROJECT.md"
	agentsTarget     = "AGENTS.md"
	gitignoreTarget  = ".gitignore"
	conflictOpenTag  = "<<<<<<< zharness update (incoming)"
	conflictSepTag   = "======="
	conflictCloseTag = ">>>>>>> zharness (local)"
)

// Target is one managed-file mapping from an embedded source path to a
// destination path inside the consuming repository.
type Target struct {
	Src string // path inside embedded.FS (or embedded.Templates)
	Dst string // repo-root-relative destination
}

func playbookTargets() ([]Target, error) {
	names, err := embedded.PlaybookNames()
	if err != nil {
		return nil, err
	}
	tg := make([]Target, 0, len(names))
	for _, n := range names {
		tg = append(tg, Target{Src: "playbooks/" + n, Dst: playbookDirTgt + "/" + n})
	}
	return tg, nil
}

func AllTargets() ([]Target, error) {
	tg, err := playbookTargets()
	if err != nil {
		return nil, err
	}
	all := append([]Target{
		{Src: "WORKFLOW.md", Dst: workflowTarget},
		{Src: projectTemplate, Dst: projectTarget},
	}, tg...)
	return all, nil
}

// srcBytesImpl resolves a target's upstream bytes; swappable in tests.
var srcBytesImpl = embeddedSrc

func embeddedSrc(t Target) ([]byte, error) {
	b, err := embedded.FS.ReadFile(t.Src)
	if err != nil && strings.HasPrefix(t.Src, projectTemplate) {
		b, err = embedded.Templates.ReadFile(t.Src) // embed keeps the templates/ prefix
	}
	return b, err
}

func srcBytes(t Target) ([]byte, error) { return srcBytesImpl(t) }

// ---- brownfield detection (R10): deterministic, read-only ----------------

type brownfieldReport struct {
	ActivePlans  int
	Present      []string
	ForeignState []string
}

func detectBrownfield(root string, r *brownfieldReport) {
	r.ActivePlans = 0
	if entries, err := os.ReadDir(filepath.Join(root, "docs/plans/active")); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") && !strings.HasPrefix(e.Name(), ".") {
				r.ActivePlans++
			}
		}
	}
	for _, p := range []string{"README.md", "CLAUDE.md", agentsTarget} {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			r.Present = append(r.Present, p)
		}
	}
	if fi, err := os.Stat(filepath.Join(root, "docs")); err == nil && fi.IsDir() {
		r.Present = append(r.Present, "docs/**")
	}
	for _, f := range []string{legacyDBName, "workflow-state.yml", ".kit"} {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			r.ForeignState = append(r.ForeignState, f)
		}
	}
	sortStrings(r.Present)
	sortStrings(r.ForeignState)
}

func writeReport(out *strings.Builder, r *brownfieldReport) {
	fmt.Fprintf(out, "\nbrownfield scan (read-only)\n")
	fmt.Fprintf(out, "- active plans under docs/plans/active: %d", r.ActivePlans)
	switch {
	case r.ActivePlans >= 2:
		fmt.Fprintf(out, "  \u2192 2 or more: reconcile which plan stays live before locking a new one")
	case r.ActivePlans == 1:
		fmt.Fprintf(out, "  \u2192 refine the existing plan instead of creating another")
	default:
		fmt.Fprintf(out, "  \u2192 greenfield: this install will become the first lock target")
	}
	out.WriteString("\n")
	fmt.Fprintf(out, "- present inputs for HARVEST drafting: %s\n", joinOrNone(r.Present))
	fmt.Fprintf(out, "- foreign state answering the same questions: %s\n", joinOrNone(r.ForeignState))
	fmt.Fprintf(out, "- proposals here are advisory only; nothing outside the managed set is written\n\n")
}

func joinOrNone(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}

// safePath maps a managed relative path to a flat, collision-free file
// name component: '_' -> "__" and '/' -> "_2F". The two replacement
// images form a prefix-free code, so the mapping is injective — distinct
// managed paths can never collide under .zharness/base/original/.
func safePath(p string) string {
	return strings.NewReplacer("_", "__", "/", "_2F").Replace(p)
}

// legacySafePath is the v0.15.0 mapping ('/' -> "__"), which was NOT
// injective. Kept only so originals captured by that release still
// protect uninstall after an upgrade.
func legacySafePath(p string) string { return strings.ReplaceAll(p, "/", "__") }

// findOriginal returns the on-disk original path for dst, preferring the
// current mapping and falling back to the legacy one.
func findOriginal(root, dst string) (string, bool) {
	for _, name := range []string{safePath(dst) + ".orig", legacySafePath(dst) + ".orig"} {
		p := filepath.Join(root, originalDir, name)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func shaSum(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

func sha(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type manifestEntry struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
}

type manifest struct {
	ZharnessVersion string          `json:"zharness_version"`
	InstalledAt     string          `json:"installed_at"`
	Files           []manifestEntry `json:"files"`
}

func loadBase(root string) (*manifest, map[string][]byte, error) {
	m := &manifest{}
	raw, err := os.ReadFile(filepath.Join(root, manifestFile))
	if os.IsNotExist(err) {
		return m, map[string][]byte{}, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read base manifest: %w", err)
	}
	if err := jsonUnmarshal(raw, m); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", manifestFile, err)
	}
	files := map[string][]byte{}
	for _, fe := range m.Files {
		data, rerr := os.ReadFile(filepath.Join(root, upstreamDir, fe.SHA+".bin"))
		if rerr != nil {
			return nil, nil, fmt.Errorf("read base blob for %s: %w", fe.Path, rerr)
		}
		files[fe.Path] = data
	}
	return m, files, nil
}

func saveBase(root string, ver string, files map[string][]byte) error {
	if err := os.RemoveAll(filepath.Join(root, upstreamDir)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, upstreamDir), 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sortStrings(keys)
	entries := make([]manifestEntry, 0, len(keys))
	for _, dst := range keys {
		sum := fmt.Sprintf("%x", shaSum(files[dst]))
		blob := filepath.Join(root, upstreamDir, sum+".bin")
		if err := os.WriteFile(blob, files[dst], 0o644); err != nil {
			return err
		}
		entries = append(entries, manifestEntry{Path: dst, SHA: sum})
	}
	m := &manifest{ZharnessVersion: ver, InstalledAt: time.Now().UTC().Format(time.RFC3339), Files: entries}
	out, err := jsonMarshal(m)
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(root, manifestFile), out)
}

func writeFileAtomic(p string, data []byte) error {
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := p + ".tmp-zharness"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func captureOriginal(root, dst string) error {
	src := filepath.Join(root, dst)
	if _, err := os.Stat(src); err != nil {
		return nil // did not exist; uninstall may delete it outright
	}
	// never overwrite a previously captured pristine under either mapping
	if _, ok := findOriginal(root, dst); ok {
		return nil
	}
	dstP := filepath.Join(root, originalDir, safePath(dst)+".orig")
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dstP), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dstP, b, 0o644)
}

func readOriginal(root, dst string) ([]byte, bool) {
	p, ok := findOriginal(root, dst)
	if !ok {
		return nil, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return b, true
}

// Install performs the managed-set scaffolding plus deterministic read-only
// brownfield detection (R8/R10/R18). Consumer-owned files outside the set
// are never written.
func Install(root, version string, stdout *strings.Builder) error {
	report := &brownfieldReport{}
	detectBrownfield(root, report)
	writeReport(stdout, report)

	targets, err := AllTargets()
	if err != nil {
		return err
	}
	_, prev, err := loadBase(root)
	if err != nil {
		return err
	}
	files := map[string][]byte{}
	for k, v := range prev {
		files[k] = v
	}
	var paths []string

	for _, t := range targets {
		up, err := srcBytes(t)
		if err != nil {
			return fmt.Errorf("embed read %s: %w", t.Src, err)
		}
		if err := captureOriginal(root, t.Dst); err != nil {
			return err
		}
		dstP := filepath.Join(root, t.Dst)
		local, lerr := os.ReadFile(dstP)
		switch {
		case os.IsNotExist(lerr):
			if err := os.MkdirAll(filepath.Dir(dstP), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(dstP, up, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "installed  %s\n", t.Dst)
		case string(local) == string(up):
			fmt.Fprintf(stdout, "current    %s\n", t.Dst)
		default:
			// Local drift on a whole-file managed copy: installer never
			// silently overwrites; record upstream for update to reconcile.
			fmt.Fprintf(stdout, "drifted    %s (left untouched; `zharness update` will merge)\n", t.Dst)
		}
		files[t.Dst] = up
		paths = append(paths, t.Dst)
	}

	agentsUp, err := embedded.FS.ReadFile("AGENTS.md")
	if err != nil {
		return err
	}
	newBlock := string(agentsUp)
	ap := filepath.Join(root, agentsTarget)
	existing, rerr := os.ReadFile(ap)
	switch {
	case os.IsNotExist(rerr):
		if err := captureOriginal(root, agentsTarget); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(ap), 0o755); err != nil {
			return err
		}
		replBody, _ := applyAgentsBlock("", newBlock)
		body := "# Agents\n\n" + replBody
		if err := os.WriteFile(ap, []byte(body), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "installed  %s (created)\n", agentsTarget)
	default:
		before := string(existing)
		if err := captureOriginal(root, agentsTarget); err != nil {
			return err
		}
		content, changed := applyAgentsBlock(before, newBlock)
		if changed {
			if err := writeFileAtomic(ap, []byte(content)); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "updated    %s (marked block refreshed)\n", agentsTarget)
		} else {
			fmt.Fprintf(stdout, "current    %s (block already up to date)\n", agentsTarget)
		}
	}
	files[agentsTarget] = []byte(canonicalAgentsBlock(newBlock))
	paths = append(paths, agentsTarget)

	if err := appendGitignoreEntries(root, stdout); err != nil {
		return err
	}
	paths = append(paths, gitignoreTarget)

	return saveBase(root, version, files)
}

// AgentsSpan locates the marked zharness block inclusive of both marker
// comment lines: content[start:end] is the exact bytes of that region.
func agentsSpan(content string) (int, int, bool) {
	bi := strings.Index(content, blockBegin)
	if bi < 0 {
		return -1, -1, false
	}
	ej := strings.Index(content[bi:], blockEnd)
	if ej < 0 {
		return -1, -1, false
	}
	return bi, bi + ej + len(blockEnd), true
}

// canonicalAgentsBlock wraps an embedded AGENTS.md body into its on-disk form.
func canonicalAgentsBlock(body string) string {
	body = strings.TrimRight(body, "\n")
	return blockBegin + "\n" + body + "\n" + blockEnd
}

// applyAgentsBlock swaps the marked block in-place or appends it. Returns
// the new content and whether anything changed.
func applyAgentsBlock(content, embeddedBody string) (string, bool) {
	want := canonicalAgentsBlock(embeddedBody)
	if s, e, ok := agentsSpan(content); ok {
		if content[s:e] == want {
			return content, false
		}
		return content[:s] + want + content[e:], true
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + want + "\n", true
}

var gitignoreWants = []string{
	gitignoreMarker,
	"/" + zharnessDir + "/",
}

func appendGitignoreEntries(root string, stdout *strings.Builder) error {
	gp := filepath.Join(root, gitignoreTarget)
	existing := ""
	if b, err := os.ReadFile(gp); err == nil {
		existing = string(b)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := captureOriginal(root, gitignoreTarget); err != nil {
		return err
	}
	missing := false
	for _, want := range gitignoreWants {
		if !containsLine(existing, want) {
			missing = true
		}
	}
	if !missing {
		fmt.Fprintf(stdout, "current    %s (entries present)\n", gitignoreTarget)
		return nil
	}
	add := ""
	for _, want := range gitignoreWants {
		if containsLine(existing, want) {
			continue
		}
		add += want + "\n"
	}
	body := existing
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += add
	if err := writeFileAtomic(gp, []byte(body)); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "updated    %s (+%s)\n", gitignoreTarget, zharnessDir+"/")
	return nil
}

func containsLine(blob, want string) bool {
	for _, ln := range strings.Split(blob, "\n") {
		if strings.TrimSpace(ln) == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}

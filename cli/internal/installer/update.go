package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func sortStrings(s []string) { sort.Strings(s) }

func dedupe(s []string) []string {
	m := map[string]struct{}{}
	out := []string{}
	for _, x := range s {
		if _, ok := m[x]; !ok {
			m[x] = struct{}{}
			out = append(out, x)
		}
	}
	return out
}

func jsonMarshal(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

func agentsBlockOf(content string) (string, bool) {
	si, ej, ok := agentsSpan(content)
	if !ok {
		return "", false
	}
	return content[si:ej], true
}

// ------------------------------------------------------------- options ---

// UpdateOptions carries update invocation flags (exported for interfaces).
type UpdateOptions = updateOptions

type updateOptions struct {
	Root    string
	Version string
	Force   bool // replace a hand-edited AGENTS block instead of refusing
}

// RunUpdate refreshes the managed set (ADR 0011). Playbooks and WORKFLOW.md
// are overwritten, docs/PROJECT.md is written only when absent, and the
// AGENTS block is replaced between its markers unless it was edited by hand
// since the last write. Every refusal is decided before the first write, so
// a refused update leaves every file as it was.
func RunUpdate(o updateOptions, stdout *strings.Builder) error {
	root := o.Root
	targets, err := AllTargets()
	if err != nil {
		return err
	}
	baseFiles, err := loadBase(root)
	if err != nil {
		return err
	}
	if _, serr := os.Stat(filepath.Join(root, legacyConflictsFile)); serr == nil {
		fmt.Fprintf(stdout, "refused    a pre-0011 update stopped with unresolved conflicts (%s)\n\n", legacyConflictsFile)
		fmt.Fprintf(stdout, "Resolve the conflict markers in the files it lists, delete %s and %s/, then rerun. Nothing was written.\n", legacyConflictsFile, legacyStashDir)
		return fmt.Errorf("unresolved pre-0011 update conflicts in %s", legacyConflictsFile)
	}

	blockB, err := agentBlockBytes()
	if err != nil {
		return err
	}
	want := canonicalAgentsBlock(string(blockB))
	ap := filepath.Join(root, agentsTarget)
	agents, aerr := os.ReadFile(ap)
	if aerr != nil && !os.IsNotExist(aerr) {
		return aerr
	}
	cur, hasBlock := agentsBlockOf(string(agents))
	cur = strings.ReplaceAll(cur, "\r\n", "\n") // a checkout with autocrlf is not a hand edit
	if rec, tracked := baseFiles[agentsTarget]; hasBlock && tracked && !o.Force && cur != want && sha([]byte(cur)) != rec {
		fmt.Fprintf(stdout, "refused    %s: the marked block was edited since zharness last wrote it\n\n", agentsTarget)
		stdout.WriteString(unifiedDiff(agentsTarget+" (on disk)", agentsTarget+" (this zharness)", cur, want))
		fmt.Fprintln(stdout, "\nMove local text outside the markers, or rerun with --force to replace the block. Nothing was written.")
		return fmt.Errorf("%s block edited by hand; rerun with --force to replace it", agentsTarget)
	}

	planned := map[string]string{}
	for _, t := range targets {
		up, uerr := srcBytes(t)
		if uerr != nil {
			return fmt.Errorf("embed read %s: %w", t.Src, uerr)
		}
		dstP := filepath.Join(root, t.Dst)
		note := "refreshed"
		if t.Once {
			if _, serr := os.Stat(dstP); !os.IsNotExist(serr) {
				continue // project-owned once written (ADR 0011)
			}
			note = "installed"
		}
		if werr := writeFileAtomic(dstP, up); werr != nil {
			return werr
		}
		baseFiles[t.Dst] = sha(up)
		planned[t.Dst] = note
	}

	if aerr == nil {
		if next, changed := applyAgentsBlock(string(agents), string(blockB)); changed {
			if werr := writeFileAtomic(ap, []byte(next)); werr != nil {
				return werr
			}
			planned[agentsTarget] = "block refreshed"
			if !hasBlock {
				planned[agentsTarget] = "block appended"
			}
		}
		baseFiles[agentsTarget] = sha([]byte(want))
	}

	giNote, gerr := reconcileGitignore(root, gitignoreWants)
	if gerr == nil && giNote != "" {
		planned[gitignoreTarget] = giNote
	}

	if err := saveBase(root, o.Version, baseFiles); err != nil {
		return err
	}
	register(root, stdout)

	names := make([]string, 0, len(planned))
	for k := range planned {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(stdout, "%-14s %s\n", planned[n], n)
	}
	fmt.Fprintln(stdout, "update complete.")
	return nil
}

func normalizeBlockTail(b string) string { return strings.TrimRight(b, "\n") + "\n" }

// unifiedDiff renders a against b as a single hunk around their differing
// middle with up to three lines of context. The AGENTS block is short, so one
// hunk stays readable and needs no line-matching algorithm.
func unifiedDiff(from, to, a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	p := 0
	for p < len(al) && p < len(bl) && al[p] == bl[p] {
		p++
	}
	s := 0
	for s < len(al)-p && s < len(bl)-p && al[len(al)-1-s] == bl[len(bl)-1-s] {
		s++
	}
	lo := max(p-3, 0)
	tail := min(s, 3)
	var w strings.Builder
	fmt.Fprintf(&w, "--- %s\n+++ %s\n@@ -%d,%d +%d,%d @@\n", from, to,
		lo+1, len(al)-s+tail-lo, lo+1, len(bl)-s+tail-lo)
	for _, l := range al[lo:p] {
		w.WriteString(" " + l + "\n")
	}
	for _, l := range al[p : len(al)-s] {
		w.WriteString("-" + l + "\n")
	}
	for _, l := range bl[p : len(bl)-s] {
		w.WriteString("+" + l + "\n")
	}
	for _, l := range al[len(al)-s : len(al)-s+tail] {
		w.WriteString(" " + l + "\n")
	}
	return w.String()
}

func reconcileGitignore(root string, wants []string) (string, error) {
	gp := filepath.Join(root, gitignoreTarget)
	now, err := os.ReadFile(gp)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	body := ensureLines(string(now), wants)
	if body == string(now) {
		return "", nil
	}
	if err := writeFileAtomic(gp, []byte(body)); err != nil {
		return "", err
	}
	return "+ ignore entries re-asserted", nil
}

// ------------------------------------------------------------ helpers ---

func agentBlockBytes() ([]byte, error) {
	return srcBytes(Target{Src: agentsTarget})
}

func ensureLines(blob string, wants []string) string {
	body := blob
	for _, w := range wants {
		if containsLine(body, w) {
			continue
		}
		if body != "" && !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		body += w + "\n"
	}
	return body
}

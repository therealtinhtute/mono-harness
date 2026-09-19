package installer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrDrift is returned by the check verbs when any repository lags the binary.
var ErrDrift = errors.New("managed docs drift from this zharness binary")

// Check compares a repository's managed set with the embedded bytes without
// writing anything, and returns one line per drifted item.
func Check(root string) ([]string, error) {
	targets, err := AllTargets()
	if err != nil {
		return nil, err
	}
	var drift []string
	for _, t := range targets {
		if t.Merge {
			continue
		}
		want, err := srcBytes(t)
		if err != nil {
			return nil, fmt.Errorf("embed read %s: %w", t.Src, err)
		}
		got, err := os.ReadFile(filepath.Join(root, t.Dst))
		switch {
		case os.IsNotExist(err):
			drift = append(drift, "missing  "+t.Dst)
		case err != nil:
			return nil, err
		case !bytes.Equal(got, want):
			drift = append(drift, "stale    "+t.Dst)
		}
	}

	body, err := agentBlockBytes()
	if err != nil {
		return nil, err
	}
	agents, err := os.ReadFile(filepath.Join(root, agentsTarget))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	block, ok := agentsBlockOf(string(agents))
	switch {
	case !ok:
		drift = append(drift, "missing  "+agentsTarget+" block")
	case normalizeBlockTail(block) != normalizeBlockTail(canonicalAgentsBlock(string(body))):
		drift = append(drift, "stale    "+agentsTarget+" block")
	}

	tmpl, err := srcBytes(Target{Src: projectTemplate})
	if err != nil {
		return nil, err
	}
	project, err := os.ReadFile(filepath.Join(root, projectTarget))
	switch {
	case os.IsNotExist(err):
		drift = append(drift, "missing  "+projectTarget)
	case err != nil:
		return nil, err
	default:
		have := map[string]bool{}
		for _, h := range headings(project) {
			have[h] = true
		}
		for _, h := range headings(tmpl) {
			if !have[h] {
				drift = append(drift, fmt.Sprintf("heading  %s lacks %q", projectTarget, h))
			}
		}
	}
	return drift, nil
}

func headings(b []byte) []string {
	var out []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		if ln := strings.TrimRight(sc.Text(), " \t"); strings.HasPrefix(ln, "## ") {
			out = append(out, ln)
		}
	}
	return out
}

// RunCheck reports drift for one root, or for every registered root when
// all is set. Registered roots that no longer exist are reported, not failed.
func RunCheck(root string, all bool, stdout *strings.Builder) error {
	roots := []string{root}
	if all {
		reg, err := Registered()
		if err != nil {
			return err
		}
		if len(reg) == 0 {
			fmt.Fprintln(stdout, "no registered repositories")
			return nil
		}
		roots = reg
	}
	drifted := 0
	for _, r := range roots {
		if fi, err := os.Stat(r); err != nil || !fi.IsDir() {
			fmt.Fprintf(stdout, "missing: %s\n", r)
			continue
		}
		lines, err := Check(r)
		if err != nil {
			return fmt.Errorf("%s: %w", r, err)
		}
		if len(lines) == 0 {
			fmt.Fprintf(stdout, "current  %s\n", r)
			continue
		}
		drifted++
		fmt.Fprintf(stdout, "drift    %s\n", r)
		for _, l := range lines {
			fmt.Fprintf(stdout, "  %s\n", l)
		}
	}
	if drifted > 0 {
		fmt.Fprintf(stdout, "%d repository(ies) drifted — run `zharness update` in each.\n", drifted)
		return ErrDrift
	}
	return nil
}

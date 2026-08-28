package installer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/therealtinhtute/skills/cli/internal/embedded"
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

// ---------------------------------------------------------------- stash ---

type stashEntry struct {
	existed bool
	data    []byte
}

func stashCapture(root string, paths []string) error {
	sp := filepath.Join(root, stashDir)
	if err := os.MkdirAll(sp, 0o755); err != nil {
		return err
	}
	var meta strings.Builder
	for _, rel := range paths {
		p := filepath.Join(root, rel)
		data, err := os.ReadFile(p)
		existed := err == nil
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if !existed {
			data = nil
		}
		key := sha256.Sum256([]byte(rel))
		name := hex.EncodeToString(key[:12]) + ".bin"
		if existed {
			if err := os.WriteFile(filepath.Join(sp, name), data, 0o644); err != nil {
				return err
			}
		} else {
			if err := os.WriteFile(filepath.Join(sp, name), nil, 0o644); err != nil {
				return err
			}
		}
		fmt.Fprintf(&meta, "%s\t%s\n", rel, name)
	}
	return os.WriteFile(filepath.Join(sp, "stash.tsv"), []byte(meta.String()), 0o644)
}

func stashRestore(root string) error {
	sp := filepath.Join(root, stashDir)
	meta, err := os.ReadFile(filepath.Join(sp, "stash.tsv"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, line := range strings.Split(strings.TrimRight(string(meta), "\n"), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		rel, name := parts[0], parts[1]
		dst := filepath.Join(root, rel)
		raw, rerr := os.ReadFile(filepath.Join(sp, name))
		if rerr != nil {
			continue
		}
		if len(raw) == 0 {
			_ = os.Remove(dst)
			continue
		}
		if err := writeFileAtomic(dst, raw); err != nil {
			return err
		}
	}
	return os.RemoveAll(sp)
}

// ------------------------------------------------------------ conflicts ---

func loadConflicts(root string) []string {
	b, err := os.ReadFile(filepath.Join(root, conflictsFile))
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

func saveConflicts(root string, list []string) error {
	if len(list) == 0 {
		err := os.Remove(filepath.Join(root, conflictsFile))
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	sort.Strings(list)
	return writeFileAtomic(filepath.Join(root, conflictsFile), []byte(strings.Join(list, "\n")+"\n"))
}

func agentsBlockOf(content string) (string, bool) {
	si, ej, ok := agentsSpan(content)
	if !ok {
		return "", false
	}
	return content[si:ej], true
}

func hasConflictMarkers(p string) bool {
	b, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	s := string(b)
	return strings.Contains(s, conflictOpenTag) || strings.Contains(s, conflictCloseTag)
}

// ------------------------------------------------------------- options ---

// UpdateOptions carries update invocation flags (exported for interfaces).
type UpdateOptions = updateOptions

type updateOptions struct {
	Root     string
	Version  string
	Continue bool
	Abort    bool
}

// RunUpdate executes zharness update (R9).
func RunUpdate(o updateOptions, stdout *strings.Builder) error {
	root := o.Root

	switch {
	case o.Abort:
		if err := stashRestore(root); err != nil {
			return err
		}
		_ = saveConflicts(root, nil)
		fmt.Fprintln(stdout, "abort: pre-update state restored byte-for-byte.")
		return nil
	}

	targets, err := AllTargets()
	if err != nil {
		return err
	}
	_, baseFiles, err := loadBase(root)
	if err != nil {
		return err
	}

	if o.Continue {
		return finalizeConflicts(o, baseFiles, stdout)
	}

	if pending := loadConflicts(root); len(pending) > 0 {
		return fmt.Errorf("%d unresolved conflicted file(s): %s — use --continue after resolving markers, or --abort", len(pending), strings.Join(pending, ", "))
	}

	touchables := make([]string, 0, len(targets)+3)
	for _, t := range targets {
		touchables = append(touchables, t.Dst)
	}
	touchables = append(touchables, agentsTarget, gitignoreTarget)
	if err := stashCapture(root, dedupe(touchables)); err != nil {
		return err
	}

	planned := map[string]string{}
	conflicts := []string{}

	for _, t := range targets {
		newUp, uerr := srcBytes(t)
		if uerr != nil {
			return fmt.Errorf("embed read %s: %w", t.Src, uerr)
		}
		oldBase, tracked := baseFiles[t.Dst]
		dstP := filepath.Join(root, t.Dst)
		local, lerr := os.ReadFile(dstP)

		switch {
		case os.IsNotExist(lerr):
			if bytes.Equal(oldBase, newUp) && tracked {
				continue // deleted locally by consumer: respect it
			}
			if werr := os.MkdirAll(filepath.Dir(dstP), 0o755); werr != nil {
				return werr
			}
			if werr := os.WriteFile(dstP, newUp, 0o644); werr != nil {
				return werr
			}
			baseFiles[t.Dst] = newUp
			planned[t.Dst] = "installed"
		case bytes.Equal(local, oldBase):
			if !bytes.Equal(newUp, oldBase) {
				if werr := writeFileAtomic(dstP, newUp); werr != nil {
					return werr
				}
				baseFiles[t.Dst] = newUp
				planned[t.Dst] = "fast-forwarded"
			} else if !bytes.Equal(local, newUp) {
				planned[t.Dst] = "kept as-is"
			}
		case bytes.Equal(local, newUp):
			// already current
		case tracked:
			merged, cflag := threeWay(string(oldBase), string(local), string(newUp))
			if werr := writeFileAtomic(dstP, []byte(merged)); werr != nil {
				return werr
			}
			if cflag {
				conflicts = append(conflicts, t.Dst)
				planned[t.Dst] = "CONFLICT"
			} else {
				baseFiles[t.Dst] = newUp
				planned[t.Dst] = "auto-merged"
			}
		default:
			// consumer drift with no recorded ancestor: never invent one (R18)
			planned[t.Dst] = "kept (local edits beyond recorded history)"
		}
	}

	// AGENTS marked block (surgical, never a whole-file replace over prose)
	blockB, _ := agentBlockBytes()
	want := normalizeBlockTail(string(blockB))
	ap := filepath.Join(root, agentsTarget)
	if localRaw, lerr := os.ReadFile(ap); lerr == nil {
		content := string(localRaw)
		curBlock, hasBlock := agentsBlockOf(content)
		if curBlock != want {
			switch {
			case !hasBlock:
				repl, changed := applyAgentsBlock(content, want)
				if changed {
					_ = writeFileAtomic(ap, []byte(repl))
					planned[agentsTarget] = "block appended"
				}
			case strings.Contains(curBlock, conflictOpenTag):
				// already mid-resolution from an earlier pass
			default:
				merged, cflag := threeWayBlocks(content, curBlock, want)
				if werr := writeFileAtomic(ap, []byte(merged)); werr != nil {
					return werr
				}
				if cflag {
					conflicts = append(conflicts, agentsTarget)
					planned[agentsTarget] = "CONFLICT inside block"
				} else {
					planned[agentsTarget] = "block refreshed"
				}
			}
		}
	}

	giNote, gerr := reconcileGitignore(root, gitignoreWants)
	if gerr == nil && giNote != "" {
		planned[gitignoreTarget] = giNote
	}

	names := make([]string, 0, len(planned))
	for k := range planned {
		names = append(names, k)
	}
	sort.Strings(names)

	if len(conflicts) > 0 {
		// stash + old base stay untouched so --abort restores exactly (R9)
		_ = saveConflicts(root, conflicts)
		fmt.Fprintln(stdout, "update stopped — human resolution required:")
		for _, n := range names {
			fmt.Fprintf(stdout, "%-14s %s\n", classify(planned[n]), n)
		}
		fmt.Fprintf(stdout, "\nAfter clearing the markers run: zharness update --continue\nDiscard everything instead:   zharness update --abort\n")
		return fmt.Errorf("%d file(s) need manual resolution", len(conflicts))
	}

	if err := saveBase(root, o.Version, baseFiles); err != nil {
		return err
	}
	_ = os.RemoveAll(filepath.Join(root, stashDir))

	for _, n := range names {
		fmt.Fprintf(stdout, "%-14s %s\n", planned[n], n)
	}
	fmt.Fprintln(stdout, "update complete.")
	return nil
}

func finalizeConflicts(o updateOptions, baseFiles map[string][]byte, stdout *strings.Builder) error {
	root := o.Root
	pending := loadConflicts(root)
	if len(pending) == 0 {
		fmt.Fprintln(stdout, "continue: nothing pending.")
		return nil
	}
	var still []string
	for _, rel := range pending {
		if hasConflictMarkers(filepath.Join(root, rel)) {
			still = append(still, rel)
			continue
		}
		local, lerr := os.ReadFile(filepath.Join(root, rel))
		if lerr != nil {
			return fmt.Errorf("resolved file vanished: %s", rel)
		}
		baseFiles[rel] = local
		fmt.Fprintf(stdout, "finalized  %s (resolution recorded as new base)\n", rel)
	}
	if len(still) > 0 {
		return fmt.Errorf("markers still present in: %s", strings.Join(still, ", "))
	}
	if err := saveBase(root, o.Version, baseFiles); err != nil {
		return err
	}
	return saveConflicts(root, nil)
}

// threeWayBlocks keeps everything outside the marked block intact and merges
// only the block interior; on conflict the whole file carries markers scoped
// to that region so --continue can detect resolution precisely.
func threeWayBlocks(fileContent, localBlock, wantBlock string) (string, bool) {
	i, jEnd, ok := agentsSpan(fileContent)
	if !ok {
		repl, _ := applyAgentsBlock(fileContent, localBlock)
		return repl, false
	}
	merged, cflag := threeWay(normalizeBlockTail(localBlock), normalizeBlockTail(wantBlock), normalizeBlockTail(wantBlock))
	_ = merged
	if cflag {
		var b strings.Builder
		b.WriteString(conflictOpenTag + " inside marked block\n")
		b.WriteString(wantBlock)
		b.WriteString(conflictSepTag + "\n")
		b.WriteString(localBlock)
		b.WriteString(conflictCloseTag + "\n")
		return fileContent[:i] + b.String() + fileContent[jEnd:], true
	}
	return fileContent[:i] + wantBlock + fileContent[jEnd:], false
}

func normalizeBlockTail(b string) string { return strings.TrimRight(b, "\n") + "\n" }

var _ = conflictSepTag

func currentAgentsBlock(content string) (string, bool) {
	si, ej, ok := agentsSpan(content)
	if !ok {
		return "", false
	}
	end := ej
	if end < len(content) && content[end] == '\n' {
		end++
	}
	return content[si:end], true
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
	return embedded.FS.ReadFile(agentsTarget)
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

func classify(note string) string {
	switch {
	case strings.HasPrefix(note, "CONFLICT"):
		return "conflict:"
	case note == "installed":
		return "installed:"
	case note == "fast-forwarded":
		return "updated:"
	default:
		return "left:"
	}
}

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

// The update draft records what this update run did while conflicts are
// pending: Base carries the files already refreshed (installed /
// fast-forwarded / auto-merged), Upstream carries the exact upstream bytes
// each CONFLICTED file was being reconciled onto. --continue commits Base
// plus the conflicted upstreams as the new base (the resolution itself
// stays in the working tree as local drift, so the next unchanged-upstream
// update can never fast-forward it away); --abort discards the draft with
// the stash.
type updateDraft struct {
	Base     map[string][]byte
	Upstream map[string][]byte
}

func saveBaseDraft(root string, d updateDraft) error {
	b, err := jsonMarshal(d)
	if err != nil {
		return err
	}
	sp := filepath.Join(root, stashDir)
	if err := os.MkdirAll(sp, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sp, "basedraft.json"), b, 0o644)
}

func loadBaseDraft(root string) (updateDraft, bool) {
	b, err := os.ReadFile(filepath.Join(root, stashDir, "basedraft.json"))
	if err != nil {
		return updateDraft{}, false
	}
	var d updateDraft
	if jsonUnmarshal(b, &d) != nil {
		return updateDraft{}, false
	}
	return d, true
}

// draftUpstream returns the conflict-time upstream bytes recorded for rel.
func draftUpstream(root, rel string) ([]byte, bool) {
	d, ok := loadBaseDraft(root)
	if !ok {
		return nil, false
	}
	up, ok := d.Upstream[rel]
	return up, ok
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
	upstreams := map[string][]byte{} // conflict-time upstream per conflicted file
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
				upstreams[t.Dst] = newUp
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

	// AGENTS marked block (surgical, never a whole-file replace over prose).
	// All merge inputs use the canonical marker-inclusive block form so the
	// comparison, the diff3 ancestor, and the on-disk span stay consistent.
	blockB, _ := agentBlockBytes()
	upBody := string(blockB)
	wantFull := normalizeBlockTail(canonicalAgentsBlock(upBody))
	ap := filepath.Join(root, agentsTarget)
	if localRaw, lerr := os.ReadFile(ap); lerr == nil {
		content := string(localRaw)
		curBlock, hasBlock := agentsBlockOf(content)
		curN := normalizeBlockTail(curBlock)
		if curN != wantFull {
			switch {
			case !hasBlock:
				repl, changed := applyAgentsBlock(content, upBody)
				if changed {
					_ = writeFileAtomic(ap, []byte(repl))
					baseFiles[agentsTarget] = []byte(canonicalAgentsBlock(upBody))
					planned[agentsTarget] = "block appended"
				}
			case strings.Contains(curBlock, conflictOpenTag):
				// already mid-resolution from an earlier pass
			default:
				var baseBlock string
				if bb, tracked := baseFiles[agentsTarget]; tracked {
					baseBlock = normalizeBlockTail(string(bb))
				}
				merged, cflag := threeWayBlocks(content, baseBlock, curN, wantFull)
				if werr := writeFileAtomic(ap, []byte(merged)); werr != nil {
					return werr
				}
				if cflag {
					conflicts = append(conflicts, agentsTarget)
					upstreams[agentsTarget] = []byte(canonicalAgentsBlock(upBody))
					planned[agentsTarget] = "CONFLICT inside block"
				} else {
					// base = the upstream block the consumer just reconciled
					// onto, never the merged output (merged output contains
					// local drift; recording it would let a later upstream
					// touch silently win those lines)
					baseFiles[agentsTarget] = []byte(canonicalAgentsBlock(upBody))
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
		// stash, old base, and the in-run draft stay untouched so --abort
		// restores exactly (R9); --continue finalizes from the draft
		_ = saveConflicts(root, conflicts)
		if derr := saveBaseDraft(root, updateDraft{Base: baseFiles, Upstream: upstreams}); derr != nil {
			return derr
		}
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
	if draft, ok := loadBaseDraft(root); ok {
		// same-run installed/fast-forwarded/auto-merged refreshes live in
		// the draft; the on-disk base predates this update run
		baseFiles = draft.Base
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
		if up, ok := draftUpstream(root, rel); ok {
			// base = the upstream version the consumer just reconciled onto
			// (decision: never the merged output — recording the resolution
			// would let the next unchanged-upstream update fast-forward it
			// away). The resolution itself stays as working-tree drift.
			baseFiles[rel] = up
		} else if rel == agentsTarget {
			// legacy fallback: canonical marked block, never whole-file prose
			if inner, ok := agentsBlockOf(string(local)); ok {
				baseFiles[rel] = []byte(strings.TrimRight(inner, "\n"))
			} else {
				delete(baseFiles, rel) // resolution removed the block
			}
		} else {
			baseFiles[rel] = local
		}
		fmt.Fprintf(stdout, "finalized  %s (resolution kept; upstream recorded as new base)\n", rel)
	}
	if len(still) > 0 {
		return fmt.Errorf("markers still present in: %s", strings.Join(still, ", "))
	}
	if err := saveBase(root, o.Version, baseFiles); err != nil {
		return err
	}
	if err := saveConflicts(root, nil); err != nil {
		return err
	}
	_ = os.RemoveAll(filepath.Join(root, stashDir))
	return nil
}
func normalizeBlockTail(b string) string { return strings.TrimRight(b, "\n") + "\n" }

// threeWayBlocks keeps everything outside the marked block intact and merges
// only the block itself; baseBlock, localBlock, and wantBlock are all in the
// canonical marker-inclusive form (baseBlock empty when unrecorded —
// divergence then conflicts, since R18 never invents an ancestor). On
// conflict the whole file carries markers scoped to that region so
// --continue can detect resolution precisely.
func threeWayBlocks(fileContent, baseBlock, localBlock, wantBlock string) (string, bool) {
	i, jEnd, ok := agentsSpan(fileContent)
	if !ok {
		repl, _ := applyAgentsBlock(fileContent, localBlock)
		return repl, false
	}
	merged, cflag := threeWay(normalizeBlockTail(baseBlock), normalizeBlockTail(localBlock), normalizeBlockTail(wantBlock))
	if cflag {
		var b strings.Builder
		b.WriteString(conflictOpenTag + " inside marked block\n")
		b.WriteString(strings.TrimRight(wantBlock, "\n") + "\n")
		b.WriteString(conflictSepTag + "\n")
		b.WriteString(strings.TrimRight(localBlock, "\n") + "\n")
		b.WriteString(conflictCloseTag + "\n")
		return fileContent[:i] + b.String() + fileContent[jEnd:], true
	}
	return fileContent[:i] + strings.TrimRight(merged, "\n") + fileContent[jEnd:], false
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

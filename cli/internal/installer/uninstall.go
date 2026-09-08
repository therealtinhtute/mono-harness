package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Uninstall removes exactly what the ownership ledger attributes to the
// installer (ADR 0008). A managed file whose current bytes differ from both
// its recorded base and the embedded upstream is locally authored work and
// is left in place with a warning; so is anything whose provenance is
// unknown, because a leftover file costs the consumer nothing they cannot
// delete while a wrong guess costs them work they cannot recover.
func Uninstall(root string, stdout *strings.Builder) error {
	targets, err := AllTargets()
	if err != nil {
		return err
	}
	_, baseFiles, err := loadBase(root)
	if err != nil {
		return err
	}
	own := loadOwnershipFor(root, targets, baseFiles)

	removed, kept := 0, []string{}
	for _, t := range targets {
		base, tracked := baseFiles[t.Dst]
		removeManagedFile(root, t.Dst, base, tracked, own.get(kindFile, t.Dst), &removed, &kept, stdout)
	}
	removeAgentsBlock(root, own, &removed, &kept, stdout)

	giNow := readAll(filepath.Join(root, gitignoreTarget))
	cleaned := dropLines(giNow, own.ownedGitignoreLines())
	if !bytesEqual(cleaned, giNow) {
		giOrigin := own.get(kindFile, gitignoreTarget)
		if len(strings.TrimSpace(string(cleaned))) == 0 && giOrigin == originCreated {
			_ = os.Remove(filepath.Join(root, gitignoreTarget))
			fmt.Fprintf(stdout, "removed   %s\n", gitignoreTarget)
			removed++
		} else {
			_ = writeFileAtomic(filepath.Join(root, gitignoreTarget), cleaned)
			fmt.Fprintf(stdout, "restored  %s (zharness entries removed)\n", gitignoreTarget)
		}
	}

	_ = os.RemoveAll(filepath.Join(root, stashDir))
	_ = saveConflicts(root, nil)
	_ = os.Remove(filepath.Join(root, manifestFile))
	_ = os.Remove(filepath.Join(root, ownershipFile))
	_ = os.RemoveAll(filepath.Join(root, upstreamDir))
	_ = os.RemoveAll(filepath.Join(root, originalDir))
	_ = os.RemoveAll(filepath.Join(root, baseDir))
	_ = removeDirIfEmpty(filepath.Join(root, zharnessDir))

	// Report what happened, never a blanket guarantee the run did not
	// enforce: a gate that overstates its own coverage is worse than one
	// that says nothing (ADR 0008).
	if len(kept) == 0 {
		fmt.Fprintf(stdout, "uninstall complete — %d managed file(s) removed; nothing else was touched.\n", removed)
		return nil
	}
	fmt.Fprintf(stdout, "uninstall complete — %d managed file(s) removed, %d kept:\n", removed, len(kept))
	for _, k := range kept {
		fmt.Fprintf(stdout, "  %s\n", k)
	}
	return nil
}

func readAll(p string) []byte { b, _ := os.ReadFile(p); return b }

func dropLines(blob []byte, lines []string) []byte {
	s := string(blob)
	for _, l := range lines {
		s = dropLine(s, l)
	}
	return []byte(s)
}

func bytesEqual(a, b []byte) bool { return len(a) == len(b) && string(a) == string(b) }

func dropLine(blob, want string) string {
	var keep []string
	for _, ln := range strings.Split(blob, "\n") {
		if strings.TrimSpace(ln) == strings.TrimSpace(want) {
			continue
		}
		keep = append(keep, ln)
	}
	out := strings.Join(keep, "\n")
	for strings.Contains(out, "\n\n\n") {
		out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	}
	out = strings.TrimRight(out, "\n") + "\n"
	if out == "\n" {
		return ""
	}
	return out
}

// removeManagedFile compares against the recorded base (the last upstream
// version the consumer reconciled onto), not the live embedded bytes. The
// ledger decides whether the file may be deleted at all.
func removeManagedFile(root, rel string, base []byte, tracked bool, origin string, removed *int, kept *[]string, stdout *strings.Builder) {
	dstP := filepath.Join(root, rel)
	local, err := os.ReadFile(dstP)
	if os.IsNotExist(err) {
		return
	}
	orig, hasOrig := readOriginal(root, rel)
	keep := func(reason string) {
		fmt.Fprintf(stdout, "KEPT      %s (%s)\n", rel, reason)
		*kept = append(*kept, rel+" — "+reason)
	}
	switch {
	case !tracked:
		// No recorded base at all: the file cannot be compared to anything,
		// so it is not "locally modified" — it is unattributable. Say that.
		keep("no recorded base; provenance unknown, delete manually if intended")
	case isSame(local, base):
		switch {
		case origin == originPreexisting && hasOrig:
			_ = os.WriteFile(dstP, orig, 0o644)
			fmt.Fprintf(stdout, "restored  %s (pre-install original)\n", rel)
		case origin == originCreated:
			_ = os.Remove(dstP)
			fmt.Fprintf(stdout, "removed   %s\n", rel)
			*removed++
		default:
			// preexisting with a lost original, or no recorded origin at
			// all: unknown provenance is never resolved by deleting.
			keep("provenance unknown; delete manually if intended")
		}
	case hasOrig && isSame(local, orig):
		fmt.Fprintf(stdout, "restored  %s (already at pre-install original)\n", rel)
	default:
		keep("locally modified; delete manually if intended")
	}
	_ = removeDirIfEmpty(filepath.Dir(dstP))
}

func removeAgentsBlock(root string, own *ownership, removed *int, kept *[]string, stdout *strings.Builder) {
	ap := filepath.Join(root, agentsTarget)
	raw, err := os.ReadFile(ap)
	if err != nil {
		return
	}
	content := string(raw)
	i, ej, ok := agentsSpan(content)
	if !ok {
		return
	}
	end := ej
	if end < len(content) && content[end] == '\n' {
		end++
	}
	outside := content[:i] + content[end:]
	remainder := strings.TrimSpace(outside)
	origin := own.get(kindFile, agentsTarget)

	// Creating the file does not confer ownership of every byte written into
	// it afterwards (F01). Delete only when nothing but this installer's own
	// generated header remains.
	generatedOnly := origin == originCreated && remainder == agentsCreatedHeader

	switch {
	case remainder == "":
		_ = os.Remove(ap)
		fmt.Fprintf(stdout, "removed   %s (nothing outside the block)\n", agentsTarget)
		*removed++
	case generatedOnly:
		_ = os.Remove(ap)
		fmt.Fprintf(stdout, "removed   %s (created by install; only generated boilerplate outside the block)\n", agentsTarget)
		*removed++
	default:
		body := strings.TrimRight(outside, "\n") + "\n"
		_ = writeFileAtomic(ap, []byte(body))
		fmt.Fprintf(stdout, "unmarked  %s (block removed, surrounding text preserved)\n", agentsTarget)
		*kept = append(*kept, agentsTarget+" — text outside the block preserved")
	}
	_ = removeDirIfEmpty(filepath.Dir(ap))
}

func removeDirIfEmpty(dir string) error {
	ent, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	if len(ent) != 0 {
		return nil
	}
	return os.Remove(dir)
}

func isSame(a, b []byte) bool { return bytesEqual(a, b) }

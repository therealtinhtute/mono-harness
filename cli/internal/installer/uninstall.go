package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Uninstall removes exactly the managed set. A managed file whose current
// bytes differ from both its recorded base and the embedded upstream is
// treated as locally authored work and left in place with a warning (R12,
// consumer bytes are never destroyed).
func Uninstall(root string, stdout *strings.Builder) error {
	targets, err := AllTargets()
	if err != nil {
		return err
	}
	_, baseFiles, err := loadBase(root)
	if err != nil {
		return err
	}

	for _, t := range targets {
		removeManagedFile(root, t.Dst, baseFiles[t.Dst], stdout)
	}
	removeAgentsBlock(root, stdout)

	giNow := readAll(filepath.Join(root, gitignoreTarget))
	cleaned := dropLines(giNow, gitignoreWants)
	if !bytesEqual(cleaned, giNow) {
		if len(strings.TrimSpace(string(cleaned))) == 0 {
			_ = os.Remove(filepath.Join(root, gitignoreTarget))
			fmt.Fprintf(stdout, "removed   %s\n", gitignoreTarget)
		} else {
			_ = writeFileAtomic(filepath.Join(root, gitignoreTarget), cleaned)
			fmt.Fprintf(stdout, "restored  %s (zharness entries removed)\n", gitignoreTarget)
		}
	}

	_ = os.RemoveAll(filepath.Join(root, stashDir))
	_ = saveConflicts(root, nil)
	_ = os.Remove(filepath.Join(root, manifestFile))
	_ = os.RemoveAll(filepath.Join(root, upstreamDir))
	_ = os.RemoveAll(filepath.Join(root, originalDir))
	_ = os.RemoveAll(filepath.Join(root, baseDir))
	_ = removeDirIfEmpty(filepath.Join(root, zharnessDir))

	fmt.Fprintln(stdout, "uninstall complete — consumer-owned files were never touched.")
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

func removeManagedFile(root, rel string, upstream []byte, stdout *strings.Builder) {
	dstP := filepath.Join(root, rel)
	local, err := os.ReadFile(dstP)
	if os.IsNotExist(err) {
		return
	}
	switch {
	case isSame(local, upstream):
		_ = os.Remove(dstP)
		fmt.Fprintf(stdout, "removed   %s\n", rel)
	default:
		orig, hasOrig := readOriginal(root, rel)
		if hasOrig && isSame(local, orig) {
			_ = os.WriteFile(dstP, orig, 0o644)
			fmt.Fprintf(stdout, "restored  %s (pre-install original)\n", rel)
			return
		}
		fmt.Fprintf(stdout, "KEPT      %s (locally modified; delete manually if intended)\n", rel)
	}
	_ = removeDirIfEmpty(filepath.Dir(dstP))
}

func removeAgentsBlock(root string, stdout *strings.Builder) {
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
	remainder := strings.TrimSpace(content[:i] + content[end:])
	if _, hasOrig := readOriginal(root, agentsTarget); !hasOrig {
		// the file did not exist before install: it is wholly ours
		_ = os.Remove(ap)
		fmt.Fprintf(stdout, "removed   %s (created by install)\n", agentsTarget)
	} else if remainder == "" {
		_ = os.Remove(ap)
		fmt.Fprintf(stdout, "removed   %s (nothing outside the block)\n", agentsTarget)
	} else {
		body := strings.TrimRight(content[:i]+content[end:], "\n") + "\n"
		_ = writeFileAtomic(ap, []byte(body))
		fmt.Fprintf(stdout, "unmarked  %s (block removed, surrounding text preserved)\n", agentsTarget)
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

var _ = fmt.Sprintf
var _ = filepath.Join

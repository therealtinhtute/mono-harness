package installer

import (
	"fmt"
	"strings"
)

// hunk replaces ancestor lines [start,end) with Lines.
type hunk struct {
	start, end int
	lines      []string
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// lcsCellCap bounds the quadratic LCS matrix allocated by diffHunks:
// 8M cells ≈ 64 MB of int on 64-bit — far above anything a managed doc
// legitimately reaches, far below the OOM cliff a ballooned consumer file
// would otherwise hit.
const lcsCellCap = 8_000_000

// diffHunks produces merged replacement hunks between base and other,
// absorbing change regions separated by <=3 common lines so formatting
// jitter never fabricates extra conflict sites.
func diffHunks(base, other []string) []hunk {
	n, m := len(base), len(other)
	// R4: the LCS matrix is quadratic in (n+1)*(m+1) ints. Above the cap,
	// fall back to a single whole-side hunk — bounded memory, conservative
	// merge semantics (any opposing change overlaps and conflicts rather
	// than silently merging against an approximated diff).
	if (n+1)*(m+1) > lcsCellCap {
		return []hunk{{start: 0, end: n, lines: append([]string{}, other...)}}
	}
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if base[i] == other[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	type span struct{ bs, be, os_, oe int }
	var spans []span
	i, j := 0, 0
	for i < n || j < m {
		if i < n && j < m && base[i] == other[j] {
			i++
			j++
			continue
		}
		bs, o0 := i, j
		for (i < n || j < m) && !(i < n && j < m && base[i] == other[j]) {
			switch {
			case i == n:
				j++
			case j == m:
				i++
			case lcs[i+1][j] >= lcs[i][j+1]:
				i++
			default:
				j++
			}
		}
		spans = append(spans, span{bs, i, o0, j})
	}
	const gap = 3
	var out []hunk
	k := 0
	for k < len(spans) {
		cur := spans[k]
		be, oe := cur.be, cur.oe
		for k+1 < len(spans) && spans[k+1].bs-be <= gap {
			k++
			be, oe = spans[k].be, spans[k].oe
		}
		out = append(out, hunk{cur.bs, be, append([]string{}, other[cur.os_:oe]...)})
		k++
	}
	return out
}

func overlapSpan(a, b hunk) bool {
	if a.start == b.start && a.end == b.end {
		return true // identical placement (incl. zero-width insertions) is ambiguous
	}
	return a.start < b.end && b.start < a.end
}

// renderCluster renders one side's version of ancestor region [u0,u1) with
// an ordered cluster of hunks applied.
func renderCluster(L []string, u0, u1 int, hs []hunk) []string {
	out := make([]string, 0, u1-u0)
	pos := u0
	for _, h := range hs {
		if h.start > pos {
			out = append(out, L[pos:minInt(h.start, len(L))]...)
			pos = h.start
		}
		out = append(out, h.lines...)
		if h.end > pos {
			pos = h.end
		}
	}
	if pos < u1 {
		out = append(out, L[pos:minInt(u1, len(L))]...)
	}
	return out
}

func writeLines(sb *strings.Builder, lines []string) {
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
}

func emitBaseRange(sb *strings.Builder, L []string, from, to int) int {
	f := minInt(from, len(L))
	toC := minInt(to, len(L))
	for f < toC {
		sb.WriteString(L[f])
		sb.WriteString("\n")
		f++
	}
	return maxInt(f, minInt(to, len(L)))
}

// threeWay merges two edits of a common ancestor (diff3). Returns the merged
// text; conflicted regions are left inline between explicit markers.
func threeWay(base, local, update string) (string, bool) {
	L := splitLines(base)
	HL := diffHunks(L, splitLines(local))
	HN := diffHunks(L, splitLines(update))

	var sb strings.Builder
	conflict := false
	cur := 0
	il, iu := 0, 0

walk:
	for {
		// A hunk is only consumed once applied; a zero-width insertion
		// exactly at cur is still pending and must stay selectable.
		for il < len(HL) && (HL[il].end < cur || (HL[il].end == cur && HL[il].start < cur)) {
			il++
		}
		for iu < len(HN) && (HN[iu].end < cur || (HN[iu].end == cur && HN[iu].start < cur)) {
			iu++
		}
		var hl, hn *hunk
		sl, sn := int(^uint(0)>>1), int(^uint(0)>>1)
		if il < len(HL) && HL[il].start >= cur {
			hl = &HL[il]
			sl = HL[il].start
		}
		if iu < len(HN) && HN[iu].start >= cur {
			hn = &HN[iu]
			sn = HN[iu].start
		}

		switch {
		case hl != nil && hn != nil && overlapSpan(*hl, *hn):
			u0 := minInt(sl, sn)
			u1 := maxInt(hl.end, hn.end)
			lcl := []hunk{*hl}
			up := []hunk{*hn}
			il++
			iu++
			// Absorb any later hunk from either side that reaches into
			// the cluster; otherwise one straddling cur leaves the
			// walker with nothing selectable and it spins forever.
			for {
				absorbed := false
				for il < len(HL) && HL[il].start < u1 {
					u1 = maxInt(u1, HL[il].end)
					lcl = append(lcl, HL[il])
					il++
					absorbed = true
				}
				for iu < len(HN) && HN[iu].start < u1 {
					u1 = maxInt(u1, HN[iu].end)
					up = append(up, HN[iu])
					iu++
					absorbed = true
				}
				if !absorbed {
					break
				}
			}
			cur = emitBaseRange(&sb, L, cur, u0)
			localSeg := renderCluster(L, u0, u1, lcl)
			upSeg := renderCluster(L, u0, u1, up)
			if equalLines(localSeg, upSeg) {
				writeLines(&sb, localSeg)
			} else {
				conflict = true
				sb.WriteString(conflictOpenTag + fmt.Sprintf(" near ancestor line %d\n", u0+1))
				writeLines(&sb, upSeg)
				sb.WriteString(conflictSepTag + "\n")
				writeLines(&sb, localSeg)
				sb.WriteString(conflictCloseTag + "\n")
			}
			cur = maxInt(cur, u1)
		case hl != nil && sl <= sn:
			cur = emitBaseRange(&sb, L, cur, hl.start)
			writeLines(&sb, hl.lines)
			cur = maxInt(cur, hl.end)
			il++
		case hn != nil:
			cur = emitBaseRange(&sb, L, cur, hn.start)
			writeLines(&sb, hn.lines)
			cur = maxInt(cur, hn.end)
			iu++
		default:
			// nothing selectable at/after cursor (defensive; cluster
			// absorption above prevents straddlers): flush remainder
			break walk
		}
	}
	for ; cur < len(L); cur++ {
		sb.WriteString(L[cur])
		sb.WriteString("\n")
	}

	return sb.String(), conflict
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a < b {
		return b
	}
	return a
}
func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k] != b[k] {
			return false
		}
	}
	return true
}

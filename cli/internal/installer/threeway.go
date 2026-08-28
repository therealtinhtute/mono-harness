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

// diffHunks produces merged replacement hunks between base and other,
// absorbing change regions separated by <=2 common lines so formatting
// jitter never fabricates extra conflict sites.
func diffHunks(base, other []string) []hunk {
	n, m := len(base), len(other)
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

// renderRegion renders one side's version of ancestor region [u0,u1).
func renderRegion(L []string, u0, u1 int, h hunk) []string {
	clampLo, clampHi := minInt(u0, len(L)), minInt(u1, len(L))
	hs, he := minInt(h.start, len(L)), minInt(h.end, len(L))
	pre := L[clampLo:minInt(hs, clampHi)]
	post := L[minInt(he, clampHi):clampHi]
	out := make([]string, 0, len(pre)+len(h.lines)+len(post))
	out = append(out, pre...)
	out = append(out, h.lines...)
	out = append(out, post...)
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

	for {
		for il < len(HL) && HL[il].end <= cur {
			il++
		}
		for iu < len(HN) && HN[iu].end <= cur {
			iu++
		}
		if il >= len(HL) && iu >= len(HN) {
			break
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
			localSeg := renderRegion(L, u0, u1, *hl)
			upSeg := renderRegion(L, u0, u1, *hn)
			cur = emitBaseRange(&sb, L, cur, u0)
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
			il++
			iu++
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
			// nothing selectable at/after cursor: flush remainder
			break
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

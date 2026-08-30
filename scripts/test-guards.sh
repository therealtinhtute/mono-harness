#!/usr/bin/env bash
# Fixture tests for the ZGUARD-CORE guard (guard-v3 plan R1/R2; keeps the v2
# hardening semantics honest). Extracts the core exactly like CI does, then
# asserts observable guard behavior. Run: bash scripts/test-guards.sh
set -u
cd "$(dirname "$0")/.."

GUARD=$(mktemp)
awk '$0=="# ZGUARD-CORE-BEGIN"{on=1;next} $0=="# ZGUARD-CORE-END"{on=0} on' \
	scripts/install-git-hooks.sh > "$GUARD"
grep -q '^zharness_guard_entries_of_file()' "$GUARD" || {
	echo "FAIL - guard core extraction failed"; exit 1; }
grep -q '^zharness_guard_at_most_one_active_plan()' "$GUARD" || {
	echo "FAIL - R5 one-plan guard missing from core"; exit 1; }
# shellcheck disable=SC1090
source "$GUARD"

pass=0
fail=0
ok() { pass=$((pass + 1)); echo "  ok  - $1"; }
bad() { fail=$((fail + 1)); echo "  FAIL - $1"; }

# S1 / R1: a verdict token quoted in a sub-bullet cannot shadow the
# entry's own first-line verdict (the guard-v3 first-line rule).
v=$(zharness_anchored_verdict '- `2026-08-28T00:00:00Z` — gate verdict `APPROVED` (mode: gate)
  - quoted prior review: verdict: REQUEST_CHANGES
  - `true`')
[ "$v" = "APPROVED" ] &&
	ok "S1 sub-bullet prose does not shadow first-line APPROVED" ||
	bad "S1 expected APPROVED, got '$v'"

# R1 converse: a genuine first-line REQUEST_CHANGES still wins.
v=$(zharness_anchored_verdict '- `2026-08-28T00:00:00Z` — review verdict: REQUEST_CHANGES
  - quoted earlier: verdict: APPROVED
  - `false`')
[ "$v" = "REQUEST_CHANGES" ] &&
	ok "S1 genuine first-line REQUEST_CHANGES respected" ||
	bad "S1 converse expected REQUEST_CHANGES, got '$v'"

# R1 grammar: both verdict token forms are recognized.
v=$(zharness_anchored_verdict '- `2026-08-28T00:00:00Z` — gate verdict: `APPROVE_WITH_REQUESTS`')
[ "$v" = "APPROVE_WITH_REQUESTS" ] &&
	ok "R1 colon form recognized" ||
	bad "R1 colon form expected APPROVE_WITH_REQUESTS, got '$v'"

v=$(zharness_anchored_verdict '- `2026-08-28T00:00:00Z` — gate verdict `APPROVED` (no colon)')
[ "$v" = "APPROVED" ] &&
	ok "R1 backtick form recognized" ||
	bad "R1 backtick form expected APPROVED, got '$v'"

# R1: no verdict token on the first line -> empty (entry ignored), even
# when the body prose mentions one.
v=$(zharness_anchored_verdict '- `2026-08-28T00:00:00Z` — narrative entry
  - sub-bullet mentions verdict: APPROVED')
[ -z "$v" ] &&
	ok "R1 verdictless first line yields no verdict despite body prose" ||
	bad "R1 expected empty verdict, got '$v'"

# S2 / R2: an entry without a leading timestamp is still split out as its
# own entry; indented lines continue the current entry.
tmp=$(mktemp -d)
mkdir -p "$tmp/out"
cat > "$tmp/plan.md" <<'EOF'
---
lane: high-risk
---

## Validation

- undated APPROVED entry for the fixture, verdict `APPROVED`
  - `true`

- `2026-08-28T00:00:00Z` — dated sibling entry
  - `true`

## Current State and Next Action

- after validation
EOF
zharness_dump_entries "$tmp/plan.md" "$tmp/out"
count=$(ls "$tmp/out" | wc -l | tr -d '[:space:]')
[ "$count" -eq 2 ] &&
	ok "S2 undated entry split as its own entry (2 entries)" ||
	bad "S2 expected 2 entries, got $count"

if [ -f "$tmp/out/e0001.txt" ] && head -1 "$tmp/out/e0001.txt" | grep -q "^- undated"; then
	ok "S2 undated entry is first, full text preserved"
else
	bad "S2 undated entry content wrong"
fi

# S3 / R6 parity: fail-case rejects, pass-case accepts, same-session on a
# high-risk lane rejects. Fixtures mirror the repository's real entry
# grammar (verdict `APPROVED`, no colon) and carry frontmatter.
cat > "$tmp/old.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`
EOF
cat > "$tmp/new-fail.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- `2026-08-28T00:00:00Z` — new gate verdict `APPROVED` with a sabotaged proof
  - `echo sabotaged-proof && false`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-fail.md" 2>/dev/null; then
	bad "S3 fail-case must reject"
else
	ok "S3 fail-case (sabotaged proof) rejects"
fi

cat > "$tmp/new-pass.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- `2026-08-28T00:00:00Z` — new gate verdict `APPROVED` with a clean proof
  - `true`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-pass.md" 2>/dev/null; then
	ok "S3 pass-case accepts"
else
	bad "S3 pass-case must accept"
fi

cat > "$tmp/new-judge.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- `2026-08-28T00:00:00Z` — gate verdict `APPROVED` judged same-session
  - judge: `same-session`
  - `true`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-judge.md" 2>/dev/null; then
	bad "S3 same-session on high-risk lane must reject"
else
	ok "S3 same-session judge on high-risk lane rejects"
fi

cat > "$tmp/new-full-ss.md" <<'EOF'
---
lane: normal
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- [2026-08-30 (initiative close)] mode `full` verdict `APPROVED` — judge: `same-session`
  - `true`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-full-ss.md" 2>/dev/null; then
	bad "H3 mode full + same-session must reject"
else
	ok "H3 mode full + same-session rejects"
fi

cat > "$tmp/new-gate-ss.md" <<'EOF'
---
lane: normal
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session`
  - `true`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-gate-ss.md" 2>/dev/null; then
	ok "H3 mode gate + same-session on normal lane accepts"
else
	bad "H3 mode gate + same-session must accept"
fi

cat > "$tmp/new-gate-quote-full.md" <<'EOF'
---
lane: normal
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session`
  - prior close quoted: mode: full
  - `true`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-gate-quote-full.md" 2>/dev/null; then
	ok "H3 quoted mode-full in a gate sub-bullet still accepts"
else
	bad "H3 quoted mode-full in body must not reject a gate entry"
fi

# S2/R2 end-to-end: a newly added UNDATED APPROVED entry is visible to the
# whole guard flow and its proofs are re-executed (fail-case rejects).
cat > "$tmp/new-undated-fail.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — older entry, already on record
  - `true`

- undated entry, verdict `APPROVED`, invisible to the v2 guard
  - `echo undated-sabotage && false`
EOF
if zharness_guard_entries_of_file p.md "$tmp/old.md" "$tmp/new-undated-fail.md" 2>/dev/null; then
	bad "S2 undated APPROVED entry must be re-executed (fail-case must reject)"
else
	ok "S2 undated APPROVED entry is now guarded (fail-case rejects)"
fi

# S4 portability: the pre-commit hook's shebang is `#!/bin/bash`, which on
# macOS is bash 3.2 — no associative arrays, and a hex array subscript there is
# evaluated as arithmetic and kills the shell. Re-run the two decisive cases
# under a legacy bash if one is present, so the guard core can never regress
# into a bash-4-only construct. Old side carries an entry in both cases: that
# is what populates the hash set and what used to blow up.
LEGACY_BASH=""
for cand in /bin/bash /usr/local/bin/bash-3.2; do
	if [ -x "$cand" ] && case "$("$cand" -c 'echo ${BASH_VERSINFO[0]}')" in 3) true ;; *) false ;; esac; then
		LEGACY_BASH="$cand"
		break
	fi
done

if [ -n "$LEGACY_BASH" ]; then
	cat > "$tmp/legacy-old.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — an entry already on record
  - `true`
EOF
	cat > "$tmp/legacy-pass.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — an entry already on record
  - `true`

- `2026-08-28T00:00:00Z` — honest work, verdict `APPROVED`
  - `true`
EOF
	cat > "$tmp/legacy-fail.md" <<'EOF'
---
lane: high-risk
---

## Validation

- `2026-08-27T00:00:00Z` — an entry already on record
  - `true`

- `2026-08-28T00:00:00Z` — verdict `APPROVED` with a sabotaged proof
  - `echo legacy-sabotage && false`
EOF
	legacy_run() {                            # <new-file>; echoes the guard exit code
		"$LEGACY_BASH" -c '
			source "$1" || exit 90
			zharness_guard_entries_of_file p.md "$2" "$3" >/dev/null 2>&1
		' _ "$GUARD" "$tmp/legacy-old.md" "$1"
		echo $?
	}

	rc=$(legacy_run "$tmp/legacy-pass.md")
	[ "$rc" = 0 ] &&
		ok "S4 legacy bash ($("$LEGACY_BASH" -c 'echo $BASH_VERSION')): honest entry accepted" ||
		bad "S4 legacy bash: honest entry must be accepted, guard exited $rc"

	rc=$(legacy_run "$tmp/legacy-fail.md")
	[ "$rc" != 0 ] &&
		ok "S4 legacy bash: sabotaged entry rejected" ||
		bad "S4 legacy bash: sabotaged entry must be rejected"
else
	echo "  skip - S4 legacy bash 3.x not found on this machine"
fi

# S5 / #85: entry hashes must not depend on trailing blank lines. When
# `## Validation` is the final section, the last entry ends at EOF; appending a
# new entry below gives it a trailing blank line. Before the fix that changed
# its hash and R2 re-executed an entry nobody had touched. The old-side proof
# carries an observable side effect so a re-run is visible in the output.
cat > "$tmp/tail-old.md" <<'EOF'
---
lane: normal
---

## Validation

- `2026-08-27T00:00:00Z` — entry already on record, verdict `APPROVED`
  - `echo MARKER-OLD-ENTRY-RERUN`
EOF
cat > "$tmp/tail-new.md" <<'EOF'
---
lane: normal
---

## Validation

- `2026-08-27T00:00:00Z` — entry already on record, verdict `APPROVED`
  - `echo MARKER-OLD-ENTRY-RERUN`

- `2026-08-28T00:00:00Z` — newly appended entry, verdict `APPROVED`
  - `true`
EOF
s5out=$(zharness_guard_entries_of_file p.md "$tmp/tail-old.md" "$tmp/tail-new.md" 2>&1)
s5rc=$?
[ "$s5rc" -eq 0 ] &&
	ok "S5 appending to a trailing Validation section is accepted" ||
	bad "S5 append must be accepted, guard exited $s5rc"

if printf '%s' "$s5out" | grep -q MARKER-OLD-ENTRY-RERUN; then
	bad "S5 untouched entry was re-executed (#85: trailing blank line changed its hash)"
else
	ok "S5 untouched entry is not re-executed when a sibling is appended"
fi

# R5: at most one non-empty docs/plans/active/*.md. Zero is idle. Empty files
# do not count. No `local -A`.
r5=$(mktemp -d)
mkdir -p "$r5/docs/plans/active"
if zharness_guard_at_most_one_active_plan "$r5" 2>/dev/null; then
	ok "R5 zero active plans accepted"
else
	bad "R5 zero must accept"
fi
printf 'one\n' > "$r5/docs/plans/active/one.md"
if zharness_guard_at_most_one_active_plan "$r5" 2>/dev/null; then
	ok "R5 one active plan accepted"
else
	bad "R5 one must accept"
fi
printf 'two\n' > "$r5/docs/plans/active/two.md"
if zharness_guard_at_most_one_active_plan "$r5" 2>/dev/null; then
	bad "R5 two active plans must reject"
else
	ok "R5 two active plans rejected"
fi
: > "$r5/docs/plans/active/two.md"
if zharness_guard_at_most_one_active_plan "$r5" 2>/dev/null; then
	ok "R5 empty second file does not count"
else
	bad "R5 empty second file must not reject"
fi
rm -rf "$r5"

rm -rf "$tmp" "$GUARD"
echo
echo "guards: $pass passed, $fail failed"
[ "$fail" -eq 0 ]

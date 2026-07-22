# PLAN: Slim Playbooks S3 — `zharness resume` full-recap render + watzup slim

**Date:** 2026-07-22
**Origin:** `.kit/plans/2026-07-21-slim-playbooks/PLAN.md` Phase S3 (locked). Continues S1 (scaffold, done) + S2 (`next`, done — see `.kit/reports/check/20260722-slim-playbooks-s2-gate.md`).
**Lane:** high-risk (same as S1/S2 — public CLI contract change, embedded-playbook rewrite).

## Scope decision (stated up front, not silently assumed)

The master plan's own wording is ambiguous between two designs: "renders the harness-state block... fully, [git/WIP] stays agent-gathered" (implies a harness-only text snippet) vs. "Default human output = the Recap format... forbidden-phrase-safe **by construction**" (implies the CLI renders the *entire* recap, which only makes sense if the agent feeds git/WIP facts into it). Presented both plus a middle ground to the user via `AskUserQuestion`; **user chose full-recap render**. This plan proceeds on that basis:

- `zharness resume` gains a `--facts '<json>'` flag. The agent still gathers git state itself (Steps 1/3/4 in watzup.md are unchanged — git commands, theme extraction, WIP-diff reading all require an actual `git` process and LLM judgment the CLI doesn't have) and still decides *what* counts as a risk and what the one next action is (Step 5/6 judgment stays agent-side). But once those facts are decided, the agent hands them to `zharness resume --facts` as structured JSON, and the CLI — not the agent — builds and prints the final Vietnamese Recap text (Title/Trạng thái/Context/Thay đổi/Risks/Next), enforcing every current Output-Contract rule (forbidden phrases, risk-table columns, severity ladder, 25-line empty-state form, drift-recovery-overrides-next-action) as Go validation/construction instead of an agent self-check paragraph.
- `resume` still never runs git itself — "no git access" is preserved literally; the agent is the one gathering git facts, the CLI only formats them. This mirrors `next`'s carve-out pattern (CLI encodes what's deterministic; git-dependent judgment stays agent-side) rather than contradicting it.
- `--json` behavior is **unchanged** (still the existing machine shape, no facts needed). `--facts` and `--json` are mutually exclusive (`invalid_arguments` if both set). Calling `resume` with neither flag keeps today's one-line scripting output (`phase=... readiness=... drift=N`) — no behavior change for existing scripted callers.
- **CONTRACT.md's `resume` entry is not updated in this phase** — same precedent as S1/S2, neither of which documented `scaffold`/`next` in `cli/docs/CONTRACT.md` either. Tracked as a known pre-existing gap (all three phases), not new scope creep introduced here.

## Steps

1. **Domain**: add `RecapFacts`/`RecapRisk` types + a pure `RenderRecap(view ResumeView, facts RecapFacts) (string, error)` function in `internal/application/resume.go` (or a new `recap.go` alongside it — keep `Resume()` itself untouched).
   - `RecapFacts`: `branch string`, `ahead int`, `behind int`, `uncommitted_files int`, `uncommitted_adds int`, `uncommitted_dels int`, `handoff_summary string`, `changes []string` (committed themes, agent caps at 3), `wip []string` (agent caps at what's significant), `risks []RecapRisk` (`{risk, severity, action}`), `next_action string`.
   - Render logic:
     - **Empty-state branch**: `ahead==0 && behind==0 && uncommitted_files==0 && handoff_summary=="" && view.Readiness=="clean"` → emit the exact 2-line empty-state form, ignore all other facts fields.
     - **Forbidden-phrase check (by construction)**: scan every free-text field (`branch`, `handoff_summary`, each `changes`/`wip` entry, each risk's 3 fields, `next_action`) for the fixed forbidden substrings (git commands, `--stat`/`--oneline`/etc., `Quality: N/10` / `Score: N/10` pattern, the literal `"git "` trailing-space substring) ported verbatim from watzup.md's current Section 1. Any hit → `domain.ValidationError{Code: "forbidden_phrase", ...}` naming the field and phrase — the agent must fix its facts and retry, not the CLI silently stripping/rendering bad text.
     - **Severity validation**: each risk's `severity` must be one of `cao|vừa|thấp` → `invalid_severity` otherwise.
     - **Drift-override**: if `view.Readiness == "drifted"`, the rendered `Next:` line is `view.Drift[0].Recovery` verbatim — `facts.NextAction` is ignored in this case (matches the existing "print the first drift entry's recovery field verbatim — do not paraphrase" rule, now enforced instead of merely instructed).
     - **No-harness branch**: `view.Readiness == "no-harness"` renders the Context bullet as "Không có handoff" + a `.kit/` presence note is NOT inferable by the CLI (it doesn't check `.kit/planning/`) — agent still supplies this via `handoff_summary`/a `legacy_kit_present bool` fact if it wants the Example-4-style phrasing; keep the render generic ("Readiness: no-harness") if that fact is absent rather than guessing.
     - **List caps**: `changes`+`wip` combined render as the Thay đổi section, capped at 5 bullets (truncate extras rather than erroring — the agent is expected to already cap, this is a defensive backstop for the 25-line target).
     - Risk table + Title format + section ordering: port directly from watzup.md's current Section 3/4/5 (exact formats), now as Go string-building instead of prose rules.
   - Verify: unit — one test per current format-contract rule: title format, risk table columns/severity rejection, forbidden-phrase rejection (git command substring, `Quality: 8/10` pattern), drift overrides next_action, empty-state exact output, Thay đổi capped at 5, no-harness branch.

2. **Interface**: `internal/interfaces/resume.go` — add `--facts` string flag to the existing `resume` command.
   - `--facts` + `--json` both set → `invalid_arguments` error.
   - `--facts` set, `--json` not: parse JSON → `RecapFacts` (`facts_malformed` on bad JSON); call the existing `Resume(db, version)` (or the existing no-harness short-circuit) to get `ResumeView` unchanged; call `RenderRecap(view, facts)`; print its result verbatim to stdout. Any `ValidationError` from step 1 surfaces the same way `next`/`scaffold` map `domain.ValidationError` today (`mapValidationError`).
   - Neither flag set: unchanged existing one-line output (no regression).
   - Verify: command-output — `go build && zharness resume --facts '{...}'` smoke test against a few real facts blobs (clean/in-progress/drifted cases using this repo's actual current `resume --json` state) plus a scratch no-harness fixture.

3. **Integration test**: `RenderRecap` exercised against real `ResumeView` fixtures for all four readiness states (`clean`, `in-progress`, `drifted`, `no-harness`) built the same way `next_test.go`'s `freshDB`+`t.Chdir`+`t.TempDir` helpers already do — satisfies the plan's `integration` proof-class cell without a separate fixture format.

4. **Playbook slim**: rewrite `cli/docs/embedded/playbooks/watzup.md`.
   - **Keep, unchanged**: Steps 1 (Branch State), 3 (Committed Work Summary), 4 (WIP Analysis) — these require actual `git` output + LLM judgment the CLI still can't do. Step 5 (Risk Assessment) keeps its judgment criteria table (deciding *which* signals count as risks is still agent-side) but drops the "map drift type → recovery" sub-instruction (now automatic in `RenderRecap`). Step 6 (Next Action) keeps its decision table (still agent judgment) but drops the "drifted state prints recovery verbatim" instruction (now enforced by the CLI, not a rule the agent has to remember).
   - **Rewrite Step 2**: from "load harness state, extract fields, manually build Trạng thái/Context text" → "assemble the facts JSON (branch/ahead/behind/uncommitted counts from Step 1, themes from Step 3, wip from Step 4, risks from Step 5, next_action from Step 6) and call `zharness resume --facts '<json>'` exactly once; print its stdout verbatim as the final answer — do not reformat, re-derive, or add text around it." This is the *only* call to `resume` (matches the existing "call resume --json exactly once" invariant, now for `--facts` instead).
   - **Delete wholesale**: the "Output Contract" section (current lines 106-208 — Forbidden Phrases, Allowed Vocabulary, §2.5 Readiness States and Recovery, Title Format, Risk Table Contract, Output Layout, Self-Check) — all of it is now enforced by `RenderRecap`, not agent prose. Replace with a 3-4 line note: "Formatting, forbidden-phrase safety, the risk-table shape, and the drifted-state recovery override are enforced by `zharness resume --facts` itself — verify by reading its exit code/error, not by manually inspecting the printed text against a rule list."
   - **Delete wholesale**: all 4 worked "Examples" (current lines 210-291) — they existed to teach manual formatting, which no longer exists. Replace with ONE compact example: a sample `--facts` JSON blob + the exact stdout it produces (in-progress case, mirrors current Example 1's scenario so the illustration isn't lost, just no longer needed as a formatting guide).
   - **Shrink** "Readiness State" section (current lines 95-104) to 2-3 lines: readiness still comes from `resume` only, never independently derived; it's now baked into the `--facts` render rather than a value the agent copies into hand-built text.
   - Update "Command Reference" section: `zharness resume --facts '<json>' ` replaces the old `zharness resume --json` reference as the primary call; keep `--json` documented as the machine-readable fallback.
   - Target: watzup.md lands under 160 lines (master plan's own target), down from 308.
   - Verify: `gofmt -l .`, `go build ./...`, `go vet ./...`, `go test ./...` all pass after the embed changes.

5. **Gate**: run `check full` on the S3 diff (same pattern as S1-W3/S2), persist report, skip `check record` again (no RUN artifact for this informal track — same as S1/S2).

## Non-scope

- `CONTRACT.md`'s `resume` entry — not updated this phase (pre-existing gap shared with `scaffold`/`next`, not new).
- `resume` gaining actual git access — still zero; git-gathering stays 100% agent-side, only the *text assembly* moves to Go.
- Phase 4 (single-source-playbooks, projects final text + drift-guard test) — separate, comes after all of S1-S3 per the master plan's sequencing.
- Any change to `--json`'s existing machine shape.

## Rollback

Additive flag (`--facts`) + playbook-text-only change, same as S1/S2 — revertable independently via git, no schema change. `--json` and the no-flag default paths are untouched, so any existing scripted caller of `resume` is unaffected regardless of rollback state.

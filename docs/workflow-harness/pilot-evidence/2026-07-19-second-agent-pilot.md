# Second-Agent Pilot — Codex CLI, 2026-07-19

Evidence for `agent-pilot` (phase 6 of `playbook-authoring → agent-pilot`), which exists to test R9 (agent-agnosticism): *"any coding agent — not just Claude Code — should be able to execute the full lifecycle by reading `AGENTS.md` + docs and calling `zharness`, with zero SKILL.md exposure."*

**R9 acceptance criterion (verbatim, agent-pilot-CONTEXT.md):** "completed without reading any SKILL.md, `zharness validate --json` passes on the produced chain."

## Verdict

**NO-GO on the literal R9 gate.** `zharness validate --json` on the produced chain returns `valid: false` (5 findings, all traceable to one root cause — see Findings). The "zero SKILL.md reads" half of R9 **passed** cleanly.

This is a NO-GO on the acceptance gate **and** a strong positive result for the underlying thesis (agent-agnosticism). Both are true and both matter:

- The docs (`AGENTS.md` → `.kit/docs/{AUTHORITY,CONTEXT_RULES}.md` → `playbooks/*.md`) were sufficient, unaided, for a genuinely cold non-Claude agent to: pass the version gate, self-recover from a real broken-link bug it found on its own, correctly classify and scope the task, lock a spec, run intake, apply the scope guard to choose simple mode, write working code with passing tests, produce a check report in the correct template, attempt the correct DB-registration commands, and — critically — correctly self-diagnose the resulting failure as a harness defect rather than its own error, then report that honestly instead of fabricating success.
- It failed the gate for one specific, well-understood, reproducible harness bug (filed as [#38](https://github.com/therealtinhtute/skills/issues/38)) in the run-registration data model for simple mode — not because the docs were unclear or insufficient, and not because of any Codex-specific limitation.
- Confirmed separately (see "Baseline: does validate ever pass?" below): `zharness validate --json` does not pass clean on this repo's own **Claude-produced** chain either, for an overlapping reason (the check-side counterpart of the same simple/ad-hoc-mode gap, already on record in the thin-triggers check report). The gate is not currently achievable by any agent — this is a pre-existing structural gap in simple mode's DB contract, not something the pilot introduced or that is specific to a second agent.

## Setup

- **Runtime:** `codex exec` (Codex CLI), the only genuinely independent non-Claude coding-agent runtime available on this machine (Cursor unavailable). Non-interactive, single-shot invocation:
  `codex exec -C <scratch-dir> -s workspace-write --skip-git-repo-check --json -o last-message.txt - < prompt.txt > transcript.jsonl 2> stderr.log`
- **Caveat, disclosed and accepted by the user before running:** the authenticated `codex` install carries account-level personalization (`~/.codex/memories_1.sqlite`, `goals_1.sqlite`, etc.) that is not removable via `--ignore-user-config` (confirmed: persona leakage into a smoke-test response persisted even with that flag) and no local-model isolation path was available (`--oss` needs ollama/lmstudio, neither installed). This affects *tone*, not task competence — see "Why this doesn't invalidate the pilot" below.
- **Scratch target:** fresh git repo, `zharness init`-scaffolded, seeded only with a `README.md` and the (bugged) root `AGENTS.md` shim.
- **Protocol prompt** (the entirety of what Codex was given — no harness-mechanics coaching, no mention of `zharness`/phase/check/trace/playbook anywhere in the prompt):

  > This repository has an AGENTS.md file at its root. Read it first, and follow whatever process it directs for handling work in this repo. Then complete the following task.
  >
  > \# Task
  >
  > Add a small Python module `textutils.py` at the repo root with a function `slugify(text: str) -> str` that converts a string into a URL-safe slug: lowercase the input; strip leading/trailing whitespace; collapse runs of whitespace and punctuation into single hyphens; remove any character that isn't alphanumeric or a hyphen; no leading/trailing hyphens in the result. Add a test file (`test_textutils.py`) covering at least: a normal sentence with spaces, a string with punctuation, a string that's already hyphenated. All tests must pass before you're done.
  >
  > Work autonomously end-to-end. If you hit a product decision only a human can make about the task itself, ask me directly.

## Why this doesn't invalidate the pilot

The account-level persona layer is a tone/wrapper concern, not a knowledge-injection concern, for this test:

- Codex had to **discover** the repo's process cold — it hit the real broken root-`AGENTS.md` link (issue [#37](https://github.com/therealtinhtute/skills/issues/37)) on its very first navigation step and self-recovered via `rg --files` search, unaided, to find `.kit/docs/AUTHORITY.md` and `.kit/docs/CONTEXT_RULES.md`.
- Nothing in the transcript shows foreknowledge of `zharness`, the playbook structure, or this repo's conventions — every command sequence traces directly to something it just read.
- The FK bug was **discovered by failing**, not known in advance, and Codex's own diagnosis of it (see its check report below) is independently correct and matches this doc's root-cause analysis.
- Zero project-specific or `zharness`-specific leakage appeared anywhere in the transcript.

## Command trail (abridged, full order preserved)

1. `zharness --version` — version gate pass
2. Read `AGENTS.md` at root → attempted `AUTHORITY.md`, `CONTEXT_RULES.md` at root → **exit 1, files don't exist** (issue #37, hit live)
3. Self-recovered: `rg --files -g 'AUTHORITY.md' -g 'CONTEXT_RULES.md' -g 'playbooks/**' ...` → found real paths under `.kit/docs/`
4. `zharness --help`, read `README.md`, `git status --short --branch`
5. `rg --files -uu .kit` (full scaffold discovery)
6. Read `.kit/docs/AUTHORITY.md`, `.kit/docs/CONTEXT_RULES.md`
7. Read `.kit/docs/playbooks/brainstorm.md`
8. Generated a ULID inline (`python3` one-liner), locked a spec: `zharness init --json` → `{"db_path":".kit/harness.db","schema_version":2,"status":"exists"}`; `zharness intake --type new-spec --summary "Add a dependency-free Python slugify utility with tests" --lane tiny --json` → `{"id":"01KXWZSQ904VSZF11F5CB7MQG1"}`
9. Read `.kit/docs/playbooks/work.md`
10. Chose **simple mode** correctly per the scope guard (2 files, ~10 lines — well under the 5-file/100-line threshold)
11. Attempted run registration: `zharness db changeset apply .kit/changesets/01KXWZTQTSWKT01H1Q64ZDFYWQ.changeset.jsonl --json` →
    `{"error":{"code":"changeset_malformed","message":"db changeset apply: changeset_malformed: insert runs: constraint failed: FOREIGN KEY constraint failed (787)"}}`
12. Investigated: `zharness query phases --json` → `[]`; inspected schema directly (`sqlite3`/`python3 sqlite3` — DB is `.kit/harness.db`, path guesses first failed, then succeeded): `runs.story_slug TEXT NOT NULL REFERENCES stories(slug)`, confirming no `stories` row exists to reference
13. Retried the changeset with a modified `story_slug` → different failure: `{"error":{"code":"changeset_malformed","message":"db changeset apply: changeset_malformed: update meta: constraint failed: FOREIGN KEY constraint failed (787)"}}`
14. Wrote `textutils.py` and `test_textutils.py`; ran `python3 -m unittest -v test_textutils.py` → **3/3 pass**
15. Cleanup: tried `trash` on a leftover changeset file → sandbox blocked it → self-corrected by moving artifacts to `/private/tmp/agent-pilot-run-recovery-20260719-1750/` instead of forcing deletion. **Never used `rm`, never touched files outside its own scratch output.**
16. Read `.kit/docs/playbooks/check.md`
17. `zharness audit --json`, `python3 -B -m unittest discover -v` (full 3/3 pass, re-confirmed)
18. `zharness check record --verdict APPROVE_WITH_REQUESTS --run-id 01KXWZTQTSWKT01H1Q64ZDFYWQ ... --json` →
    `{"error":{"code":"unknown_run_id","message":"check record: run_id 01KXWZTQTSWKT01H1Q64ZDFYWQ not found"}}` (cascading from #13's unregistered run)
19. Wrote a check report by hand (`.kit/reports/check/20260719-1751-textutils.md`) with an accurate self-diagnosis (quoted verbatim below) and an honest `harness_verdict: not recorded`
20. **Zero commits.** `git log --oneline --all` in the scratch dir shows only the human-made init commit.

## Codex's own diagnosis (from its check report, unedited)

> The harness cannot register a phase-less simple run because `runs.story_slug` requires a story, while the simple-mode work playbook creates no story.
> ...
> Harness contract mismatch prevents durable run and check registration for simple mode; application code is unaffected.
> ...
> Fix the harness simple-run contract separately; the requested Python files are ready.

This is independently correct and matches this doc's root-cause finding exactly — arrived at with zero coaching, purely from reading the failure and the schema.

## Findings

### F1 — `work` simple mode cannot register a run in the harness DB (blocking, filed)

**Root cause:** `runs.story_slug TEXT NOT NULL REFERENCES stories(slug)`. `work.md`'s Execution Loop step 2 instructs every invocation — the step is inline in a section headered "full mode" but explicitly carves out simple-mode behavior ("In simple mode, the slug comes from the prompt or brainstorm file") — to register the run via a changeset whose `story_slug` field has no corresponding `stories` row in simple mode, because simple mode's only artifact command (`zharness intake`) creates a SPEC-level record, not a story, and `zharness story` is explicitly phase-scoped ("Record a new story **(phase)**"). 100% reproducible for every simple-mode run. Cascades into `check record` failing (`unknown_run_id`) and `validate` reporting `stale_pointer` on both the run and check links.

**Severity:** Major — blocks R9's literal acceptance bar; does not corrupt or block the actual code-writing workflow.
**Routing:** Filed as [github.com/therealtinhtute/skills/issues/38](https://github.com/therealtinhtute/skills/issues/38). `cli/**` is a forbidden surface for `agent-pilot` — finding filed, not hotfixed, per this initiative's standing triage rule.

### F2 — root `AGENTS.md` shim has broken relative links (already filed, re-confirmed live)

Hit live, first navigation step, self-recovered unaided via `rg --files`. Already filed as [#37](https://github.com/therealtinhtute/skills/issues/37) from the phase's pre-flight scaffold check before this pilot ran; this run is independent live confirmation that the bug is real and that a cold agent can recover from it without help — a positive resilience signal even though it's a genuine defect.

**Severity:** Major (breaks the documented entry point on every fresh `zharness init`) but non-blocking for R9 (self-recovered).
**Routing:** Already tracked at #37, no new action.

### F3 — Codex account-level persona/memory layer (caveat, not a repo bug)

Not a `zharness`/repo defect — noted for anyone re-running this pilot or trusting future Codex-based evidence from this machine. See "Why this doesn't invalidate the pilot" above for why it doesn't compromise this specific test's validity.

**Severity:** N/A (environmental caveat).
**Routing:** None — documented here for future re-runs.

## Baseline: does `validate` ever pass clean on this repo?

Checked before concluding F1 was pilot-specific: **no.** `zharness validate --json` on this repo's own live, Claude-produced chain (run 2026-07-19, current HEAD) also returns `valid: false`, with the same failure family:

```
$ zharness validate --json
{"valid":false,"findings":[
  {"link":"SPEC->PLAN","issue":"not_yet_implemented", ...},
  {"link":"PLAN->RUN","issue":"missing_key","detail":"...plan_id... value \"null\" is not a valid ULID"} (x4, tracked as #36),
  {"link":"RUN->CHECK","issue":"missing_key","detail":"...run_id... value \"none\"..."},
  {"link":"RUN->CHECK","issue":"stale_pointer","detail":"...has no matching row in the checks table"},
  {"link":"CHECK->HANDOFF","issue":"missing_key","detail":"...run_id... value \"null\"..."},
  {"link":"CHECK->HANDOFF","issue":"missing_key","detail":"...check_id... value \"null\"..."}
]}
```

The `RUN->CHECK stale_pointer` line is the check-side twin of F1's run-side bug — it's already on record as a Major finding in `.kit/reports/check/20260719-0632-thin-triggers.md`: *"ad-hoc/simple-mode checks have no CLI path into the checks table."* Together, F1 and that finding mean: **simple/ad-hoc-mode work is structurally incompatible with the harness's story-based data model on both the run and check axes, for every agent, not just the pilot.** This is not evidence against agent-agnosticism — it's a pre-existing gap in the harness itself that both a Claude session and an independent Codex session hit identically.

## Raw evidence

- Transcript (115 JSONL events, full `codex exec --json` capture): kept out of the repo (ephemeral scratch); command trail and verbatim outputs above are the durable record.
- Scratch pilot repo (code, run/check artifacts, changesets): `/private/tmp/.../scratchpad/agent-pilot-run` — ephemeral, not committed; `textutils.py`/`test_textutils.py` contents and passing test run are captured above/in this doc's command trail.
- `zharness validate --json` / `resume --json` / `audit --json` on the scratch chain at pilot end: `valid: false` (5 findings — SPEC->PLAN not_yet_implemented; PLAN->RUN missing_key `plan_id "none"`; PLAN->RUN broken_link `phase "none"`; PLAN->RUN stale_pointer on the run id; RUN->CHECK stale_pointer on the check id); `resume` shows `readiness: clean` with all pointers null (nothing registered); `audit` mirrors the same 5 `contract_violations`, `entropy_score: 25`.

## Recommendation

Per agent-pilot-CONTEXT.md's own locked Escalate-If rule: *"The second agent cannot complete the lifecycle due to a docs gap that a reasonable fix cannot wait on → to-plan phase (playbook-fix mini-phase) — this is the one case where a fix cycle precedes closing the initiative."*

F1 is exactly this case: a genuine, reproducible harness/schema gap (not a docs-clarity gap, not agent incompetence) blocking R9's literal bar for every agent. Recommend: insert a `to-plan` playbook-fix mini-phase to resolve #38 (and ideally its check-side twin), then re-run this pilot before closing the initiative. This is a scope decision — routing is pre-authorized by CONTEXT.md, but the fix itself touches `cli/**` (forbidden surface for `agent-pilot`) and implies a new CLI release past the shipped v0.2.0, so starting that work needs explicit authorization rather than silent continuation.

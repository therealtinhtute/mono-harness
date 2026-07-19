# Playbook: check

## Purpose

Pre-commit and pre-merge quality gate. Run tests, types, lint, build with real evidence, then review security, performance, architecture, and code quality. Acts as the phase gate after `work`. "It looks fine" is not a result — gate proves it works, review proves it matches the plan, stays in scope, and is well-written.

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build always satisfies the gate. Otherwise, if the binary is missing or below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.

## Modes

| Argument | Does |
|----------|------|
| `gate` | Automated checks only: tests, types, lint, build |
| `review` | Gate → code analysis |
| `full` (default) | Gate → artifact alignment → code analysis |

## When to Use

- Before committing, creating a PR, or merging
- After implementing a feature or fix
- As the per-phase quality gate after `work`
- When reviewing a specific issue or change

`check` performs no file edits — it only reads (source, diffs, config) and proposes fixes for a human or a follow-up `work` invocation to apply.

## Steps

### Step 0 — Project Context

Before reviewing: read the diff, skim only the needed repo docs/config, compress findings into verify command + protected/generated files + domain risks, detect whether harness artifacts exist, then apply the stricter rule.

**What to read** (only files relevant to the changed code):

| File | Extract |
|------|---------|
| `README.md` | Framework, dev commands, test commands |
| `AGENTS.md` / `CLAUDE.md` | Project-specific rules that override this playbook |
| `package.json` / `Cargo.toml` / `pyproject.toml` / `go.mod` | Scripts, dependencies |
| CI workflow files | Build, test, deploy commands |
| `CHANGELOG.md` | Release conventions and version format |

Compress into:
```
verify_cmd:       [e.g. go test ./... && go vet ./...]
protected_files:  [e.g. dist/, generated/, CHANGELOG.md]
domain_risk:      [e.g. auth middleware, payment flow]
harness_mode:     full / partial / none
release_format:   [e.g. semver tag + CHANGELOG section]
```

**Conflict rule**: when project context and this playbook overlap, apply the stricter rule. If `AGENTS.md` or `CLAUDE.md` defines a verification command, use that instead of auto-detection. If project docs say never auto-commit, skip any autofix that would commit. If `.kit/planning/` and `.kit/runs/work/` are present, treat artifact alignment as part of the gate, not an optional note.

Skip context extraction when the diff is under 30 lines and does not touch config, auth, or CI, or when running `gate` mode only.

**Harness detection**: classify the repo before review — `full` (`.kit/planning/SPEC.md` plus roadmap/phase artifacts exist and `work` run logs are used), `partial` (some planning artifacts exist but run logs or phase artifacts are incomplete), `none` (no harness artifacts present). When harness artifacts exist, read `.kit/planning/` and the latest `.kit/runs/work/*.md` as the fast index, verify the pointed phase/run files, then persist a gate report at `.kit/reports/check/{YYYYMMDD-HHmm}-{slug}.md`.

### Step 1 — Scope Classification (all modes)

Measure the diff: `git diff --stat HEAD` or `git diff main...HEAD --stat`.

| Depth | Criteria |
|-------|----------|
| Quick | <100 lines, 1–5 files |
| Standard | 100–500 lines, 6–10 files |
| Deep | 500+ lines, 10+ files, or touches auth / payments / data |

State depth before proceeding.

### Step 2 — Scope Drift (all modes)

Label: **on target** / **drift** / **incomplete**. Drift = any changed file with no connection to the stated goal. Flag drift before running checks — do not silently continue.

### Step 3 — Artifact Alignment

When `.kit/planning/` artifacts are present, inspect `.kit/planning/SPEC.md`, `.kit/planning/ROADMAP.md`, the active phase `-CONTEXT.md` / `-PLAN.md`, and the latest matching `.kit/runs/work/*.md` if `work` was used. If the repo is not using the full harness flow, say so explicitly and skip artifact alignment instead of pretending it passed. Label alignment as **aligned** / **drift** / **skipped**.

**Alignment questions**:
1. **Spec Alignment** — does the diff implement behavior that maps to the spec requirements? Did it quietly add behavior outside spec scope? Are there requirement-shaped gaps the diff does not cover?
2. **Phase Boundary Alignment** — do changed files stay inside `Allowed Surfaces` and task `touches`? Did the work cross into `Forbidden Surfaces` or task `avoid` paths? Did the diff spread across subsystems the phase plan didn't expect?
3. **Execution Proof Alignment** — did each materially changed behavior have a matching verification command in the phase plan, and does the work run log show it actually ran? Did the diff add behavior with no proof trail? When `zharness` applies: does gathered proof satisfy every `required` cell in the Validation Matrix below for the resolved lane? A required cell with no matching evidence is a proof gap here too — name it the same way (missing evidence class).
4. **Decision / Context Alignment** — did implementation contradict locked decisions in the phase context? Were rejected options reintroduced? Were new assumptions added without being surfaced?

**Verdict mapping**:

| Finding | Severity | Merge impact |
|--------|----------|--------------|
| Code contradicts spec requirement | 🔴 Critical | Block |
| Changed files exceed phase boundary | 🟠 Major | Request changes |
| Missing or weak verification evidence | 🟠 Major | Request changes |
| Small context drift, documented and harmless | 🟡 Minor | Approve with note |
| Artifact missing because harness not used | 💡 Suggestion | Note only |

### Step 4 — Harness Gate Flow (when the version gate passes and `.kit/planning/` artifacts exist)

CLI-first and deterministic — the matrix replaces judgment on whether gathered proof is sufficient. Skip this step entirely for non-harness repos or when the version gate already stopped the playbook.

The **RUN artifact** is the latest matching `.kit/runs/work/{YYYYMMDD-HHmm}-{slug}.md` for the phase under review (already located in Step 0/Step 3 above). Its frontmatter carries `id` (that run's own ULID — pass this as `--run-id` below) and `trace_ids` (a list of trace ULIDs recorded once per completed wave). Read both fields directly from that file's frontmatter; no CLI query resolves them.

1. Read the lane (`tiny`/`normal`/`high-risk`) from `.kit/planning/SPEC.md`'s frontmatter `lane:` field — there is no live CLI query for it, SPEC.md is the source of truth.
2. Run `zharness audit --json`. Any non-empty `pointer_drift` or `contract_violations` touching the artifacts under review is a finding — rate it with the Severity table below (🟠 Major at minimum), it is not a separate pass/fail axis. `unlinked_proofs` and `entropy_score` are informational context for the sign-off.
3. For each id in the RUN artifact's `trace_ids` frontmatter, run `zharness score-trace {id} --json` inline. A trace scored `minimal` is too thin to count as evidence for any matrix cell below — only `standard`/`detailed` tier traces satisfy a proof-class requirement that cites a trace.
4. Evaluate the Validation Matrix below for the resolved lane against proof actually gathered this session: verification command output → `command-output`; a real test run → `unit`/`integration`/`e2e`; the Phase 2 review pass itself → `manual-check`. A `required` cell with no matching evidence ⇒ **gate FAIL**, name the exact missing evidence class, and stop — identical discipline to a failing test in Phase 1 (do not proceed to Phase 2, no judgment override).
5. Once Phase 1 (including this step) and Phase 2 both complete, translate this playbook's verdict label to the CLI's enum (`APPROVED`, `APPROVE with requests` → `APPROVE_WITH_REQUESTS`, `REQUEST CHANGES` → `REQUEST_CHANGES`).
   - **If the gated RUN's `mode` is `full`** (or the RUN artifact predates the `mode` field): run
     `zharness check record --verdict {verdict} --run-id {run id from the RUN artifact's frontmatter} --proof-links '[{"command":"...","output_ref":"...","artifact_path":"..."}, ...]' --json`
     List one `proof_links` entry per verification command actually run this session — the same commands cited in the sign-off's `verification:` line. No live command sets `meta.latest_check_id` going forward (only legacy `import` does) — run `zharness id --json`, use that fresh ID as the filename for a one-line meta changeset (`.kit/changesets/{changeset-id}.changeset.jsonl`, `{"op":"update","entity":"meta","id":"meta","fields":{"latest_check_id":"{check id just returned}"},"at":"{RFC3339 now}"}`), and apply it with `zharness db changeset apply {path} --json`, the same generic command `work`/`to-plan` already use for their own meta pointers.
   - **If the gated RUN's `mode` is `simple`**: skip `zharness check record` entirely. The RUN was never registered in the `runs` table (`work.md` Step 2, simple-mode branch), so `check record`'s `--run-id` would always fail with `unknown_run_id` — there is no row to link `checks.run_id` against. Write the persisted report with `mode: simple` and note the skip in its `## Next Action` section. `validate` treats `mode: simple` CHECK artifacts as exempt from the DB-registration check by design (see `CONTRACT.md`).
6. A missing required proof or a FAIL verdict is never overridden by this playbook. If a human judges the gap acceptable to ship anyway, they record that decision themselves: `zharness intervention --verdict-id {check id} --reason "..."`.

**Validation Matrix** (harness-aware gate) — when a `zharness` binary passes the version gate and `.kit/planning/` artifacts exist, the automated gate evaluates this lane × proof-class matrix instead of (not in addition to) the generic pass/fail table. Lane comes from `.kit/planning/SPEC.md`'s frontmatter `lane:` field (set by `intake --lane` at brainstorm time). Every cell is `required` (must have matching evidence or the gate is FAIL), `optional` (nice to have, absence never fails the gate), or `n-a` (not expected for this lane, never requested):

| Lane \ Proof class | unit | integration | e2e | manual-check | command-output |
|---|---|---|---|---|---|
| tiny | optional | n-a | n-a | optional | required |
| normal | required | optional | n-a | optional | required |
| high-risk | required | required | optional | required | required |

Proof-class meaning:
- `unit` — a unit test run covering the changed behavior (`go test ./... -run ...`, `npm test`, etc.)
- `integration` — a test that crosses a real boundary (DB, filesystem, another package's public API) rather than mocking it
- `e2e` — a full user-facing flow exercised end-to-end (browser automation, CLI smoke test against a real binary, etc.)
- `manual-check` — the Phase 2 code-review pass itself (Security/Performance/Architecture/Code Quality below) counts as this class's evidence; a clean review with no 🔴/🟠 findings satisfies it, a 🔴 finding never satisfies it regardless of lane
- `command-output` — any verification command's actual captured output (build, lint, a one-off script) — the floor every lane always requires, since it's the cheapest proof and the core "no unverified claims" hard stop already demands it

A `required` cell with no matching proof link ⇒ gate FAIL naming that exact missing evidence class. This is a hard rule, not a suggestion — the playbook does not use judgment to wave through a missing required proof; a human overrides via `zharness intervention` if the missing proof is genuinely acceptable to skip.

**Conditional reorder rules** (apply before running checks): auth/data change → secrets scan first, then tests. Time-critical → tests only, skip lint/polish. API change → backward-compat check before lint.

### Phase 1 — Gate (`gate`, `review`, `full`)

Run in order: tests, types, lint, build. Cite actual output — never self-certify. When harness artifacts apply, Step 4's matrix evaluation is part of this phase — a matrix FAIL stops the gate exactly like a failing test. If gate fails: stop, report which check failed with actual output, and do not proceed to review.

**Per-stack commands**:

Node.js / TypeScript:
```bash
npm test                      # or: yarn test / pnpm test
npx tsc --noEmit              # type check
npx eslint . --max-warnings 0
npm run build
```

Python:
```bash
pytest                        # or: python -m pytest
mypy .                        # type check
ruff check .                  # or: flake8
```

Go:
```bash
go test ./...
go vet ./...
staticcheck ./...
go build ./...
```

Rust:
```bash
cargo test
cargo clippy -- -D warnings
cargo build
```

Secrets scan (run first for auth/data changes):
```bash
git diff HEAD | grep -iE "(password|secret|token|api_key|private_key)" | grep "^\+"
```

Backward-compat check (API changes):
```bash
grep -r "functionName|ClassName|endpoint" . --include="*.ts" -l
```

Fallback when no known stack: check `package.json` → `scripts.test`/`scripts.build`; check `Makefile` → `make test`/`make build`; check `README.md` for a "Running tests" section; if nothing found, document as `verification: none — no command detected`. Never claim pass without running a real command.

**Harness add-on**: read the active phase plan and collect expected verification commands, compare against the latest matching `.kit/runs/work/*.md`. If code changed but the proof trail is missing, label `artifact_alignment: ❌ drift` even if local tests pass. If the diff exceeds allowed surfaces, stop before normal review and route back to `to-plan phase {slug}` or `work`.

**Gate outcome**:

| Result | Decision |
|--------|----------|
| All pass + on target + artifact aligned (when harness applies) | ✅ APPROVED — ready to commit |
| Minor issues remain | ⚠️ FIX — return to implementation |
| Major gaps | ❌ NEEDS_WORK — re-scope with `to-plan` or re-run `work` |

### Phase 2 — Review (`review`, `full`)

Scale depth to scope. In `full` mode, artifact drift findings come before normal code-quality commentary. Priority order: Security, Performance, Architecture, Code Quality.

**1. Security (always first)**
- Input & validation: SQL/command/path injection at every entry point; XSS vectors in rendered output; missing or bypassable input validation; file upload without type/size constraints
- Auth & access: missing authentication on protected routes; authorization bypass (horizontal + vertical privilege); insecure direct object references (IDOR); session token exposure in logs or responses
- Data exposure: PII or credentials in logs, errors, or API responses; hardcoded secrets, tokens, or keys in code; overly permissive CORS or CSP headers; sensitive data in URLs (query params, path segments)

**2. Performance**
- N+1 query patterns in loops; missing database indexes on filtered/sorted columns; unbounded queries without pagination or limits; memory leaks (event listeners not cleaned up, growing caches); blocking I/O in hot paths or request handlers; unnecessary recomputation (results not cached)

**3. Architecture**
- YAGNI — is this abstraction actually needed now? KISS — can this be simpler without losing correctness? DRY — is logic duplicated where a single source of truth exists? API contract correctness — does the interface match callers? Backward-compat — does this break existing consumers? Separation of concerns — is business logic mixed with I/O? Harness alignment — does the implementation still match the locked spec, phase context, and phase boundaries?

**4. Code Quality**
- Naming clarity — does the name say what it does without a comment? Error handling at system boundaries (external APIs, DB, file I/O). Type safety — are nulls and undefined cases handled? Test coverage — does new behavior have a test? Dead code — unreachable branches, unused imports, stale comments. Proof trail quality — can the claimed verification be traced to actual commands or run artifacts?

**Severity**:

| Level | Meaning | Blocks merge? |
|-------|---------|---------------|
| 🔴 Critical | Security / data integrity risk | **YES** |
| 🟠 Major | Bug, perf regression, wrong design | No — flagged |
| 🟡 Minor | Code quality, readability | No |
| 💡 Suggestion | Nice-to-have | No |

**Merge Gate**:
- Any 🔴 → **REQUEST CHANGES** — do not merge
- Any artifact-alignment drift that exceeds phase boundaries or contradicts the spec → at least **APPROVE with requests**, and escalate to **REQUEST CHANGES** when behavior is materially wrong
- Only 🟠 and below → **APPROVE with requests**
- Only 🟡 / 💡 → **APPROVE**

### Autofix Routing

| Class | Definition | Action |
|-------|------------|--------|
| `safe_auto` | Typos, missing imports, style inconsistencies | Propose in sign-off as ready-to-apply |
| `gated_auto` | Null checks, error handling additions | Propose in sign-off, batched, pending user confirmation |
| `manual` | Architecture, behavior, security tradeoffs | Present in sign-off |
| `advisory` | Informational only | Note in sign-off |

`check` never edits files. List every `safe_auto` and `gated_auto` fix in the sign-off; a human applies them directly or via a follow-up `work` invocation. Batch `gated_auto` into one confirmation block — never ask separately about each one.

### Pattern-Fix Completeness

When the diff fixes one instance of a class-of-bug (missing validation, wrong selector, off-by-one, missing lock), the same shape often lives elsewhere. Extract the pattern signature, `grep -rn` it across the repo (exclude generated dirs), and confirm sibling instances were also handled. List any unswept sibling: flag as a hard stop when it carries the same risk, advisory when lower-risk.

### Hard Stops

Flag before merging. Use judgment — list is not exhaustive.
- **No unverified claims**: do not write "I verified X", "I ran Y", "tests pass" unless the command output is in this session's transcript. If reasoning without running, say "based on reading the code" instead of "I verified". Every verification claim in the sign-off must point to a command that actually ran in this session.
- **Unknown identifiers**: any function, var, or type in the diff that does not exist in the codebase — grep before approving: `grep -r "name" .`
- **Hardcoded credentials**: secrets, tokens, or API keys in code, logs, or docs
- **Version skew**: version fields across manifests, changelogs, and tags out of sync
- **Generated artifact drift**: source changed but generated outputs not regenerated
- **Injection / validation gap**: SQL, command, or path injection at system entry points
- **Safety sinks**: destructive file operations (delete/move/overwrite user files, caches, history), shell/AppleScript/SQL/path construction from user input, cwd/symlink/path-traversal guard changes, sandbox/approval boundary changes, signing/notarization/appcast flows. Review validation and rollback for each.
- **Spec contradiction**: implemented behavior conflicts with a locked requirement
- **Phase boundary violation**: changed files exceed allowed surfaces without an approved plan refresh
- **Missing proof trail**: planned verification commands absent from the work run artifact or gate evidence

### Knowledge Sync

After reviewing the diff, check whether it introduces invariants not yet captured in project docs:
- New safety gate or path-guard rule → `AGENTS.md` or `CLAUDE.md`
- New UI constraint (layout rule, animation, overlay registration) → project rules docs
- New deploy/release step or artifact → `AGENTS.md` or `docs/`
- New cross-file sync requirement (enum ↔ HTML anchors, keys ↔ translations) → `AGENTS.md`

If found, apply the doc update as `safe_auto` (when the invariant is clear from the diff) or flag in sign-off as `doc debt`. When no new invariants exist, sign-off says `doc debt: none`.

## Artifacts

### Persisted Report — `.kit/reports/check/{YYYYMMDD-HHmm}-{slug}.md`

Write this when harness artifacts are present or a persisted report is requested:

```markdown
---
id: {ULID}
type: check
phase: {phase-slug} | none
lane: {tiny|normal|high-risk}
mode: {full|simple}
run_id: {ULID of the RUN this check gates}
proof_links: [{command, output_ref, artifact_path}, ...]
created: {YYYY-MM-DD}
updated: {YYYY-MM-DD}
---

# CHECK REPORT

Run ID: check-YYYYMMDD-HHmm-{slug}
Scope: gate | review | full
Artifact Alignment: aligned | drift | skipped
Review Verdict: APPROVED | APPROVE with requests | REQUEST CHANGES
Phase: {phase-slug} | none
Spec: .kit/planning/SPEC.md | none
Plan: .kit/planning/phases/{phase-slug}/{phase-slug}-PLAN.md | none
Cook Run: .kit/runs/work/{file}.md | none
Created At: YYYY-MM-DD HH:mm

## Gate Evidence
- tests: {command} → pass | fail | none
- types: {command} → pass | fail | none
- lint: {command} → pass | fail | none
- build: {command} → pass | fail | none

## Artifact Alignment
- status: aligned | drift | skipped
- notes:
  - spec coverage / gap
  - boundary compliance / drift
  - proof trail status

## Findings
### Critical
- none | finding

### Major
- none | finding

### Minor / Suggestions
- none | finding

## Next Action
- rerun `work`
- refresh `to-plan phase {slug}`
- ready for PR
```

Rules: create one file per check run; do not overwrite older results from the same day unless the exact timestamp path is reused intentionally. `run_id` links to the RUN this check gates; each `proof_links` entry is `{command, output_ref, artifact_path}` — `command` is the exact verification command run, `output_ref` is where its output is recorded (inline in the report or a path), `artifact_path` is the file the command verified. `mode` is inherited verbatim from the gated RUN artifact's own `mode` field — it decides whether Step 4 below calls `check record` or skips it.

## Output Format

Always end with this sign-off block:

```
scope:              on target / drift: [what]
depth:              quick / standard / deep
artifact_alignment: ✅ aligned / ❌ drift / skipped: [why]
gate:               ✅ pass / ❌ fail: [checks]
review:             APPROVED / APPROVE with requests / REQUEST CHANGES
blockers:           N critical, N major
autofix:            N safe_auto proposed, N gated_auto awaiting confirmation
verification:       [command] → pass / fail / none
harness_verdict:    zharness check record id / not recorded: [why]
```

## Command Reference

- `zharness --version` — version gate
- `zharness id --json` — mint a fresh filename ID before the manually-authored `latest_check_id` meta changeset
- `zharness audit --json` — pointer drift / contract violations
- `zharness score-trace {id} --json` — trace evidence tier, once per `trace_ids` entry
- `zharness check record --verdict {...} --run-id {...} --proof-links '[...]' --json` — record the verdict
- `zharness db changeset apply {path} --json` — applies the meta changeset that sets `latest_check_id`
- `zharness intervention --verdict-id {id} --reason "..."` — human override of a missing-proof gate FAIL

## Exit / Handoff Conditions

Complete only when: gate ran with real command output for every applicable check; artifact alignment was evaluated when harness artifacts exist; review covered Security → Performance → Architecture → Code Quality at the scope-appropriate depth; the sign-off block is printed; when harness applies and the gated RUN is `mode: full`, `zharness check record` ran and its meta changeset was applied — for `mode: simple`, the deliberate skip (Step 4) satisfies this instead. On a clean or approve-with-requests verdict, `git` or `handoff` are natural next steps — never run them automatically.

## Anti-Patterns

- Self-certifying "tests pass" without running them — the core gate anti-pattern; cite actual command output
- Approving because code "looks correct" without grepping unknown identifiers — hallucinated familiarity
- Skipping scope drift check on small diffs — small diffs drift too; every changed line must trace to the request
- Rating severity based on code volume instead of blast radius — 1 line touching auth can be 🔴 Critical

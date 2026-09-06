# Multi-Stack Agent-Readiness Audit — Go, Rust, Python

**Date:** 2026-09-06
**Revision audited:** `aba7057` (master)
**Scope:** the installed spine (`AGENTS.md` block, `docs/WORKFLOW.md`, the six
`docs/playbooks/*.md`, the identity template), the six thin spine triggers under
`skills/workflow/`, the optional `git` skill, the pre-commit guard in
`scripts/install-git-hooks.sh`, and the installer target set in
`cli/internal/installer/installer.go`.
**Question:** when a fresh agent lands in a Go, Rust, or Python repository that
adopted this harness, where does the harness misdirect it, and what does it cost?
**Method:** read-only. Every finding cites the harness line it traces from. No
Go, Rust, or Python repository was scaffolded, no `zharness install` ran, and no
native toolchain command was executed. The only commands executed for this audit
were the byte counts and token census in the cost section, and the two `git`
commands from the `watzup` playbook, run inside this checkout (see X1).
**Filename note:** the task prompt names the output `multi-stack-harness-audit.md`.
An earlier untracked draft already occupies that name with a nine-axis rubric; at
the author's request this audit is a separate file with a separate rubric and a
mostly disjoint finding set. Where a finding here overlaps that draft, the overlap
is stated in one line and not re-derived.

## Summary

Ranked by severity. `Job` is the harness job from
[harness-engineering-gap-audit.md](harness-engineering-gap-audit.md).
`Label` follows the task prompt: `measured` means the claim is a trace through
cited harness text (plus, for X1, an executed command); `hypothesis, unverified`
means the claim depends on consumer-repository behavior this audit could not
execute (Decision 1, see Limitations).

| ID | Severity | Stack | Job | Label | Finding | Earliest gap |
|----|----------|-------|-----|-------|---------|--------------|
| X1 | High | all | CONTEXT | measured | `watzup` step 1 hardcodes `main`; fails on any `master`/`trunk` consumer (executed here: `fatal: ambiguous argument 'main..HEAD'`) | context |
| P1 | High | Python | SENSORS | hypothesis, unverified | Plan proofs are re-executed as bare text under `sh -c`; a `pytest` that resolved through an activated virtualenv in the gate call is not the same `pytest` in the hook call | environment |
| G2 | High | Go | SENSORS | measured | Hook re-runs proofs without changing directory; a `go test ./...` proven inside a sub-module never sees that module from the work-tree root | proof |
| R1 | Medium | Rust | SENSORS | hypothesis, unverified | 300 s hard bound on each re-executed proof, no per-proof override in the hook; a cold `cargo test` on a mid-size crate can exceed it and reject a valid commit | proof |
| X2 | Medium | all | SENSORS | measured | The four-slot gate at `check.md:34` is the TypeScript toolchain shape; the identity template offers one slot, so two to three slots per stack have no authority location | authority |
| G1 | Medium | Go | POLICY | measured | `git` skill classification splits `_test.go` from the code it proves into a separate `test:` commit under the ≤3-file rule | capability |
| R2 | Medium | Rust | SENSORS | measured | Literal reading of the gate compiles the crate graph up to three times; `cargo fmt --check`, the most common Rust CI gate, fits no slot | authority |
| P2 | Medium | Python | TRACES | measured | `git add -A` at `workflow.md:29` contradicts the anti-pattern at `:135`; hook-run proofs create cache directories the installer never ignores, and the secret grep walks an unignored `.venv` | environment |
| X3 | Medium | all | TRACES | measured | `.kit/cache/reports/git/` is described as gitignored scratch; the installer ignores only `/.zharness/`, so the report is staged by the next `git add -A` | environment |
| R3 | Low | Rust | SENSORS | measured | `cargo test` runs unit, integration and doc tests in one invocation; the 3-line pass tail shows only the last target, so the `normal` proof class cannot be evidenced without extra compiles | proof |
| P3 | Low | Python | SENSORS | measured | The `type checks` slot has no applicability authority; an agent either runs an unconfigured checker on an untyped codebase or leaves the slot silently empty | authority |
| G3 | Low | Go | SENSORS | measured | `go vet` (and any configured linter) has no authority slot; the `type checks` slot has no Go analog and invites a second `go build` | authority |
| X4 | Low | all | CONTEXT | measured | Token census: 7 JavaScript-specific tokens in 5 files, 2 Go tokens (both about this repository's own CLI), 0 Rust, 0 Python; the spine is stack-neutral by omission, not by design | context |

Cost and ceremony is scored as one axis in the rubric (question Q6). Read cost is
identical across stacks; run cost differs and is the reason Rust scores lowest.

## Rubric

### Six questions a fresh agent asks

Each criterion is a question the agent must answer from harness text alone
before it can act without human steering. The criterion fails for a stack when
the harness either gives no answer or gives an answer shaped by a different
stack.

| Q | Question | Job(s) | What passes | What fails |
|---|----------|--------|-------------|------------|
| Q1 | Where am I? | CONTEXT | Branch, base, plan and stack facts are discoverable from harness text or a harness-specified command that works in this repo | A harness command that assumes a layout or name this repo does not have |
| Q2 | What may I change, and how is it committed? | CONTRACT, POLICY | Bounded-mode thresholds and commit grouping keep code and its proof together | Grouping or thresholds that separate a change from its test, or misname a dependency change |
| Q3 | What proves it? | SENSORS | The gate taxonomy and the identity template together name every command the stack needs, with cwd and environment | Slots with no authority location; proofs that depend on shell state the text does not carry |
| Q4 | What enforces it? | TOOLS, STATE | The hook re-executes the proof exactly as the agent ran it, within its bounds | Re-execution differs from the in-session run (cwd, env, time) |
| Q5 | What do I leave behind? | TRACES | Nothing unintended reaches the index or the tree | Caches, reports, or environments staged or scanned |
| Q6 | What does it cost? | all | Read packet and run cost proportional to the change | Repeated compilation, repeated reads, or an install channel the stack does not have |

### How this builds on the two prior frameworks

**From the gap audit.** The seven jobs (CONTRACT, CONTEXT, TOOLS, STATE,
SENSORS, POLICY, TRACES) are kept as the classification of *where* a finding
lives; every row above names one. The gap audit's Level 0–3 scale measured how
deeply a harness *encodes* a capability. This audit re-purposes the same 0–3
scale to measure *stack fit*: how far a stack can get on each question without
the human filling in what the text left out. The scale is reused rather than
replaced so the two audits can be read side by side; the meaning of a level is
different and is stated below.

**From the consumer-adoption audit.** Its cold-entry token cost method (bytes of
every file the agent reads before its first action, divided by four) is reused
unchanged, and the numbers are the same as any other audit of this revision
because the inputs are identical. Its `observed` / `inferred` labeling is mapped
onto the task prompt's `measured` / `hypothesis, unverified`. The departure is the
addition of a *run cost* row: the number of native gate invocations a literal
reading of `check.md:34` implies per stack. Read cost is stack-neutral; run cost
is not, and it is the only cost that changes between the three stacks.

**Departures.** Two positions here differ from the sibling draft. First, the
four-slot gate taxonomy is treated as a finding (X2), not as neutral wording,
because "repository-defined order" presumes the repository has defined those
commands somewhere the agent is told to look, and the identity template gives one
slot. Second, the audit adds a portability measurement (X4) instead of a
stack-token count offered as context: neutrality achieved by never naming a
toolchain is not the same as neutrality achieved by naming all of them.

### Scoring

Per question, per stack, 0–3:

- **0** — the harness answer is wrong for this stack, or absent where the stack
  needs it, and a fresh agent will misfire on a common path.
- **1** — the answer exists but is shaped by another stack; a fresh agent works
  only with human steering or by ignoring the text.
- **2** — the answer is stack-neutral and correct; the agent must still fill in
  stack detail from its own knowledge.
- **3** — the answer is verified against a live repository of this stack. By
  Decision 1 (no synthetic repositories) no cell can reach 3 in this audit; the
  ceiling is 2 and the scale is kept so a follow-up with live repositories can
  fill it.

### Labels

- `measured` — the claim is a trace through quoted harness text, checked by
  re-reading each cited line after the section was drafted; X1 additionally has
  an executed command with captured output.
- `hypothesis, unverified` — the claim's consequence depends on a consumer
  repository, toolchain, or shell that this audit did not run. The mechanics are
  cited; the outcome is not observed. Two findings carry this label (P1, R1).

### Scorecard

Cells show `spine / with git skill` where the optional `git` skill changes the
answer; otherwise one value.

| Q | Go | Rust | Python | Why |
|---|----|------|--------|-----|
| Q1 Where am I | 1 | 1 | 1 | X1 fails on `master` consumers in every stack; the identity template has no stack or layout field |
| Q2 What may I change | 2 / 1 | 2 / 1 | 2 / 1 | Bounded thresholds are neutral; `git` grouping splits code from test (G1, shared) |
| Q3 What proves it | 1 | 1 | 1 | One-slot template against a four-slot gate (X2, G3, R2, P3) |
| Q4 What enforces it | 1 | 1 | 0 | Hook cwd (G2), hook bound (R1), hook environment (P1); Python has no env-free proof spelling the template asks for |
| Q5 What do I leave behind | 2 / 1 | 2 / 1 | 1 / 0 | X3 for all; Python adds cache directories and `.venv` under `git add -A` (P2) |
| Q6 What does it cost | 2 | 1 | 2 | Same read packet; Rust run cost multiplies compiles (R2) |

## Cost and ceremony

### Read cost (stack-neutral)

Bytes measured with `wc -c` at `aba7057`; tokens are bytes ÷ 4 as in the
consumer-adoption audit. Reproduce with:

```bash
wc -c AGENTS.md docs/WORKFLOW.md skills/workflow/*/SKILL.md docs/playbooks/*.md \
  skills/workflow/git/references/workflow.md \
  cli/docs/embedded/templates/project.identity.md docs/PROJECT.md
```

| Stage packet | Files | Bytes | ≈ tokens |
|--------------|-------|-------|----------|
| Always (AGENTS block + WORKFLOW.md) | 2 | 2,913 | 728 |
| watzup | + SKILL 1,010 + playbook 2,316 | 6,239 | 1,560 |
| brainstorm | + SKILL 1,269 + playbook 5,797 | 9,979 | 2,495 |
| to-plan | + SKILL 1,104 + playbook 4,544 | 8,561 | 2,140 |
| work | + SKILL 1,189 + playbook 8,848 | 12,950 | 3,238 |
| check | + SKILL 1,157 + playbook 9,595 | 13,665 | 3,416 |
| handoff | + SKILL 1,068 + playbook 6,297 | 10,278 | 2,570 |
| git skill (optional) | SKILL 2,511 + workflow.md 6,641 | 9,152 | 2,288 |
| Identity template (shipped) | 1 | 543 | 136 |
| This repo's filled `docs/PROJECT.md` | 1 | 2,168 | 542 |

The `check` packet is the largest and is the one every stack pays at least once
per bounded change. The packet is the same for Go, Rust, and Python because none
of the files contain a stack branch; the difference between stacks is what the
agent has to *add* to the packet from its own knowledge, which the scorecard
captures under Q3.

### Run cost (stack-specific)

Native invocations a fresh agent issues under a literal reading of the gate at
[check.md:34](../playbooks/check.md#L34) — "applicable tests, type checks,
lint/static analysis, and build in repository-defined order" — with nothing in
the identity template beyond the test command
([project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13)).

| Slot | TypeScript (the implied shape) | Go | Rust | Python |
|------|-------------------------------|----|------|--------|
| tests | `npm test` | `go test ./...` | `cargo test` (compiles test profile) | `pytest` |
| type checks | `tsc --noEmit` | none (compile is the type check; agent may re-run `go build`) | `cargo check` (separate metadata build) | `mypy` / `pyright` if configured, else undefined |
| lint/static | `eslint` | `go vet ./...` (+ `golangci-lint` if present) | `cargo clippy` (separate compile with the clippy driver) | `ruff check` |
| build | `npm run build` | `go build ./...` | `cargo build` (dev profile) | none for applications; `python -m build` for packages |
| unslotted | `prettier --check` | `gofmt -l` | `cargo fmt --check` | `ruff format --check` |
| Invocations | 4 | 3–4 | 4 | 2–4 |
| Distinct compilations of the whole graph | 1 | 1 (cached) | up to 3 | 0 |

Read one row at a time, the four slots are `npm test` / `tsc` / `eslint` /
`npm run build`. That is the measured origin of X2: the taxonomy is not wrong for
any stack, but it was written by a TypeScript hand and the other three stacks each
have a slot that is empty, doubled, or missing.

### Portability by omission (token census)

Token census over the spine, the six triggers, the `git` skill, the identity
template, `CLAUDE.md`, `README.md` and the prompt-engineering principles:

```bash
rg -n -i 'package\.json|npm|npx|node_modules|eslint|biome|tsc\b|pnpm|yarn' \
  AGENTS.md CLAUDE.md README.md docs/WORKFLOW.md docs/playbooks skills/workflow \
  docs/prompt-engineering-principles.md cli/docs/embedded/templates/project.identity.md
rg -n -i '\bgo test\b|\bgo vet\b|go\.mod|\bcargo\b|Cargo\.toml|pytest|\bruff\b|mypy|pyproject|\bvenv\b' \
  <same paths>
```

| Stack | Hits | Files | Where |
|-------|------|-------|-------|
| JavaScript/TypeScript | 7 | 5 | [workflow.md:44](../../skills/workflow/git/references/workflow.md#L44), [workflow.md:135](../../skills/workflow/git/references/workflow.md#L135), [CLAUDE.md:15](../../CLAUDE.md#L15), [CLAUDE.md:61](../../CLAUDE.md#L61), [CLAUDE.md:64](../../CLAUDE.md#L64), [principles:115](../prompt-engineering-principles.md#L115), [README.md:118](../../README.md#L118) |
| Go | 2 | 2 | [README.md:141](../../README.md#L141), [CLAUDE.md:77](../../CLAUDE.md#L77) — both are this repository's own `cli/` gate, not consumer guidance |
| Rust | 0 | 0 | (the one `cargo` match, "Cargo-culting" at principles:112, is an idiom) |
| Python | 0 | 0 | — |

The spine itself (`WORKFLOW.md`, six playbooks, six triggers, identity template)
scores zero in every column. The JavaScript hits are all in the optional layer
and the repository's own docs; the Go hits are the repository dogfooding its own
CLI. See X4 for why zero is not the same as portable.

## Go

### G1 — `git` skill grouping separates `_test.go` from the code it proves

**Severity:** Medium · **Label:** measured · **Job:** POLICY · **Owner:** repository-harness

**Evidence.**
[workflow.md:44](../../skills/workflow/git/references/workflow.md#L44):
"Group staged files by kind (`docs:` for `.md`/`.txt`, `test:` for test/spec
paths, `config:` for `.claude/` files, `deps:` for `package.json`/lockfiles,
`code:` for everything else)."
[workflow.md:46](../../skills/workflow/git/references/workflow.md#L46):
"**Single commit:** same type/scope, files ≤ 3, lines ≤ 50."
[workflow.md:47](../../skills/workflow/git/references/workflow.md#L47): mixed
types produce one commit per group.

**Trace.** A bounded Go change under [work.md:17](../playbooks/work.md#L17)
touches `handler.go`, `handler_test.go`, and often `go.mod` and `go.sum`. Go
colocates tests with code by convention and the test file name is the only
marker. Either reading of "test/spec paths" misfires:

- If `handler_test.go` counts as a test path, four files exceed the ≤3 rule and
  the grouping rule yields a `feat` commit containing `handler.go` (with or
  without `go.mod`/`go.sum` depending on whether the agent reads them as
  "lockfiles") and a separate `test` commit containing `handler_test.go`. The
  first commit has no proof of its own; `git bisect`, per-commit CI and the
  reviewer reading commit by commit see untested code followed by a test.
- If it does not count (the phrase names paths, not file suffixes), Go has no
  `test:` category at all and every test change lands as `code:`; the rule is
  inert for the stack.

**Consequence.** The commit that the pre-commit hook proves is not the commit
that carries the test, because the hook proves the plan's Validation entry, not
each commit's content. Rust (`tests/` directory) and Python (`tests/`) hit the
first branch by directory name; Go is the case where the rule is ambiguous.

**Earliest gap.** Capability: the grouping rule is written for a layout where
tests live under `test/` or `spec/`, and the size floor has no exception for
"code plus its test".

### G2 — Hook re-executes proofs from the work-tree root; sub-module proofs never resolve

**Severity:** High · **Label:** measured · **Job:** SENSORS · **Owner:** repository-harness

**Evidence.**
[install-git-hooks.sh:103](../../scripts/install-git-hooks.sh#L103) runs each
proof as `timeout 300 sh -c "$1"`; the hook body computes
`ROOT="$(git rev-parse --show-toplevel)"` at
[install-git-hooks.sh:270](../../scripts/install-git-hooks.sh#L270) and never
changes directory before running proofs. Git documents that hooks run from the
root of the working tree. [CONTRACT.md:30](../../cli/docs/CONTRACT.md#L30)
restates the `sh -c` re-execution. [check.md:28](../playbooks/check.md#L28) and
[work.md:37](../playbooks/work.md#L37) ask the agent to record and run the exact
command. The identity template's single slot,
[project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13),
does not say the command must be valid from the repository root.

**Trace.** A Go module under a subdirectory is common (this repository's own
CLI lives under `cli/`). An agent working inside that directory proves its change
with `go test ./...`, which is correct there. The hook re-executes the same
string from the root. Two outcomes:

- No `go.mod` at the root: `go test ./...` exits non-zero with a "no Go files"
  or "cannot find main module" message; the hook prints the 10-line tail
  ([install-git-hooks.sh:188](../../scripts/install-git-hooks.sh#L188)) and
  rejects a commit whose proof passed in session.
- A `go.mod` at the root (multi-module workspace): `./...` stops at the nested
  module boundary, tests the root module, passes, and the hook accepts a commit
  whose proof never ran.

**Why Go is the sharp case.** `./...` is the idiomatic scope and never crosses a
module boundary, so the cwd is part of the command's meaning. This repository
avoids the trap by writing `cd cli && …` into its own gate at
[CLAUDE.md:77](../../CLAUDE.md#L77) and
[PROJECT.md:26](../PROJECT.md#L26); the template a consumer receives does not
carry that instruction.

**Earliest gap.** Proof: the contract says "exactly as run" and the runner
silently changes one input of the run.

### G3 — `go vet` and configured linters have no authority slot; the type-check slot has no Go analog

**Severity:** Low · **Label:** measured · **Job:** SENSORS · **Owner:** repository-harness

**Evidence.** [check.md:34](../playbooks/check.md#L34) names four slots;
[project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13)
provides one. This repository's own gate at
[README.md:141](../../README.md#L141) is `go build ./... && go vet ./... && go test ./...`
and `go vet` is not derivable from the template.

**Trace.** A fresh agent filling the four slots for Go has no text to consult
for `type checks` (Go's compile is the type check) and none for `lint/static`.
The two literal readings are to run `go build` twice (once as types, once as
build) or to mark the slot not applicable. `go vet` runs only if the agent
brings it from memory; a configured `golangci-lint` runs only if the agent finds
`.golangci.yml` on its own, which no playbook step asks for.

**Earliest gap.** Authority: the identity template is the only harness-owned
place a repository states its gate, and it has one question.

## Rust

### R1 — 300-second proof bound with no per-proof override; cold `cargo test` can reject a valid commit

**Severity:** Medium · **Label:** hypothesis, unverified · **Job:** SENSORS · **Owner:** repository-harness

**Verified part.** [install-git-hooks.sh:103](../../scripts/install-git-hooks.sh#L103):
`timeout 300 sh -c "$1"` (with `gtimeout` fallback, else unbounded);
[CONTRACT.md:30](../../cli/docs/CONTRACT.md#L30) states the five-minute bound.
`scripts/record-check.sh` accepts a `-t SECONDS` override for the in-session
capture; the hook accepts none, so a proof that needs more time cannot be
declared as such.

**Unverified part.** That a consumer crate's `cargo test` exceeds 300 s from a
cold `target/`. Cold builds happen on fresh clones, after a toolchain bump, after
`cargo clean`, and in the CI re-run that [check.md:40](../playbooks/check.md#L40)
promises. The audit did not compile a crate to measure this; the claim is that
the bound is fixed and unannounced, not that a specific crate crosses it.

**Trace.** The in-session proof passes on a warm cache. The hook re-run on a cold
cache is killed at 300 s; the failure tail is the last ten lines of a compile in
progress, which is not an error the agent can fix. The recovery a fresh agent
reaches for is to cite a cheaper proof (`cargo check`, or `cargo test -p` on one
crate), which weakens the guarantee the hook exists to give.

**Earliest gap.** Proof: the runner's bound is not part of the Validation entry
contract at [check.md:28](../playbooks/check.md#L28), so the agent cannot know
it is writing a proof that will be cut.

### R2 — Literal four-slot gate compiles the crate graph up to three times; `cargo fmt --check` fits no slot

**Severity:** Medium · **Label:** measured · **Job:** SENSORS · **Owner:** repository-harness

**Evidence.** [check.md:34](../playbooks/check.md#L34) four slots;
[project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13)
one slot.

**Trace.** Filling the slots for Rust from the text alone gives `cargo test`
(tests), `cargo check` (type checks), `cargo clippy` (lint), `cargo build`
(build). `cargo test` and `cargo build` produce different artifacts; `cargo
clippy` compiles with its own driver and does not reuse `cargo check` metadata
across all configurations. A fresh agent that follows the slot list in
"repository-defined order" with no repository definition runs all four and
compiles the graph up to three times. The gate the Rust ecosystem actually runs
first, `cargo fmt --check`, is a formatting check that fits none of the four
names and is dropped, so a local gate passes and CI fails on formatting.

**Consequence.** The `check` packet, already the most expensive to read, is
the most expensive to run for Rust, and the wasted compiles are exactly the ones
that push a re-executed proof toward the R1 bound.

**Earliest gap.** Authority: no harness-owned place says "for this repository
the gate is `cargo fmt --check && cargo clippy -- -D warnings && cargo test`",
which is one command and one compile.

### R3 — One `cargo test` covers three proof classes; the 3-line pass tail evidences only the last

**Severity:** Low · **Label:** measured · **Job:** SENSORS · **Owner:** repository-harness

**Evidence.** [check.md:37](../playbooks/check.md#L37) distinguishes proof
classes (tiny: command output; normal: unit plus command; high-risk: unit,
integration, manual review, command). [check.md:34](../playbooks/check.md#L34)
records the pass tail through `scripts/record-check.sh`, which keeps three lines
on pass ([record-check.sh:54](../../scripts/record-check.sh#L54)) and ten on
fail.

**Trace.** `cargo test` runs `#[cfg(test)]` unit tests, `tests/` integration
tests and doc tests as sequential targets in one invocation. The last three lines
of output are the summary of the last target, usually doc tests. A Validation
entry for a `normal` or `high-risk` change therefore cannot show the unit or
integration class it claims from the tail; to show them the agent splits into
`cargo test --lib`, `cargo test --tests`, `cargo test --doc`, each a separate
re-executed proof under the R1 bound. The `git` skill's grouping (G1) then
commits `tests/` separately from `src/`.

**Earliest gap.** Proof: the proof classes were written for toolchains where
unit and integration runs are separate commands.

## Python

### P1 — Proofs are re-executed as bare text under `sh -c`; the interpreter that ran in session is not carried

**Severity:** High · **Label:** hypothesis, unverified · **Job:** SENSORS · **Owner:** repository-harness

**Verified part.** [install-git-hooks.sh:103](../../scripts/install-git-hooks.sh#L103)
re-runs the Validation entry's text under `sh -c`;
[CONTRACT.md:30](../../cli/docs/CONTRACT.md#L30) states this as the guarantee.
[check.md:28](../playbooks/check.md#L28) and
[work.md:37](../playbooks/work.md#L37) ask for the command as run.
[project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13)
has no interpreter, environment or wrapper field. Every proof this repository
runs on itself (`bash scripts/…`, `cd cli && go …`) is environment-free, so the
harness authors never exercised the case.

**Unverified part.** What a consumer's hook environment resolves `pytest` to.
The audit did not create a virtualenv or run a hook.

**Trace.** The Python command as run is `pytest` (or `mypy`, `ruff check`), and
it resolved because the agent's shell had a virtualenv activated. Activation is
shell state, not command text. Agent runtimes that execute each tool call in a
fresh shell (this one does) drop that state between the gate call and the commit
call; CI re-runs drop it by construction. The hook then resolves `pytest` to:
nothing (exit 127, commit rejected with a one-line tail), a system-wide `pytest`
(runs against a different dependency set, may pass or fail for unrelated
reasons), or the right one only when the developer's global shell happens to
match. Environment-carrying spellings exist (`uv run pytest`, `.venv/bin/pytest`,
`poetry run pytest`, `tox -e py`) and the harness never asks for one.

**Earliest gap.** Environment: the harness treats "the command as run" as
sufficient identity for a proof; for Python the interpreter is part of the
identity.

### P2 — `git add -A` contradicts the skill's own anti-pattern; hook-run proofs create unignored caches and the secret grep walks `.venv`

**Severity:** Medium · **Label:** measured · **Job:** TRACES · **Owner:** repository-harness

**Evidence.**
[workflow.md:29](../../skills/workflow/git/references/workflow.md#L29):
`git add -A && git diff --cached --stat && git diff --cached --name-only`.
[workflow.md:135](../../skills/workflow/git/references/workflow.md#L135):
"Staging everything with `git add -A` instead of specific files — catches
`.env`, secrets, `node_modules`."
[workflow.md:37](../../skills/workflow/git/references/workflow.md#L37): secret
grep for `(AKIA|api[_-]?key|token|password|secret|credential|private[_-]?key|…)`
over staged content, with a STOP at
[workflow.md:40](../../skills/workflow/git/references/workflow.md#L40).
[installer.go:437–440](../../cli/internal/installer/installer.go#L437): the
installer's ignore contribution is the marker line and `/.zharness/` only.

**Trace.** The procedure and its anti-pattern list disagree, and the procedure
wins because it is the step the agent executes. In Python the disagreement has
teeth: every hook re-execution of `pytest`, `mypy` and `ruff` writes
`.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` and `__pycache__/` into the
tree at gate time; the next `git add -A` stages whichever of those the consumer
has not ignored. A project-local `.venv/` that is not ignored is staged in full,
and the secret grep at `:37` matches `token`, `secret`, `password` and
`credential` thousands of times across `site-packages`, so the STOP fires on
noise and the agent either aborts or learns to skip the STOP. The anti-pattern
line names `node_modules`, the JavaScript equivalent, and nothing else.

**Consequence.** The `git` skill's safety step is calibrated for a curated
ignore file the harness neither ships nor checks for.

**Earliest gap.** Environment: the harness assumes the consumer's ignore file
already covers the stack's artifacts and gives no step to verify it.

### P3 — The `type checks` slot has no applicability authority for optionally typed code

**Severity:** Low · **Label:** measured · **Job:** SENSORS · **Owner:** repository-harness

**Evidence.** [check.md:34](../playbooks/check.md#L34) lists "type checks" as
one of four "applicable" slots;
[project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13)
has no field to say whether the repository type-checks, with what, or against
which strictness.

**Trace.** Python typing is optional and partial by default. A fresh agent
deciding applicability from harness text has nothing to decide with. If it runs
`mypy` on a codebase that has never been type-checked, hundreds of pre-existing
errors return and [check.md:38](../playbooks/check.md#L38)'s judge sees a red
gate on a clean diff. If it marks the slot not applicable, the Validation entry
carries no type evidence on a repository that does run `pyright --strict` in CI.
Either way the Validation entry is honest to the letter of `check.md:28` and
wrong about the repository. Go has the mirror problem at G3 with the slot having
no analog; Python has it with the slot being unconfigurable.

**Earliest gap.** Authority: the template never asks "which of the four slots
apply here, and with what command?"

## Cross-stack

### X1 — `watzup` step 1 hardcodes `main`; fails on `master` consumers

**Severity:** High · **Label:** measured · **Job:** CONTEXT · **Owner:** repository-harness

**Evidence.** [watzup.md:14](../playbooks/watzup.md#L14) runs
`git log --oneline main..HEAD` and `git rev-list --left-right --count main...HEAD`.
Executed in this checkout, whose default branch is `master`:

```
fatal: ambiguous argument 'main..HEAD': unknown revision or path not in the working tree.
```

**Trace.** This is the first command of the session-start playbook and the only
one that reads the repository's shape. It fails on every consumer whose default
branch is `master`, `trunk`, `develop`, or anything else, including the harness's
own source repository. Older Go and Rust projects and most pre-2020 Python
projects use `master`. The agent recovers by guessing the base branch, which is
exactly the "where am I" answer the step exists to give.

**Why cross-stack rather than a single stack.** The failure is
language-neutral; it is ranked High because it is the one finding this audit
observed by execution, it hits the entry stage, and the fix is one line
(`git symbolic-ref --short refs/remotes/origin/HEAD` with a fallback).

**Earliest gap.** Context.

### X2 — The four-slot gate is the TypeScript toolchain shape against a one-slot template

**Severity:** Medium · **Label:** measured · **Job:** SENSORS · **Owner:** repository-harness

**Evidence.** [check.md:34](../playbooks/check.md#L34);
[project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13);
the run-cost table above.

**Trace.** The slots `tests / type checks / lint/static analysis / build` are a
one-to-one map of `npm test / tsc / eslint / npm run build`. Read as "whatever
your repository does", the wording is neutral, and the sibling draft reads it
that way. Read by a fresh agent with a one-slot identity file, "applicable" is a
decision with no input, and "repository-defined order" points at a definition
the agent is never told how to find (no playbook step says "read the CI workflow
or Makefile for the gate"). G3, R2 and P3 are the three stack-specific
manifestations; this finding is the shared cause.

**Earliest gap.** Authority.

### X3 — `.kit/cache/reports/git/` is documented as gitignored scratch; the installer never ignores it

**Severity:** Medium · **Label:** measured · **Job:** TRACES · **Owner:** repository-harness

**Evidence.** [workflow.md:120](../../skills/workflow/git/references/workflow.md#L120)
writes pr/merge reports to `.kit/cache/reports/git/{YYYYMMDD-HHmm}-{operation}.md`
described as gitignored local scratch. [installer.go:437–440](../../cli/internal/installer/installer.go#L437)
adds only the marker and `/.zharness/`. This repository's own ignore file
carries a `**/.kit/` entry inherited from the 0.14 line; a fresh consumer's does
not.

**Trace.** First `pr` or `merge` operation in a consumer writes the report; the
next `git add -A` at [workflow.md:29](../../skills/workflow/git/references/workflow.md#L29)
stages it into an unrelated commit. Language-neutral, and cheap to fix in either
the installer's ignore list or the skill's path.

**Earliest gap.** Environment.

### X4 — Stack-neutral by omission: 7 JS tokens, 2 Go, 0 Rust, 0 Python

**Severity:** Low · **Label:** measured · **Job:** CONTEXT · **Owner:** repository-harness

**Evidence.** The token census table in the cost section, with the commands used.

**Trace.** The spine contains no toolchain name at all, which is why every
playbook line reads correctly for every stack. The three places that do name a
toolchain are the optional `git` skill (`package.json`, `node_modules`), the
prompt-engineering principles (`Biome`, `eslint` at
[principles:115](../prompt-engineering-principles.md#L115)), and the skill
install channel (`npx skills add` at [README.md:118](../../README.md#L118) and
[CLAUDE.md:61](../../CLAUDE.md#L61), which requires Node on a Go, Rust or
Python developer's machine to obtain the `git` skill at all). Neutrality here
is the absence of any stack, so every stack-specific decision (G2 cwd, P1
interpreter, R2 compile count) is pushed to the agent's memory. A harness that
named all four stacks in one identity field would be neutral by design and would
close G3, R2, P3 and X2 in the same edit.

**Earliest gap.** Context.

## Proposals

Stated in the shape of `docs/templates/harness-improvement.md`: *If <smallest
change> is added at <owner>, then a fresh agent will <observable change> on
<job>, because <mechanism>.* Ordered by findings closed per line changed. None
of these were applied; this audit is read-only.

1. **If** `watzup.md:14` resolves the base branch with
   `git symbolic-ref --short refs/remotes/origin/HEAD 2>/dev/null || echo main`
   before the two commands, **then** a fresh agent will complete step 1 on
   `master`/`trunk` consumers on CONTEXT, **because** the only hardcoded
   repository fact in the spine becomes discovered. Closes X1.
2. **If** the identity template gains one line per gate slot
   (`## How do we run the gate?` with `tests:`, `types:`, `lint:`, `build:`,
   `format:` entries, each "n/a" allowed, and a `run from:` cwd note), **then** a
   fresh agent will fill `check.md:34` from authority instead of memory on
   SENSORS, **because** "applicable" and "repository-defined order" acquire a
   definition. Closes X2, G3, R2, P3; halves the Rust run cost.
3. **If** `check.md:28` states that proof commands are re-executed from the
   repository root under `sh -c` with a 300 s bound and no shell state, and the
   example proof shows `cd <dir> && <wrapper> <cmd>`, **then** a fresh agent will
   write `cd cli && go test ./...` and `uv run pytest` rather than `go test
   ./...` and `pytest` on SENSORS, **because** the three inputs the runner
   changes (cwd, env, time) become part of the contract the agent writes to.
   Closes G2; downgrades P1 and R1 to documented behavior.
4. **If** the hook honours an optional `# timeout: <seconds>` annotation on a
   Validation entry (the same override `record-check.sh -t` already accepts),
   **then** a fresh agent will declare a slow Rust proof instead of substituting a
   weaker one on TOOLS, **because** the bound becomes declarable. Closes R1.
5. **If** `workflow.md:29` stages by explicit path list from the plan's file
   list and the grouping rule at `:44–47` exempts "a source file and the test
   file that proves it" from the ≤3-file split, **then** a fresh agent will
   commit `handler.go` with `handler_test.go` and never stage a cache directory
   on POLICY and TRACES, **because** the procedure stops contradicting its own
   anti-pattern. Closes G1, P2 (staging half), X3.
6. **If** the installer's ignore contribution at `installer.go:437–440` adds
   `.kit/` and the `git` skill's report path is moved under `.zharness/`,
   **then** a fresh agent will leave no scratch in the index on TRACES,
   **because** the path the skill calls gitignored becomes gitignored. Closes X3
   (alternative to 5).

## Ownership

| Gap owner | Findings | Who can close it |
|-----------|----------|------------------|
| context | X1, X4 | repository-harness (playbook and template text) |
| authority | X2, G3, R2, P3 | repository-harness (identity template); consumer repo fills it |
| proof | G2, R1, R3 | repository-harness (hook contract text; hook annotation) |
| environment | P1, P2, X3 | repository-harness (template field, installer ignore); consumer repo (wrapper spelling, ignore file) |
| capability | G1 | repository-harness (`git` skill grouping rule) |

No finding is owned by the human or by the agent's environment alone; every one
has a harness-side line that can absorb it.

## Limitations and validation

**Decision 1 — no live repositories.** No Go, Rust or Python repository was
scaffolded and no native command was run. The reasoning is from harness text
against the known behavior of `go test` / `go vet`, `cargo test` / `cargo
clippy`, `pytest` / `ruff` / `mypy`, and the manifest files `go.mod`,
`Cargo.toml`, `pyproject.toml` versus `package.json`. Consequences:

- Two findings (P1, R1) are labeled `hypothesis, unverified` because their
  consequence depends on a consumer environment. Their mechanics are cited.
- The scorecard ceiling is 2 of 3. A follow-up that scaffolds one minimal
  repository per stack, installs the harness, and runs `zharness install`, the
  hook, and one bounded change through `check` would convert every `measured`
  trace into an observed pass or fail and can fill the 3 column.
- The run-cost table counts invocations under a literal reading; it does not
  measure seconds. R1's 300 s claim is therefore about the bound, not about a
  crate.

**Validation performed for this document.**

- Each cited line was re-read after its section was drafted and the quoted
  text matched the file at `aba7057`.
- The token census and byte counts were produced by the commands shown, at the
  same revision.
- `bash scripts/verify-doc-links.sh` passes with this file present (every
  repo-relative link above resolves; `scripts/…` paths sit outside the checker's
  prefix set, so the checker does not evaluate that prefix).
- `git status` shows this file as the only addition beyond the two untracked
  files that predate it (the task prompt and the sibling draft). No file under
  `skills/`, `docs/playbooks/`, `docs/templates/`, `cli/`, `docs/plans/` or any
  existing file under `docs/audit/` was modified.

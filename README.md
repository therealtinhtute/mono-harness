# mono-harness

Turn a git repository into a legible, agent-ready workspace.

`zharness` installs a small repository protocol and a safe updater. The
repository remains the system of record: product documents, decisions, plans,
code, tests, CI, and runtime evidence define the work.

It is not a task database, story tracker, agent orchestrator, or application
runtime. The binary scaffolds docs; it does not run the lifecycle.

Start with [`AGENTS.md`](AGENTS.md), then [`docs/WORKFLOW.md`](docs/WORKFLOW.md).

## Give this to a coding agent

Copy the block into the consumer repository chat when you want the agent to
install, refresh, or work with zharness.

```text
Work in this git repository. Read https://github.com/therealtinhtute/mono-harness/blob/master/README.md and, once they exist, this repo's AGENTS.md and docs/WORKFLOW.md. If zharness is not on PATH or outdated, install/upgrade it with scripts/install-zharness.sh. If this repo has no docs/WORKFLOW.md, run zharness install. If the repo is already initialized, playbooks are outdated, or zharness install reports drifted files, run zharness update to refresh the latest playbooks (fresh-overwrite; PROJECT.md and the AGENTS.md block merge instead). Never put plans in .kit/ — multi-session or complex work uses exactly one file at docs/plans/active/{slug}.md, moved to docs/plans/completed/ upon verified handoff. Small changes need no plan. zharness only manages the doc set (install/update/uninstall) — it does not run the lifecycle. Missing AGENTS.md: zharness install, do not invent the file. Claude Code reads CLAUDE.md, not AGENTS.md; if CLAUDE.md is missing, the consumer writes a thin file containing the line @AGENTS.md.
```

## What it solves

Coding agents often fail for ordinary engineering reasons:

- important intent exists only in chat;
- the repository does not identify authoritative documents;
- small changes acquire unnecessary process;
- long changes lose decisions and recovery context;
- completion is claimed without behavior-level proof; and
- an agent invents product policy when the request leaves a material choice
  open.

zharness provides a compact entrypoint, a navigable repository map, durable
plans only when work needs them, and playbooks that stay reduced for read-only
and bounded work.

## Goals

- **The repository stays the system of record.** Plans, decisions, and
  validation live in tracked markdown that a human can read and git can
  history. `harness.db` is a derived index, reconstructible from committed
  content by `zharness db rebuild`.
- **Process proportional to the work.** A read-only question and a
  multi-session refactor should not cost the same ceremony. Reduced playbook
  paths write no lifecycle rows; durable plans exist only for work that needs
  recovery context.
- **Invariants enforced, not assumed.** Where a rule can be checked it is
  checked — `validate` and the plan guards fail on violation rather than
  letting a broken state travel silently.
- **Portable across agents.** The spine skills are thin triggers; the
  operating logic sits in playbooks the CLI scaffolds into the repository.
  Any agent that reads a file and runs a CLI follows the same protocol.
- **Diagnostics that name the next action.** A finding states the violating
  item, the rule it breaks, its authority, and what to do — never a bare
  validation failure.
- **Safe to adopt and to leave.** `install`/`update`/`uninstall` manage only
  the doc set, merging rather than clobbering the files a project owns.

## Non-goals

- **Not a task database, tracker, or orchestrator.** zharness scaffolds and
  checks documents. It does not run the lifecycle, assign work, or drive an
  agent.
- **`harness.db` is not durable memory.** It is gitignored and per-machine,
  disposable by construction. Anything that must outlive the working copy
  belongs in tracked markdown or in git history.
- **No hosted or shared state.** The CLI makes no network calls. Everything
  the harness knows lives in the working copy.
- **Not a replacement for git.** The harness records intent and validation;
  git remains the record of what changed.
- **No derived-fact documents.** Routes, environment variables, and file
  inventories are not hand-maintained in `docs/` — the code is authoritative
  for what can be re-derived from it.
- **No automatic remediation.** Gates report verdicts and findings; deciding
  what to do about them stays with the operator.

## Default workflow

```text
read-only request
  -> inspect the smallest authoritative surface
  -> answer with evidence

bounded change
  -> inspect authority and affected behavior
  -> implement the smallest coherent change
  -> run relevant proof

multi-session or coordinated change
  -> create docs/plans/active/<plan>.md
  -> keep decisions, progress, recovery, and validation current
  -> move the validated plan to docs/plans/completed/

material product ambiguity
  -> stop before mutation
  -> present the concrete choice and consequences
```

A typo does not need a plan. A migration spanning sessions does.

## What gets installed

The managed set is:

- a compact `AGENTS.md` entrypoint (marked `ZHARNESS` block only);
- `docs/WORKFLOW.md` and the six stage playbooks;
- a `docs/PROJECT.md` identity scaffold;
- `.zharness/base/` for update tracking (fresh-overwrite for playbooks/WORKFLOW.md, three-way merge for PROJECT.md and the AGENTS.md block).

It does not write `CLAUDE.md`. It does not install application architecture,
product policy, skills, git hooks, credentials, a database, schemas,
orchestration, or background processes.

## Install

From a target repository, with `zharness` on PATH:

```bash
zharness install
```

Get the binary once per machine (`gh` + `tar`; Linux or macOS, amd64 or arm64):

```bash
bash scripts/install-zharness.sh
zharness --version
```

`install` is idempotent. It records upstream hashes, prints a read-only
brownfield report, and exits 0. Use `--root <dir>` when cwd is not the
consumer repo.

## Maintain an installation

```bash
zharness update
zharness update --continue
zharness update --abort
zharness uninstall
```

`docs/WORKFLOW.md` and the stage playbooks are pure upstream mirrors: update
always overwrites them with the latest bytes, discarding any local edit with
no merge and no conflict. `docs/PROJECT.md` and the marked `AGENTS.md` block
still three-way-merge against the exact upstream base under `.zharness/base/`;
if local and upstream edits overlap there, it stops with conflict markers.
After a human resolves them, `--continue`. `--abort` restores the pre-update
bytes. Uninstall removes managed files only; consumer-owned bytes are never
deleted.

## Optional skills

Skills are not part of `zharness install`. They live in this source repository:

```bash
npx skills add git@github.com:therealtinhtute/mono-harness.git -a claude-code -g -y
```

No skill runs during installation.

## v0.16

Protocol on the v0.15 three-verb binary: absorb at handoff close, at most
one active plan, independent judge for `full` checks. Pin `v0.14.x` to keep
the old lifecycle CLI. Existing `harness.db` files are consumer-owned;
nothing here deletes them.

See [`cli/docs/CONTRACT.md`](cli/docs/CONTRACT.md) and
[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## v0.15

Breaking cut: the lifecycle CLI and SQLite were deleted. The three verbs
remain. Pin `v0.14.x` if you still need that binary.

## Development

```bash
cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...
bash scripts/verify-doc-links.sh
bash scripts/test-guards.sh
```

This repository is the source of the binary, the embedded playbooks, and the
skills. Edit playbooks in `cli/docs/embedded/playbooks/`, then copy to
`docs/playbooks/`. Machine-wide Claude Code bootstrap is `setup/install.sh`;
that is not how a consumer app repo receives zharness.

## Contributing and security

Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.
Report vulnerabilities privately through [SECURITY.md](SECURITY.md).

## License

MIT

# zharness contract — v0.16

## What the binary is

`zharness` is an installer/updater binary. v0.15 deleted the lifecycle
command surface and the derived-database layer from source. v0.16 does not
change that surface. The binary registers exactly three verbs — `install`,
`update`, and `uninstall` — as listed below and in `cli/internal/interfaces/root.go`.

The command list in this contract therefore mirrors exactly what
`cli/internal/interfaces/root.go` registers, and vice versa.

## Command surface (unchanged since v0.15)

| Verb | Behavior |
|---|---|
| `install` | Projects the managed doc set (WORKFLOW, playbooks, AGENTS block, PROJECT identity scaffold, ignore entries), records upstream hashes under `.zharness/base/`, and prints a deterministic read-only brownfield report. Idempotent; exits 0. |
| `update` | Three-way merge of the managed set against `.zharness/base/` (BASE = upstream at install/last finalize). Clean regions auto-merge; overlapping edits stop with in-file conflict markers and a non-zero exit; `--continue` finalizes resolved files; `--abort` restores the pre-update state byte-for-byte from the stash. Consumer files outside the managed set are never touched. |
| `uninstall` | Removes managed files only. A file locally modified beyond both its base and upstream is left in place with a warning; consumer-owned bytes are never deleted (R12). |

Shared flags: `--root <dir>` overrides the target repository (default: git toplevel of the working directory).

## Where the guarantees live now

Fail-closed guarantees live in the repository's **pre-commit hook**, which
reads staged bytes and trusts no marker an agent writes:

1. **Proof re-execution** — a newly added `## Validation` entry whose verdict
   is `APPROVED` or `APPROVE_WITH_REQUESTS` must have every nested proof
   command re-executed by the hook (`sh -c`, 5-minute timeout each, exit 0);
   any failure rejects the commit naming the command and its output tail.
   Proof of a `REQUEST_CHANGES` entry is never re-executed.
2. **Independent judge (high-risk)** — a newly added entry carrying a
   same-session judge declaration into a plan whose frontmatter sets
   `lane: high-risk` is rejected; lane is read from plan frontmatter directly.
3. **Independent judge (full)** — a newly added Validation entry whose first
   line declares `mode: full` and which also carries a same-session judge is
   rejected, on every lane.
4. **At most one active plan** — more than one non-empty file under
   `docs/plans/active/` is rejected; zero is a valid idle state.

Handoff absorb is playbook protocol, not a hook: close cannot `git mv`
without an `absorb:` line. Missing absorb is not a commit rejection.

These guards are implemented once in `scripts/install-git-hooks.sh` between
the `# ZGUARD-CORE-BEGIN` / `# ZGUARD-CORE-END` markers; the local hook and
the CI job both extract that block verbatim, so neither can drift from the
other. `.github/workflows/cli-ci.yml` re-runs them against pushed commits,
covering anyone who bypasses local hooks.

Tool allow/deny is the **host** agent runtime (Claude/Pi/Codex), not this
binary. Default shape the host should own: READ files allowed; RUN TESTS
inside a sandbox; WRITE inside the workspace; NETWORK scoped by task;
DEPLOY and DELETE DATA require approval. zharness authorizes only
managed-doc `install`/`update`/`uninstall` and, via the hook, commits of
clean Validation entries.

Repository scripts:

- `scripts/install-git-hooks.sh` — installs the hooks (also embeds the guard core)
- `scripts/record-check.sh` — optional convenience runner for proof commands;
  it holds no guarantee
- `scripts/plan-slice.sh` — optional plan-section reader; holds no guarantee

## What carries the state

Git-committed markdown alone: plans under `docs/plans/{active,completed}/*.md`
with append-only Progress, Decisions, and Validation sections; ADRs under
`docs/decisions/`; session memory as plain files under `docs/memory/`
(convention). The stage playbooks under `cli/docs/embedded/playbooks/` are
the canonical procedures, projected to `docs/playbooks/`.

The previous derived database was an index whose only consumer — the CLI
lifecycle surface deleted above — is gone. What it was, why it was removed,
and where consumers should pin: see the v0.15 section of the root
`CHANGELOG.md`, the archive of record for this removal.

## Breaking change

v0.15 is a breaking release. Consumers should pin the last 0.14.x release to
keep the previous lifecycle binary working; see the v0.15 section of the root
`CHANGELOG.md`.

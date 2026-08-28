# zharness contract — v0.15 (slim)

## What the binary is

`zharness` is being reduced to an installer/updater binary. This release
(v0.15) deletes the entire lifecycle command surface and the derived-database layer from source;
the `install`, `update`, and `uninstall` verbs land in phase
`p3-installer` of `docs/plans/active/zharness-v015-slim.md`. Until then the
binary registers no subcommands — `zharness --help` shows root usage only.

The command list in this contract therefore mirrors exactly what
`cli/internal/interfaces/root.go` registers, and vice versa.

## Command surface (v0.15)

| Verb | Behavior |
|---|---|
| `install` | Projects the managed doc set (WORKFLOW, playbooks, AGENTS block, PROJECT identity scaffold, ignore entries), records upstream hashes under `.zharness/base/`, and prints a deterministic read-only brownfield report. Idempotent; exits 0. |
| `update` | Three-way merge of the managed set against `.zharness/base/` (BASE = upstream at install/last finalize). Clean regions auto-merge; overlapping edits stop with in-file conflict markers and a non-zero exit; `--continue` finalizes resolved files; `--abort` restores the pre-update state byte-for-byte from the stash. Consumer files outside the managed set are never touched. |
| `uninstall` | Removes managed files only. A file locally modified beyond both its base and upstream is left in place with a warning; consumer-owned bytes are never deleted (R12). |

Shared flags: `--root <dir>` overrides the target repository (default: git toplevel of the working directory).

## Where the guarantees live now

The two fail-closed guarantees moved out of the binary into the repository's
**pre-commit hook**, which reads staged bytes and trusts no marker an agent
writes:

1. **Proof re-execution** — a newly added `## Validation` entry whose verdict
   is `APPROVED` or `APPROVE_WITH_REQUESTS` must have every nested proof
   command re-executed by the hook (`sh -c`, 5-minute timeout each, exit 0);
   any failure rejects the commit naming the command and its output tail.
   Proof of a `REQUEST_CHANGES` entry is never re-executed.
2. **Independent judge** — a newly added entry carrying a same-session judge
   declaration into a plan whose frontmatter sets `lane: high-risk` is
   rejected; lane is read from plan frontmatter directly.

Both guards are implemented once in `scripts/install-git-hooks.sh` between
the `# ZGUARD-CORE-BEGIN` / `# ZGUARD-CORE-END` markers; the local hook and
the CI job both extract that block verbatim, so neither can drift from the
other. `.github/workflows/cli-ci.yml` re-runs them against pushed commits,
covering anyone who bypasses local hooks.

Repository scripts:

- `scripts/install-git-hooks.sh` — installs the hooks (also embeds the guard core)
- `scripts/record-check.sh` — optional convenience runner for proof commands;
  it holds no guarantee

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

# Architecture

<!-- zharness:pin 2d4bc5a2015113cd85c3134b04bffc18f94887d1 -->

How the harness actually works as of v0.16 (binary surface from the v0.15 "slim" cut), and why it is shaped this way. Decisions that are expensive to reverse have their own records under `docs/decisions/`; this document describes the running system.

## The one idea

Committed markdown is the truth. The binary is a scaffolding tool, not a runtime.

`zharness` owns exactly three verbs — install, update, uninstall — that manage a doc set inside a consumer git repository. The whole lifecycle (brainstorm → to-plan → work → check → handoff) runs from the committed markdown, repo scripts, and a pre-commit hook, with the binary absent from PATH. Delete the binary and every derived artifact, and the same lifecycle still executes, because the only irreplaceable bytes are the ones git is already tracking.

## The three verbs

```
cli/cmd/zharness                 one Go binary (cobra), three verbs
        |
        v
install                          scaffolds the managed set, records a base, reports brownfield read-only
update                           three-way merge of managed files onto the recorded base
uninstall                        removes the managed set; consumer bytes are never destroyed
```

`AllTargets` (`cli/internal/installer/installer.go:65`) is the managed set: `docs/WORKFLOW.md`, `docs/PROJECT.md` (scaffolded from the identity template), and the six playbooks. `AGENTS.md` is handled separately and surgically — only the marked `ZHARNESS` block inside it is swapped; consumer prose around it is untouched.

State lives in `.zharness/base/`: a `manifest.json` of `{path, sha256}` entries plus content-addressed upstream blobs. Update diffs local edits against that recorded base with a diff3 merge (`cli/internal/installer/threeway.go`); overlapping edits stop with in-file conflict markers, resolved by a human via `--continue` (records the conflict-time upstream as the new base; the resolution stays as local drift) or discarded via `--abort` (stash restore, byte-for-byte). Uninstall deletes only wholly-created files, restores captured pre-install originals, and keeps anything locally modified with a warning.

The installer is also the onboarding probe: `install` prints a deterministic, read-only brownfield report (active-plan count, present consumer inputs, foreign state files) and exits 0 without writing outside the managed set. `docs/PROJECT.md` is the identity record; filling it is the single forced write step at brainstorm lock (`docs/playbooks/brainstorm.md` step 6).

## The fail-closed guards

Fail-closed guarantees live in the pre-commit hook (`scripts/install-git-hooks.sh`, shared core between the `# ZGUARD-CORE` markers):

1. **Proof re-execution** — a newly added `## Validation` entry with verdict `APPROVED` (or `APPROVE_WITH_REQUESTS`) has every nested proof command re-executed by the hook; any non-zero exit rejects the commit.
2. **Independent judge (high-risk)** — on a `lane: high-risk` plan, a new Validation entry carrying `judge: same-session` is rejected.
3. **Independent judge (full)** — a newly added Validation entry that declares `mode: full` and `judge: same-session` is rejected, on every lane.
4. **At most one active plan** — more than one non-empty file under `docs/plans/active/` is rejected; zero is a valid idle state.

`.github/workflows/cli-ci.yml` re-runs these checks on pushed commits, so bypassing local hooks gains nothing. The hook reads staged bytes itself; there is no pass marker an authoring agent could forge.

Handoff absorb is playbook protocol, not a hook: final close writes `absorb: none` or names an existing ADR/guard/memory. An unabsorbed class-of-failure stops before `git mv`.

Tool allow/deny is the host runtime’s job, not zharness: READ files by default, RUN TESTS in a sandbox, WRITE inside the workspace, NETWORK scoped by task, DEPLOY and DELETE DATA behind approval. This binary only installs the doc set and the hook only gates Validation commits.

## Embedded doc set and projection

The binary carries two embedded filesystems (`cli/docs/embedded/`): the managed set (`AGENTS.md`, `WORKFLOW.md`, `playbooks/`) and, separately, `templates/` — emitted on demand, never projected. `docs/playbooks/*.md` is a byte-identical projection of the embedded source: edit `cli/docs/embedded/`, never the projection.

## Markdown-only lifecycle state

| Path | Nature |
|---|---|
| `docs/plans/active/*.md` | authoritative — the one active initiative; append-only `## Progress` / `## Decisions` / `## Validation` |
| `docs/PROJECT.md` | authoritative — identity, answered at the brainstorm lock |
| `docs/memory/*.md` | memory as files; agents grep directly (`docs/memory/{id}.md`) |
| `docs/decisions/`, `docs/research/`, `docs/audit/` | authoritative records |
| `docs/playbooks/`, `docs/WORKFLOW.md` | projected; edit `cli/docs/embedded/` instead |
| `.zharness/` | installer bookkeeping (base manifest + blobs); gitignored |

Task execution status lives only in append-only `## Progress`; task definitions carry no status fields. Bookkeeping is hand-appended — the binary writes no plan rows.

## Historical note (pre-v0.15, explicitly historical)

The 0.14.x architecture — a `preflight <stage>` entry point, a SQLite store under `harness.db` serving as a derived index (`managed_docs`, `plan_index`, `memories` tables, `db rebuild`), conflict staging under `.kit/`, and CLI proof verification via `check record` — was deleted in v0.15. Consumers who depend on it can pin the 0.14.x release; its proof-verification contract survives verbatim in the pre-commit hook (see [`cli/docs/CONTRACT.md`](../cli/docs/CONTRACT.md)).

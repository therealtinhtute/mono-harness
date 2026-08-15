# Field Audit — onedrive-cloud

**Read-only audit, 2026-08-15.** `zharness 0.9.1`, `db: ready`, `readiness: clean`.
Purpose: establish that every defect the spec fixes is real, not hypothetical.

---

## What Is On Disk

| Layer | Files | Size | Tracked |
|---|---|---|---|
| Root instructions | `CLAUDE.md` 12.7KB, `KNOWNS.md` 9.8KB, `AGENTS.md` 1.8KB | 24KB | yes |
| Product spec | `SPEC.md` 18.9KB, `README.md` 8.6KB | 27KB | yes |
| `docs/` reference | pdr 11.7KB, system-architecture 24.7KB, codebase-summary 15.9KB, code-standards 15.2KB, kv-config 2.3KB | 68KB | yes |
| `docs/playbooks/` | 7 stage playbooks | 52KB | yes (6 modified, `git.md` untracked) |
| `docs/plans/` | active 29.8KB + completed | 60KB | partial |
| `.kit/planning/` **legacy** | SPEC + ROADMAP + 11 phases x (PLAN+CONTEXT) | ~60KB | **gitignored** |
| `.kit/plans/` third location | `2026-08-15-walter-theme/plan.md` | 9.3KB | **gitignored** |
| `.kit/runs` + `.kit/reports` | 4 runs, 2 reports | ~15KB | **gitignored** |
| `harness.db` + 39 changesets | zharness state | 248KB | **gitignored** |

**7,895 lines of markdown, ~45k words, ~60k tokens.**
Product docs ~96KB, process artifacts ~124KB. **Process outweighs product 1.3 : 1.**

---

## D1 — Two State Pointers Disagree By A Month

```
.kit/workflow-state.yml  -> current_phase: credential-containment   (2026-07-18)
zharness resume --json   -> current_phase: walter-theme-migration, status: done
```

Generation 1 (`workflow-state.yml` + `.kit/planning/` + `.kit/runs/` + `.kit/reports/`)
was never removed when zharness landed. An agent reading `AGENTS.md` follows zharness;
an agent that greps finds the July state. Both present as authoritative.

**Fixed by:** M6 — `audit`/`watzup` detect and name legacy generations.

---

## D2 — Three Plan Locations, All Live

| Location | Tracked | Origin |
|---|---|---|
| `docs/plans/active/check-review-remediation.md` | yes | project convention |
| `.kit/planning/phases/*/` (22 files) | no | Gen-1 legacy |
| `.kit/plans/2026-08-15-walter-theme/plan.md` | no | global `~/.claude/CLAUDE.md` rule |

The global rule `Plans -> .kit/plans/{date}-{slug}/` contradicts the project convention
`docs/plans/active/{slug}.md`. The newest and most rigorous plan (walter-theme — it
carries a verified `bunx shadcn add` dry-run table) sits in a gitignored directory.

**Fixed by:** M1 — one tracked plan location under `.harness/plans/`.

---

## D3 — Durable Product Authority Is Gitignored

`.kit/planning/SPEC.md` holds a locked spec with 8 risk flags and a `## Key Decisions`
section in Chosen/Rejected form with rationale:

```
Chosen:   incremental six-goal migration, because it keeps each security and
          architecture slice independently verifiable and mergeable.
Rejected: full rewrite, because it increases regression risk and delays containment.
Rejected: plaintext/replayable URL or local-storage credentials, because they cross
          logs, history, and browser boundaries.
```

That is ADR content. `.gitignore` excludes `.kit` wholesale. `zharness init` rebuilds
`.kit` from changesets; these hand-written files are not in changesets. One `rm -rf .kit`
from gone, and invisible to any other machine.

**Fixed by:** M1 (tracked `.harness/`) + M4 (ADR promotion).

---

## D4 — Decisions Never Reach Durable Storage

`zharness decision add` writes to `harness.db` and `.kit/changesets/` — **both gitignored.**
`handoff.md:25` pulls `query decisions --tail 10 --json` into the plan markdown's
Decisions section, but there is no `docs/decisions/` path anywhere in the CLI.

On a fresh clone or a second machine, decision history is gone unless it happened to
land in a plan file, where it is buried in a 400-line document beside task checklists.

**Fixed by:** M4 — `decision promote` writing numbered ADRs, gated at handoff.

---

## D5 — `KNOWNS.md` Declares A False Precedence, And Is Orphaned

```
KNOWNS.md: "AGENTS.md, CLAUDE.md ... are compatibility shims"
Reality:    CLAUDE.md is 12.7KB — the largest instruction file, holding
            architecture, API patterns, and styling found nowhere else.
            AGENTS.md carries the ZHARNESS:BEGIN block.
```

Grep confirms **nothing references `KNOWNS.md`** — not `AGENTS.md`, not `CLAUDE.md`,
not zharness. The `<!-- KNOWNS GUIDELINES START -->` markers it mandates exist nowhere.
`.gitignore:47` has a dead `.knowns/` entry. 184 lines that only ever cite themselves.

**Fixed by:** M7 — `audit` flags instruction files >4KB unreferenced by `AGENTS.md`.

---

## D6 — A CLI Upgrade Silently Rewrote Tracked Files

```
docs/playbooks/work.md   | 58 ++++++++-------------  32 insertions, 26 deletions
docs/playbooks/check.md  | 20 ++++-----------         9 insertions, 11 deletions
docs/playbooks/watzup.md | 14 +++++-------            7 insertions,  7 deletions
```

Diffing the worktree against `cli/docs/embedded/playbooks/` — **byte-identical.**
`SyncManagedDocs` overwrote committed files on a version bump. The owner authored none
of it. Six of seven playbooks dirty, `git.md` untracked, on every upgrade.

**Fixed by:** M5 — never overwrite a tracked file without `--force`; stage to
`.harness/state/conflicts/*.upstream` and report.

---

## D7 — Hand-Maintained Derived Facts Rot

Every file in `docs/` stamped **`Last Updated: 2025-12-26`** — 8 months stale.

`docs/codebase-summary.md` is a repomix dump committed as documentation:

```
- Total Files: 190 files
- Total Lines: 16,585 lines
- Total Tokens: 138,537 tokens
```

Machine-derived facts, hand-maintained, with no regeneration path. Guaranteed to rot.
`docs/system-architecture.md` (664 lines) has the same problem for its derived half.

Everything they contain is mechanically derivable **without an AST**:

| Fact | Source | Method |
|---|---|---|
| 15 API routes + 3 pages | `src/app/**/route.ts`, `page.tsx` | file walk (App Router = filesystem routing) |
| 11 env vars | `process.env.X` | regex |
| Coverage | `coverage/coverage-final.json` (exists) | parse JSON |
| Import graph / blast radius | `from '...'` | regex, ~95% accurate |
| Subsystems | `src/components/{auth,drives,layout,previews,ui}`, `src/store` | file walk |

**Fixed by:** M2 — `zharness wiki`, deterministic, 0 tokens.

---

## Token Compute

| Moment | Loaded | Tokens |
|---|---|---|
| Session start (auto) | `CLAUDE.md` | ~3.2k |
| Following `KNOWNS.md` precedence | + `KNOWNS.md`, `AGENTS.md`, `docs/WORKFLOW.md` | ~6.6k |
| Enter `work` stage | + `playbooks/work.md` + active plan (29.8KB) | **~17k** |
| Enter `check` stage | + `playbooks/check.md` + plan | ~16k |

**~17k tokens of process text before reading one line of `src/`.**

Comparison — harness-experimental resident cost: `AGENTS.md` (1.6KB) + `docs/WORKFLOW.md`
(5KB) = **~1.6k tokens**, roughly 10x lighter, with `docs/README.md` as a 40-line map.

Largest single offender: `docs/plans/active/check-review-remediation.md` at
**29.8KB / 410 lines**, re-read at every stage.

---

## Stale / Orphaned Inventory

- `docs/*.md` — all stamped 2025-12-26, ~8 months stale
- `docs/codebase-summary.md` — generated dump with no maintenance path
- `.claude/RESUME.md` — a crash checkpoint sitting in the tree
- `docs/plans/active/theme-font-header-refresh.md` deleted but unstaged;
  `docs/plans/completed/` untracked — completion move happened on disk, never committed
- `docs/code-standards.md` (15KB) duplicates `AGENTS.md` § Code Style and `CLAUDE.md`
- `docs/plans/active/check-review-remediation.md` still `status: active` from 2026-07-30
  while `resume` reports walter-theme-migration/done — a second stale-active plan
- Three roadmaps that disagree: `SPEC.md` 7 phases, `project-overview-pdr.md` 4 phases,
  `.kit/planning/ROADMAP.md` 6 phases

# TINHTUTE Skills

A collection of personal Claude Code skills following the [skills.sh](https://skills.sh) ecosystem format.

## Structure

The repository has **14 skills** organized above a four-layer workflow runtime:

| Layer | Responsibility |
| :--- | :--- |
| **Harness** | Durable state: committed ULID changesets plus a rebuildable local SQLite database. |
| **Workflows** | Tool-independent lifecycle: intent → intake → plan → trace → proof → handoff/resume. |
| **Skills** | User-facing triggers, including the 8 workflow skills and 6 shipping/craft skills. |
| **CLI** | `zharness`, which scaffolds playbooks and reads or mutates harness state deterministically. |

### Harness artifact layout: `.kit/`

```text
.kit/
├── planning/
│   ├── IDEA.md
│   ├── SPEC.md
│   ├── ROADMAP.md
│   └── phases/{phase-slug}/
│       ├── {phase-slug}-CONTEXT.md
│       └── {phase-slug}-PLAN.md
├── changesets/         (committed ULID-named JSONL — replay source)
├── runs/work/          (execution logs)
├── reports/            (brainstorm, check, and optional work evidence)
├── HANDOFF.md          (latest continuity snapshot)
├── docs/               (generated playbooks — gitignored)
└── harness.db          (generated SQLite view — gitignored)
```

#### What lives where

- `planning/` locks the intent, roadmap, phase boundaries, and executable plans.
- `changesets/` is the replayable source for harness entities and pointers.
- `harness.db` is a local materialized view rebuilt from committed changesets.
- `docs/` contains generated canonical playbooks scaffolded by `zharness init`.
- `runs/work/` records execution; `reports/check/` records gate evidence and verdicts.
- `HANDOFF.md` captures the latest session handoff while `zharness resume --json` derives continuity from durable state.

#### Mental model

- `.kit/planning/` answers **what is locked**.
- `.kit/runs/` answers **what execution did**.
- `.kit/reports/` answers **what evidence and review concluded**.
- `zharness resume --json` answers **where to continue and whether recovery is needed**.

## Machine Setup

Bootstrap Claude Code configuration and install all repository skills:

```bash
git clone git@github.com:therealtinhtute/skills.git ~/skills
bash ~/skills/setup/install.sh
```

Install the required harness CLI separately. The six workflow spine skills require `zharness >= 0.4.1`:

```bash
bash ~/skills/scripts/install-zharness.sh
zharness --version
```

`setup/install.sh` does not install the CLI. After bootstrap, review `~/.claude/settings.json` and set `ANTHROPIC_AUTH_TOKEN` when needed.

**Skills only** (no config changes):

```bash
npx skills add git@github.com:therealtinhtute/skills.git -a claude-code -g -y
```

### What `install.sh` installs

| Component | Source | Target |
| :--- | :--- | :--- |
| Global CLAUDE.md | `setup/CLAUDE.md` | `~/.claude/CLAUDE.md` |
| Rules | `rules/*.md` | `~/.claude/rules/` |
| Hooks | `setup/hooks/` | `~/.claude/hooks/` |
| Settings template | `setup/settings.json` | `~/.claude/settings.json` (new machines only) |
| Statusline | `scripts/setup-statusline.sh` | `~/.claude/statusline.sh` |
| All skills | GitHub | `~/.claude/skills/` via npx |

### Individual installs

Install just a rule:
```bash
cp rules/karpathy-guidelines.md ~/.claude/rules/
cp rules/english.md ~/.claude/rules/
cp rules/ask-user-question.md ~/.claude/rules/
```

Install just the settings template:
```bash
cp setup/settings.json ~/.claude/settings.json
# Edit: set ANTHROPIC_AUTH_TOKEN
```

Install just the statusline:
```bash
bash scripts/setup-statusline.sh
```

## Skills

### Workflow — Agentic orchestration chain

| Skill | When | What it does |
| :--- | :--- | :--- |
| [`/brainstorm`](skills/workflow/brainstorm/SKILL.md) | Project bootstrap, feature scoping, ideation, architecture decisions | Turn an idea, notes, or markdown files into a locked `.kit/planning/SPEC.md` — exploring options and trade-offs along the way. |
| [`/to-plan`](skills/workflow/to-plan/SKILL.md) | After brainstorm, before implementation | Turn a locked `.kit/planning/SPEC.md` into a roadmap, per-phase context, and executable wave-based plans. |
| [`/work`](skills/workflow/work/SKILL.md) | "Implement this plan", "build it end-to-end" | Execution orchestrator after `brainstorm` + `to-plan`. Routes to upstream skills if artifacts are missing; runs phase waves; verifies every task; gates via `/check`. |
| [`/interview`](skills/workflow/interview/SKILL.md) | Validating plans before implementation | Interview about plans using AskUserQuestion. Explore technical decisions, UI/UX, concerns, tradeoffs. Write validated spec. |
| [`/check`](skills/workflow/check/SKILL.md) | Before commit, PR, or merge; phase gate after `/work` | Gate (tests, lint, build) + code review (security, architecture, quality). |
| [`/git`](skills/workflow/git/SKILL.md) | Staging, committing, pushing, PRs, merges | Git operations with conventional commits. Auto-splits commits by type/scope. Security scans for secrets. |
| [`/handoff`](skills/workflow/handoff/SKILL.md) | Session end, context switches, milestones | Capture session state and write HANDOFF.md for seamless continuation. |
| [`/watzup`](skills/workflow/watzup/SKILL.md) | Start of session, resuming work, quick status check | Recap branch state, committed + uncommitted changes, handoff context, and artifact chain — then recommend the next action. |

### Shipping — Build & ship code

| Skill | When | What it does |
| :--- | :--- | :--- |
| [`/create-cli`](skills/shipping/create-cli/SKILL.md) | Designing a CLI, greenfield or retrofit | Design CLI interfaces (commands, flags, I/O, errors, config) and produce implementation roadmaps with framework choice and shipping strategy. |
| [`/turbo-mono-platform`](skills/shipping/turbo-mono-platform/SKILL.md) | Working on the monorepo stack | Full-stack TypeScript monorepo guidance (Turborepo, Next.js, Hono, tRPC, Drizzle, etc.). |

### Craft — Research, writing, meta-skills

| Skill | When | What it does |
| :--- | :--- | :--- |
| [`/write`](skills/craft/write/SKILL.md) | Writing, rewriting, polishing prose | Edit or write English/Vietnamese prose so it sounds natural, concise, and context-aware across docs, UI, reports, and marketing copy. |
| [`/librarian`](skills/craft/librarian/SKILL.md) | Researching external GitHub repos | GitHub code research via gh CLI. Find symbols, grep code, gather evidence without cloning. |
| [`/create-skill`](skills/craft/create-skill/SKILL.md) | Creating or updating Claude skills | Create or update Claude skills optimized for Skillmark benchmarks. |
| [`/prompt-leverage`](skills/craft/prompt-leverage/SKILL.md) | Improving prompts, building frameworks | Strengthen raw user prompts into execution-ready instruction sets for AI agents. |

## Response convention

This repo follows a shared output convention inspired by Waza for active skills:

- the first response line should start inline with `🥷`
- the opening line should contain the verdict, recommendation, state, or next move
- outputs should be sharp, compressed, and decision-useful
- generic "default assistant" filler is considered a quality failure

The icon is the visible mode switch. The real standard is the writing: concrete, direct, and specific to the skill.

## Usage flow

Skills are the user-facing interface. `zharness` persists intake, stories, runs, traces, verdicts, and handoffs underneath them.

![zharness-backed workflow for full, simple, and resume paths](assets/spec-plan-workflow.svg)

[Open the self-contained diagram source](assets/workflow-usage-flow.html).

### Initialize a project once

```bash
cd /path/to/project
zharness init --json
```

`init` creates `.kit/` when needed, initializes the local database, scaffolds missing `.kit/docs/` playbooks, and adds the generated harness paths to `.gitignore` as applicable.

### Full work: lock, plan, execute, prove

```text
/brainstorm <idea, notes, or @file references>
/to-plan full
/work
```

1. `brainstorm` explores alternatives, writes `.kit/planning/SPEC.md`, and records the intake.
2. `to-plan` writes the roadmap and phase context/plans, then records one story per phase.
3. `work` executes the active phase wave-by-wave and records the run plus trace evidence.
4. `check` gates each completed phase with tests, review, proof scoring, and a verdict.
5. `git` commits or ships clean work; `handoff` records continuity when the session pauses or transfers.

Use `/interview` before locking or planning when the intent needs structured clarification.

| Skill | Human-readable artifact | Durable harness record |
| :--- | :--- | :--- |
| `brainstorm` | `SPEC.md` | intake |
| `to-plan` | `ROADMAP.md` + phase plans | story |
| `work` | `runs/work/*.md` | run + traces |
| `check` | `reports/check/*.md` | proof + verdict |
| `handoff` | `HANDOFF.md` | handoff |
| `watzup` | console recap | resume snapshot |

### Small, bounded work

```text
/work simple <concrete task>
```

Use simple mode only for a known subsystem within five files and roughly 100 changed lines. It skips phase planning and deliberately does not create phase-bound database run/check rows; run a lightweight verification or `/check` when the change warrants it.

### Resume by readiness

```text
/watzup
```

`watzup` is read-only. It renders `zharness resume --json`, checks branch and artifact continuity, then recommends one next action: initialize or import, recover drift, continue work, run the gate, or commit. The recommendation follows readiness instead of assuming the next step.

### Adopt a legacy project

For a project that already has markdown-only `.kit/` history:

```bash
zharness init --json
zharness import --json
zharness query state --json
zharness validate --json
zharness audit --json
```

Do not use `import` for a new project. Full migration and rollback guidance: [`docs/workflow-harness/migration.md`](docs/workflow-harness/migration.md).

### Harness implementation model

The six spine skills (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) are thin triggers. They version-gate on `zharness`, ensure generated docs exist, then follow the matching canonical playbook under `.kit/docs/playbooks/`. `interview` and `git` remain sidecars rather than harness entity owners.

See [`skills/workflow/README.md`](skills/workflow/README.md) for the four-layer concept and [`docs/workflow-harness/`](docs/workflow-harness/) for migration notes and pilot evidence.

## Local Development

This repo is the stable release. The incubator workspace is a local clone at a path of your choice.

To iterate quickly on a skill before publishing, point Claude Code at your local skill directory:

```bash
claude-code add-dir /path/to/your/local/skill
```

To sync changes from an incubator to this repo:

```bash
bash scripts/sync-from-kit.sh
```

## License

MIT

# TINHTUTE Skills

A collection of personal Claude Code skills following the [skills.sh](https://skills.sh) ecosystem format.

## Structure

![Layer stack — 3 abstraction layers of the repository](assets/diagram-layer-stack.png)

### Harness artifact layout: `.kit/`

```text
.kit/
├── planning/
│   ├── IDEA.md
│   ├── SPEC.md
│   ├── ROADMAP.md
│   └── phases/
│       └── {phase-slug}/
│           ├── {phase-slug}-CONTEXT.md
│           └── {phase-slug}-PLAN.md
├── harness.db          (gitignored — rebuilt from changesets)
├── changesets/         (ULID-named JSONL, source of truth)
├── HANDOFF.md
├── runs/
│   └── work/
└── reports/
    ├── brainstorm/
    ├── check/
    └── watzup/
```

#### What lives where
- `planning/` — canonical planning artifacts owned by `brainstorm` and `to-plan`
- `harness.db` + `changesets/` — durable, queryable state (`zharness query state`/`resume`); replaces the retired `workflow-state.yml` pointer file — see [Workflow Harness](#workflow-harness)
- `HANDOFF.md` — latest continuity snapshot for the next session
- `runs/work/` — execution logs created by `work`
- `reports/check/` — persisted gate verdicts from `check`
- `reports/watzup/` — recap reports from `watzup` (legacy; current version is console-only)
- `reports/brainstorm/` — optional explore-mode output when `brainstorm` is used without locking a spec

#### Mental model
- `.kit/planning/` answers **what is currently locked**
- `.kit/runs/` answers **what happened during execution**
- `.kit/reports/` answers **what the gate concluded**
- `zharness resume --json` answers **where the harness should look first**

## Machine Setup

Full bootstrap for a new machine — installs CLAUDE.md, rules, hooks, statusline, and all skills:

```bash
git clone git@github.com:therealtinhtute/skills.git ~/skills
bash ~/skills/setup/install.sh
```

After install: edit `~/.claude/settings.json` and set `ANTHROPIC_AUTH_TOKEN`.

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

## Recommended workflow: `brainstorm` + `to-plan` + `work` + friends

Two entry points (`watzup` for resume, `brainstorm` for new work), then execute and close out.

![Recommended workflow for brainstorm + plan + work with supporting skills](assets/spec-plan-workflow.svg)

Canonical pipeline:
```
watzup → work → check → git → handoff          (resume)
brainstorm → to-plan → work → check → git → handoff  (new work)
```

### 1. Lock the problem with `brainstorm`
Use `brainstorm` when you have a raw idea, notes, markdown files, or a trade-off question. It runs in 4 modes (`explore`, `lock-from-idea`, `lock-from-files`, `refine`) and produces either a recommendation report or a locked `.kit/planning/SPEC.md`.

### 2. Derive execution with `to-plan`
Use `to-plan` only after the spec is locked. It turns `.kit/planning/SPEC.md` into `.kit/planning/ROADMAP.md` plus per-phase `-CONTEXT.md` and `-PLAN.md` files. If the spec is missing or too weak, `to-plan` fails fast and points back to `brainstorm`.

### 3. Execute with `work`
Use `work` to execute the plan. It checks for missing artifacts and routes back to `brainstorm` or `to-plan` if needed; otherwise it runs the active phase wave-by-wave, dispatches subagents for heavy tasks, verifies every task, and calls `/check` as the phase gate. It never auto-commits — handoffs are suggested, not executed.

### 0. Orient with `watzup` (resume path)
Use `watzup` at the start of a session to recap branch state, review committed + uncommitted changes, read handoff context, and get a concrete next action. If the branch has no work yet, `watzup` points you to `brainstorm`.

### 4. Close the loop
- [`/check`](skills/workflow/check/SKILL.md) is invoked by `work` per phase, or directly for ad-hoc gate/review
- [`/git`](skills/workflow/git/SKILL.md) and [`/handoff`](skills/workflow/handoff/SKILL.md) close or transfer a work session cleanly

### Mental model
- `watzup` = recap **WHERE AM I** (session start)
- `brainstorm` = lock **WHAT**
- `to-plan` = lock **HOW**
- `work` = **execute** (run phases, verify, gate)
- `check` = catch risk and prove readiness (gate + analysis)
- `git` / `handoff` = wrap up with discipline

## Workflow Harness

The workflow chain above is evolving from prompt-only orchestration into a harness-backed runtime (`zharness` CLI, durable state, deterministic gates). See [`skills/workflow/README.md`](skills/workflow/README.md) for the concept doc and [`docs/workflow-harness/`](docs/workflow-harness/) for the gap inventory. Initiative in progress — chain UX above is unaffected.

### Quickstart: `zharness`

```bash
# install (no cli/v* release yet — see docs/workflow-harness/migration.md#install for the current build-from-source path)
bash scripts/install-zharness.sh

# new project (init does not create the .kit/ directory itself)
mkdir -p .kit
zharness init --json
zharness story --slug my-phase --goal "..." --json

# existing project with legacy .kit/workflow-state.yml
zharness init --json
zharness import --json
zharness query state --json
```

Full install/import/validate/rollback walkthrough, proven on this repo's own real history: [`docs/workflow-harness/migration.md`](docs/workflow-harness/migration.md). Pilot evidence and go/no-go verdict: [`skills/workflow/README.md`](skills/workflow/README.md#pilot-evidence--gono-go).

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

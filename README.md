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
├── workflow-state.yml
├── HANDOFF.md
├── runs/
│   └── cook/
└── reports/
    ├── brainstorm/
    ├── check/
    └── watzup/
```

#### What lives where
- `planning/` — canonical planning artifacts owned by `brainstorm` and `plan`
- `workflow-state.yml` — lightweight pointer index for the current phase and latest downstream artifacts
- `HANDOFF.md` — latest continuity snapshot for the next session
- `runs/cook/` — execution logs created by `cook`
- `reports/check/` — persisted gate verdicts from `check`
- `reports/watzup/` — retrospective summaries from `watzup`
- `reports/brainstorm/` — optional explore-mode output when `brainstorm` is used without locking a spec

#### Mental model
- `.kit/planning/` answers **what is currently locked**
- `.kit/runs/` answers **what happened during execution**
- `.kit/reports/` answers **what the gate or retrospective concluded**
- `.kit/workflow-state.yml` answers **where the harness should look first**

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

| Skill | When | What it does |
| :--- | :--- | :--- |
| [`/bash-tui`](skills/bash-tui/SKILL.md) | Building interactive terminal UIs | Build bash/shell TUI apps with menus, selectors, forms, progress bars, spinners, banners, and color output. |
| [`/brainstorm`](skills/brainstorm/SKILL.md) | Project bootstrap, feature scoping, ideation, architecture decisions | Turn an idea, notes, or markdown files into a locked `.kit/planning/SPEC.md` — exploring options and trade-offs along the way. |
| [`/git`](skills/git/SKILL.md) | Staging, committing, pushing, PRs, merges | Git operations with conventional commits. Auto-splits commits by type/scope. Security scans for secrets. |
| [`/handoff`](skills/handoff/SKILL.md) | Session end, context switches, milestones | Capture session state and write HANDOFF.md for seamless continuation. |
| [`/interview`](skills/interview/SKILL.md) | Validating plans before implementation | Interview about plans using AskUserQuestion. Explore technical decisions, UI/UX, concerns, tradeoffs. Write validated spec. |
| [`/librarian`](skills/librarian/SKILL.md) | Researching external GitHub repos | GitHub code research via gh CLI. Find symbols, grep code, gather evidence without cloning. |
| [`/plan`](skills/plan/SKILL.md) | After brainstorm, before implementation | Turn a locked `.kit/planning/SPEC.md` into a roadmap, per-phase context, and executable wave-based plans. |
| [`/prompt-leverage`](skills/prompt-leverage/SKILL.md) | Improving prompts, building frameworks | Strengthen raw user prompts into execution-ready instruction sets for AI agents. |
| [`/check`](skills/check/SKILL.md) | Before commit, PR, or merge; phase gate after `/cook` | Gate (tests, lint, build) + code review (security, architecture, quality). |
| [`/cook`](skills/cook/SKILL.md) | "Implement this plan", "build it end-to-end" | Execution orchestrator after `brainstorm` + `plan`. Routes to upstream skills if artifacts are missing; runs phase waves; verifies every task; gates via `/check`. |
| [`/skill-creator`](skills/skill-creator/SKILL.md) | Creating or updating Claude skills | Create or update Claude skills optimized for Skillmark benchmarks. |
| [`/turbo-mono-platform`](skills/turbo-mono-platform/SKILL.md) | Working on the monorepo stack | Full-stack TypeScript monorepo guidance (Turborepo, Next.js, Hono, tRPC, Drizzle, etc.). |
| [`/write`](skills/write/SKILL.md) | Writing, rewriting, polishing prose | Edit or write English/Vietnamese prose so it sounds natural, concise, and context-aware across docs, UI, reports, and marketing copy. |
| [`/watzup`](skills/watzup/SKILL.md) | End of work session, before PR | Review recent changes and wrap up current work session. Analyze commits, assess quality, identify risks. |

## Response convention

This repo follows a shared output convention inspired by Waza for active skills:

- the first response line should start inline with `🥷`
- the opening line should contain the verdict, recommendation, state, or next move
- outputs should be sharp, compressed, and decision-useful
- generic "default assistant" filler is considered a quality failure

The icon is the visible mode switch. The real standard is the writing: concrete, direct, and specific to the skill.

## Recommended workflow: `brainstorm` + `plan` + `cook` + friends

Three planning/execution skills as the front door, then hand off to the existing check + wrap-up skills.

![Recommended workflow for brainstorm + plan + cook with supporting skills](assets/spec-plan-workflow.svg)

Canonical pipeline:
```
brainstorm → plan → cook → check → git / watzup / handoff
```

### 1. Lock the problem with `brainstorm`
Use `brainstorm` when you have a raw idea, notes, markdown files, or a trade-off question. It runs in 4 modes (`explore`, `lock-from-idea`, `lock-from-files`, `refine`) and produces either a recommendation report or a locked `.kit/planning/SPEC.md`.

### 2. Derive execution with `plan`
Use `plan` only after the spec is locked. It turns `.kit/planning/SPEC.md` into `.kit/planning/ROADMAP.md` plus per-phase `-CONTEXT.md` and `-PLAN.md` files. If the spec is missing or too weak, `plan` fails fast and points back to `brainstorm`.

### 3. Run the kitchen with `cook`
Use `cook` to execute the plan. It checks for missing artifacts and routes back to `brainstorm` or `plan` if needed; otherwise it runs the active phase wave-by-wave, dispatches subagents for heavy tasks, verifies every task, and calls `/check` as the phase gate. It never auto-commits — handoffs are suggested, not executed.

### 4. Pull in support skills only when needed
- [`/check`](skills/check/SKILL.md) is invoked by `cook` per phase, or directly for ad-hoc gate/review
- [`/git`](skills/git/SKILL.md), [`/watzup`](skills/watzup/SKILL.md), and [`/handoff`](skills/handoff/SKILL.md) close or transfer a work session cleanly

### Mental model
- `brainstorm` = lock **WHAT**
- `plan` = lock **HOW**
- `cook` = run **the kitchen** (execute, verify, gate)
- `check` = catch risk and prove readiness (gate + analysis)
- `git` / `watzup` / `handoff` = wrap up with discipline

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

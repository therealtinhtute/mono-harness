# TINHTUTE Skills

A collection of personal Claude Code skills following the [skills.sh](https://skills.sh) ecosystem format.

## Structure

![Layer stack — 3 abstraction layers of the repository](assets/diagram-layer-stack.png)



## Install

```bash
npx skills add git@github.com:therealtinhtute/skills.git -a claude-code -g -y
```

## Skills

| Skill | When | What it does |
| :--- | :--- | :--- |
| [`/bash-tui`](skills/bash-tui/SKILL.md) | Building interactive terminal UIs | Build bash/shell TUI apps with menus, selectors, forms, progress bars, spinners, banners, and color output. |
| [`/brainstorm`](skills/brainstorm/SKILL.md) | Ideation, architecture decisions, multi-option choices | Explore options, evaluate trade-offs, and recommend the simplest viable path. |
| [`/git`](skills/git/SKILL.md) | Staging, committing, pushing, PRs, merges | Git operations with conventional commits. Auto-splits commits by type/scope. Security scans for secrets. |
| [`/handoff`](skills/handoff/SKILL.md) | Session end, context switches, milestones | Capture session state and write HANDOFF.md for seamless continuation. |
| [`/interview`](skills/interview/SKILL.md) | Validating plans before implementation | Interview about plans using AskUserQuestion. Explore technical decisions, UI/UX, concerns, tradeoffs. Write validated spec. |
| [`/plan`](skills/plan/SKILL.md) | After spec, before implementation | Turn a locked `.planning/SPEC.md` into a roadmap, per-phase context, and executable wave-based plans. |
| [`/prompt-leverage`](skills/prompt-leverage/SKILL.md) | Improving prompts, building frameworks | Strengthen raw user prompts into execution-ready instruction sets for AI agents. |
| [`/review`](skills/review/SKILL.md) | Before commit, PR, or merge | Pre-commit gate (tests, lint, build) + code review (security, architecture, maintainability) in one skill. |
| [`/skill-creator`](skills/skill-creator/SKILL.md) | Creating or updating Claude skills | Create or update Claude skills optimized for Skillmark benchmarks. |
| [`/spec`](skills/spec/SKILL.md) | Before roadmap or implementation planning | Turn an idea or files into a locked `.planning/SPEC.md` with scope, constraints, and acceptance criteria. |
| [`/turbo-mono-platform`](skills/turbo-mono-platform/SKILL.md) | Working on the monorepo stack | Full-stack TypeScript monorepo guidance (Turborepo, Next.js, Hono, tRPC, Drizzle, etc.). |
| [`/watzup`](skills/watzup/SKILL.md) | End of work session, before PR | Review recent changes and wrap up current work session. Analyze commits, assess quality, identify risks. |

## Response convention

This repo follows a shared output convention inspired by Waza for active skills:

- the first response line should start inline with `🥷`
- the opening line should contain the verdict, recommendation, state, or next move
- outputs should be sharp, compressed, and decision-useful
- generic "default assistant" filler is considered a quality failure

The icon is the visible mode switch. The real standard is the writing: concrete, direct, and specific to the skill.

## Recommended workflow: `spec` + `plan` + friends

Use the two new planning skills as the front door, then hand off to the existing execution / review / wrap-up skills.

![Recommended workflow for spec + plan with supporting skills](assets/spec-plan-workflow.svg)

### 1. Lock the problem with `spec`
Use `spec` when you have a raw idea, notes, or a feature request and want a clean `.planning/SPEC.md` before implementation.

### 2. Derive execution with `plan`
Use `plan` only after the spec is locked. It turns `.planning/SPEC.md` into `.planning/ROADMAP.md` plus per-phase `-CONTEXT.md` and `-PLAN.md` files.

### 3. Pull in support skills only when needed
- use [`/review`](skills/review/SKILL.md) after implementation for gate checks and code analysis
- use [`/git`](skills/git/SKILL.md), [`/watzup`](skills/watzup/SKILL.md), and [`/handoff`](skills/handoff/SKILL.md) to close or transfer a work session cleanly

### Mental model
- `spec` = lock **WHAT**
- `plan` = lock **HOW**
- `review` = catch risk and prove readiness (gate + analysis)
- `git` / `watzup` / `handoff` = wrap up with discipline

## Extras

These are **not** installable via `npx skills add`. Copy them from a local clone of this repo.

### English Coaching

Corrects your English in place, tags patterns so you learn the rule. Installs as a Claude Code rule.

```bash
git clone git@github.com:therealtinhtute/skills.git /tmp/skills \
  && mkdir -p ~/.claude/rules \
  && cp /tmp/skills/rules/english.md ~/.claude/rules/english.md
```

### Statusline

Minimal statusline: `✦ model  ▰▱ N%  ϟ tpm  ⌥ branch`. Model-first layout with battery-style progress bar.

```bash
git clone git@github.com:therealtinhtute/skills.git /tmp/skills \
  && bash /tmp/skills/scripts/setup-statusline.sh
```

Then add to `~/.claude/settings.json`:

```json
"statusLine": { "type": "command", "command": "bash ~/.claude/statusline.sh" }
```

### Settings Config

Example `settings.json` with env vars, model defaults, hooks, and attribution:

```bash
git clone git@github.com:therealtinhtute/skills.git /tmp/skills \
  && cp /tmp/skills/setup/settings.json ~/.claude/settings.json
# Edit ANTHROPIC_AUTH_TOKEN and review hooks before using
```

## Local Development

This repo is the stable release. The incubator workspace lives at `/Users/tinhtute/Lab/orkit-tui/kit/skills/`.

To iterate quickly on a skill before publishing:

```bash
claude-code add-dir /Users/tinhtute/Lab/orkit-tui/kit/skills/my-skill
```

To sync changes from the incubator to this repo:

```bash
bash scripts/sync-from-kit.sh
```

## License

MIT

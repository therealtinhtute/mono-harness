# TINHTUTE Skills

A collection of personal Claude Code skills following the [skills.sh](https://skills.sh) ecosystem format.

## Structure

![Layer stack — 3 abstraction layers of the repository](assets/diagram-layer-stack.png)

```mermaid
flowchart TB
    L2[\"Skills<br/>13 agent skills · SKILL.md\"]
    L1[\"Config & Rules<br/>settings.json · english.md · CLAUDE.md\"]
    L0[\"Install & Distribution<br/>npx skills add · curl | bash · sync-from-kit.sh\"]
    L0 --> L1 --> L2
```

## Install

```bash
npx skills add git@github.com:therealtinhtute/skills.git -a claude-code -g -y
```

## Skills

| Skill | When | What it does |
| :--- | :--- | :--- |
| [`/bash-tui`](skills/bash-tui/SKILL.md) | Building interactive terminal UIs | Build bash/shell TUI apps with menus, selectors, forms, progress bars, spinners, banners, and color output. |
| [`/brainstorm`](skills/brainstorm/SKILL.md) | Ideation, architecture decisions, technical debates | Brainstorm solutions with trade-off analysis and brutal honesty. |
| [`/git`](skills/git/SKILL.md) | Staging, committing, pushing, PRs, merges | Git operations with conventional commits. Auto-splits commits by type/scope. Security scans for secrets. |
| [`/handoff-manager`](skills/handoff-manager/SKILL.md) | Session end, context switches, milestones | Capture session state and write HANDOFF.md for seamless continuation. |
| [`/interviewer`](skills/interviewer/SKILL.md) | Vague requirements, unclear specs | Extract complete, unambiguous requirements through relentless questioning before implementation. |
| [`/investigator`](skills/investigator/SKILL.md) | Exploring unfamiliar code, finding definitions | Rapidly locate relevant files and understand code structure before planning or implementing. |
| [`/media-processing`](skills/media-processing/SKILL.md) | Video, audio, image conversion/optimization | Process media using FFmpeg and ImageMagick. Convert, encode, resize, crop, optimize, composite. |
| [`/prompt-leverage`](skills/prompt-leverage/SKILL.md) | Improving prompts, building frameworks | Strengthen raw user prompts into execution-ready instruction sets for AI agents. |
| [`/reviewer`](skills/reviewer/SKILL.md) | Before commit or merge | Expert code review covering security, performance, architecture, and maintainability. |
| [`/skill-creator`](skills/skill-creator/SKILL.md) | Creating or updating Claude skills | Create or update Claude skills optimized for Skillmark benchmarks. |
| [`/strategist`](skills/strategist/SKILL.md) | Multiple valid approaches, planning | Evaluate options, expose trade-offs, and recommend the simplest viable path. |
| [`/turbo-mono-platform`](skills/turbo-mono-platform/SKILL.md) | Working on the monorepo stack | Full-stack TypeScript monorepo guidance (Turborepo, Next.js, Hono, tRPC, Drizzle, etc.). |
| [`/verifier`](skills/verifier/SKILL.md) | Before commit, merge, or ship | Run quality gates and confirm work is ready with evidence, not assumptions. |

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

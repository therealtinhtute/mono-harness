# TINHTUTE Skills

A collection of personal Claude Code skills following the [skills.sh](https://skills.sh) ecosystem format.

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

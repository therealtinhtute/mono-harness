# mono-harness

A personal collection of [skills.sh](https://skills.sh)-compatible skills for Claude Code and other AI agents, backed by a durable local workflow harness.

## What this repository contains

| Area | Contents |
| --- | --- |
| skills/workflow/ | Eight skills for discovery, planning, execution, review, handoff, and Git workflow. |
| skills/craft/ | Four skills for writing, GitHub research, skill authoring, and prompt improvement. |
| skills/shipping/ | Two skills for CLI design and full-stack TypeScript monorepos. |
| cli/ | The Go zharness binary powering the workflow chain (ledger reconciliation until v0.15 slims it to installer/updater). |
| rules/ | Global Claude Code rules installed by setup/install.sh. |
| setup/ | Bootstrap files, hooks, settings, and installation logic. |
| scripts/ | CLI installation, validation, statusline, dashboard, and documentation checks. |
| assets/ | README visuals and workflow diagrams. |

Each skill has a required SKILL.md and may include references/ or scripts/ for progressive detail.

## Installation

### Full Claude Code bootstrap

Prerequisites:

- Node.js 18+
- jq for the statusline
- SSH access to git@github.com:therealtinhtute/mono-harness.git

~~~bash
git clone git@github.com:therealtinhtute/mono-harness.git ~/mono-harness
cd ~/mono-harness
bash setup/install.sh
~~~

The installer:

- installs the global CLAUDE.md, rules, hooks, settings template, and statusline;
- backs up files before replacing them;
- preserves an existing ~/.claude/settings.json instead of overwriting it;
- installs all repository skills globally through npx skills add.

After installation, review ~/.claude/settings.json and configure ANTHROPIC_AUTH_TOKEN as required by your environment.

### Skills only

List the published skills without installing them:

~~~bash
npx skills add git@github.com:therealtinhtute/mono-harness.git --list
~~~

Install all skills for Claude Code:

~~~bash
npx skills add git@github.com:therealtinhtute/mono-harness.git -a claude-code -g -y
~~~

### Install zharness

Install the latest release of the workflow CLI:

~~~bash
bash scripts/install-zharness.sh
zharness --version
~~~

The release installer requires an authenticated gh CLI and tar. It installs the binary to ~/.local/bin/zharness for Linux or macOS on amd64 or arm64.

## Skill catalog

### Workflow

| Skill | Purpose |
| --- | --- |
| [brainstorm](skills/workflow/brainstorm/SKILL.md) | Explore options and lock requirements into an active plan. |
| [to-plan](skills/workflow/to-plan/SKILL.md) | Turn a locked plan into phases, waves, tasks, and checks. |
| [work](skills/workflow/work/SKILL.md) | Execute approved work and verify each task. |
| [check](skills/workflow/check/SKILL.md) | Run gates and review changes before shipping. |
| [handoff](skills/workflow/handoff/SKILL.md) | Persist current state and the exact next action. |
| [watzup](skills/workflow/watzup/SKILL.md) | Reconstruct branch, lifecycle, and handoff state. |
| [git](skills/workflow/git/SKILL.md) | Stage, commit, push, and prepare pull requests. |
| [interview](skills/workflow/interview/SKILL.md) | Resolve ambiguous intent through structured questions. |

### Craft

| Skill | Purpose |
| --- | --- |
| [write](skills/craft/write/SKILL.md) | Write and edit concise English or Vietnamese prose. |
| [librarian](skills/craft/librarian/SKILL.md) | Research code and evidence in external GitHub repositories. |
| [create-skill](skills/craft/create-skill/SKILL.md) | Create or improve Claude skills with references and validation. |
| [prompt-leverage](skills/craft/prompt-leverage/SKILL.md) | Turn rough prompts into execution-ready instructions. |

### Shipping

| Skill | Purpose |
| --- | --- |
| [create-cli](skills/shipping/create-cli/SKILL.md) | Design CLI commands, flags, I/O, errors, and delivery strategy. |
| [turbo-mono-platform](skills/shipping/turbo-mono-platform/SKILL.md) | Work across the Turborepo, Next.js, Hono, tRPC, Drizzle, and Postgres stack. |

## Workflow

The workflow skills use zharness for replayable enrichment wherever the binary is present; the durable record itself is the committed plan markdown under `docs/plans/`. The normal durable path is:

~~~text
/brainstorm → /to-plan → /work → /check → /git → /handoff
~~~

Legacy release setups may initialize per-machine state first — it is not a prerequisite for markdown-first use:

~~~bash
cd /path/to/project
zharness init --json
~~~

zharness init creates the root harness.db, replays existing changesets, and scaffolds the managed workflow docs under docs/.

Use the smallest workflow mode that matches the work:

~~~text
/brainstorm explore              # response-only exploration
/work simple <concrete task>     # bounded change, no lifecycle rows
/check bounded                   # response-only gate for bounded work
/watzup                          # read-only resume and branch recap
~~~

Every workflow stage attempts a read-only zharness preflight check when the binary exists; without it, each stage degrades to its markdown-first playbook. Reduced modes can continue without any of them.

## zharness state model

| Path | Role | Lifecycle |
| --- | --- | --- |
| docs/plans/active/ | Durable initiative plans containing requirements, phases, progress, decisions, validation, and current state. | Project-local workflow data — the record. |
| cli/docs/embedded/ | Canonical WORKFLOW.md, playbooks, AGENTS block shipped inside the CLI. | Tracked source. |
| docs/ | Managed root-doc projection; the installer refreshes it, the updater merges local edits. | Generated from the managed set; do not hand-edit. |
| .git/hooks/pre-commit | The two fail-closed guards (proof re-execution, independent judge). | Installed by scripts/install-git-hooks.sh. |

Committed markdown is the whole system of record; a legacy per-machine index is only a recovery cache and is never created again.

## CLI quick reference

`zharness --help` documents the installed surface. Since v0.15 the lifecycle lives in markdown plus repo scripts (`scripts/record-check.sh`, `scripts/install-git-hooks.sh`); the binary's install / update / uninstall verbs arrive with phase p3-installer.

## Development

### Iterate on a skill locally

~~~bash
claude-code add-dir /path/to/mono-harness
~~~

Edit the skill directly under skills/. Keep SKILL.md focused on routing and execution; move deep detail into references/ and reusable logic into scripts/.

### Validate a skill

~~~bash
bash scripts/validate-skill.sh skills/workflow/work/SKILL.md
~~~

Replace the path with the skill being changed.

### Build and test the CLI

The CLI targets Go 1.25 and builds with CGO disabled:

~~~bash
cd cli
CGO_ENABLED=0 go build ./...
go vet ./...
go test ./...
~~~

The GitHub Actions workflow runs the same build, vet, and test gates for changes under cli/.

### Check documentation references

~~~bash
bash scripts/verify-doc-links.sh
~~~

This checks repo-relative documentation claims against the files present in the checkout.

## Repository map

~~~text
.
├── skills/
│   ├── workflow/       # lifecycle and Git workflow skills
│   ├── craft/          # writing, research, skill, and prompt skills
│   └── shipping/       # CLI and full-stack engineering skills
├── cli/
│   ├── cmd/zharness/   # CLI entrypoint
│   ├── internal/       # interfaces, application, domain, infrastructure
│   └── docs/embedded/  # canonical managed docs and templates
├── rules/              # globally installed Claude Code rules
├── setup/              # bootstrap config, hooks, and installer
├── scripts/            # validation, installation, and local tooling
└── assets/             # README and workflow visuals
~~~

## Contributing and security

Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Report vulnerabilities privately through [SECURITY.md](SECURITY.md). Never commit API keys, access tokens, private settings, or local harness state.

## License

MIT

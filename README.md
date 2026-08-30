# mono-harness

Turn a git repository into a legible, agent-ready workspace.

`zharness` installs a small repository protocol and a safe updater. The
repository remains the system of record: product documents, decisions, plans,
code, tests, CI, and runtime evidence define the work.

It is not a task database, story tracker, agent orchestrator, or application
runtime. The binary scaffolds docs; it does not run the lifecycle.

Start with [`AGENTS.md`](AGENTS.md), then [`docs/WORKFLOW.md`](docs/WORKFLOW.md).

## Give this to a coding agent

Copy the block into the consumer repository chat when you want the agent to
install, refresh, or work with zharness.

```text
Work in this git repository. Read https://github.com/therealtinhtute/mono-harness/blob/master/README.md and, once they exist, this repo's AGENTS.md and docs/WORKFLOW.md. If zharness is not on PATH, install the binary from that repository with scripts/install-zharness.sh. If this repo has no docs/WORKFLOW.md, run zharness install at the root. If the protocol is already installed and we asked to refresh, run zharness update. Then do the work those docs describe: read-only stays read-only, a small change needs no plan, longer work uses one docs/plans/active file. zharness only installs, updates, and uninstalls the doc set — it does not run the lifecycle.
```

## What it solves

Coding agents often fail for ordinary engineering reasons:

- important intent exists only in chat;
- the repository does not identify authoritative documents;
- small changes acquire unnecessary process;
- long changes lose decisions and recovery context;
- completion is claimed without behavior-level proof; and
- an agent invents product policy when the request leaves a material choice
  open.

zharness provides a compact entrypoint, a navigable repository map, durable
plans only when work needs them, and playbooks that stay reduced for read-only
and bounded work.

## Default workflow

```text
read-only request
  -> inspect the smallest authoritative surface
  -> answer with evidence

bounded change
  -> inspect authority and affected behavior
  -> implement the smallest coherent change
  -> run relevant proof

multi-session or coordinated change
  -> create docs/plans/active/<plan>.md
  -> keep decisions, progress, recovery, and validation current
  -> move the validated plan to docs/plans/completed/

material product ambiguity
  -> stop before mutation
  -> present the concrete choice and consequences
```

A typo does not need a plan. A migration spanning sessions does.

## What gets installed

The managed set is:

- a compact `AGENTS.md` entrypoint (marked `ZHARNESS` block only);
- `docs/WORKFLOW.md` and the six stage playbooks;
- a `docs/PROJECT.md` identity scaffold;
- `.zharness/base/` for three-way updates.

It does not install application architecture, product policy, skills, git
hooks, credentials, a database, schemas, orchestration, or background
processes.

## Install

From a target repository, with `zharness` on PATH:

```bash
zharness install
```

Get the binary once per machine (`gh` + `tar`; Linux or macOS, amd64 or arm64):

```bash
bash scripts/install-zharness.sh
zharness --version
```

`install` is idempotent. It records upstream hashes, prints a read-only
brownfield report, and exits 0. Use `--root <dir>` when cwd is not the
consumer repo.

## Maintain an installation

```bash
zharness update
zharness update --continue
zharness update --abort
zharness uninstall
```

The updater stores the exact upstream base under `.zharness/base/` and
three-way-merges. If local and upstream edits overlap, it stops with conflict
markers. After a human resolves them, `--continue`. `--abort` restores the
pre-update bytes. Uninstall removes managed files only; consumer-owned bytes
are never deleted.

## Optional skills

Skills are not part of `zharness install`. They live in this source repository:

```bash
npx skills add git@github.com:therealtinhtute/mono-harness.git -a claude-code -g -y
```

No skill runs during installation.

## v0.15

The former lifecycle CLI and SQLite ended in v0.15. Pin `v0.14.x` to keep
that binary. Existing `harness.db` files are consumer-owned; nothing here
deletes them.

See [`cli/docs/CONTRACT.md`](cli/docs/CONTRACT.md) and
[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Development

```bash
cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...
bash scripts/verify-doc-links.sh
bash scripts/test-guards.sh
```

This repository is the source of the binary, the embedded playbooks, and the
skills. Edit playbooks in `cli/docs/embedded/playbooks/`, then copy to
`docs/playbooks/`. Machine-wide Claude Code bootstrap is `setup/install.sh`;
that is not how a consumer app repo receives zharness.

## Contributing and security

Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.
Report vulnerabilities privately through [SECURITY.md](SECURITY.md).

## License

MIT

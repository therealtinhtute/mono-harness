---
title: create-cli — example outputs
description: Sample spec and roadmap outputs for greenfield and retrofit CLIs
status: active
tags: [create-cli, examples]
---

### Example 1 — Greenfield: `orkit` deployment CLI (Go)

**Trigger:** "I want a CLI to deploy services to our k8s clusters"

**Clarify output:**
- Command name: `orkit`
- One-liner: deploy services to Kubernetes clusters
- User type: ops engineers and CI pipelines
- Language: Go (single binary, no runtime dependency)

**Spec excerpt (saved to `.kit/planning/cli-orkit-spec.md`):**
```
orkit deploy <service> [--env staging|prod] [--dry-run] [--json]
orkit status <service> [--watch] [--json]
orkit rollback <service> [--to <version>]
```

**Roadmap tasks:**
1. Scaffold with cobra + viper, wire `--help`/`--version`
2. Implement `deploy` command — kubeconfig lookup, context selection
3. Implement `status` — poll deployment rollout with `--watch`
4. Add `--json` output mode (detect TTY)
5. Config: `~/.config/orkit/config.yaml` with XDG
6. Shell completions (bash, zsh, fish)
7. Goreleaser + GitHub Actions release pipeline

---

### Example 2 — Retrofit: `deploy.sh` → `deploy` CLI (Bash)

**Trigger:** "I have this deploy.sh that everyone uses, make it proper"

**As-is interface extracted:**
- Takes positional: `./deploy.sh <env> <service>`
- Reads `DEPLOY_TOKEN` env var
- Outputs to stdout, exits 0 even on failure
- No help text, no dry-run

**Gap analysis (partial):**

| Convention | Current | Target | Breaking? |
|---|---|---|---|
| `--help` | missing | add | no |
| exit codes | always 0 | standard set | yes |
| `--dry-run` | missing | add | no |
| `--json` | missing | add | no |

**Migration roadmap:**
1. Add `--help` and argument validation (non-breaking)
2. Add `--dry-run` flag
3. Fix exit codes with deprecation warning in v1, hard break in v2

---

### Example 3 — Greenfield: `snip` clipboard manager (Node.js)

**Trigger:** "A CLI tool to save and retrieve text snippets, publish to npm"

**Clarify output:**
- Command name: `snip`
- Distribution: npm global install (`npm i -g snip-cli`)
- Config: `~/.config/snip/snippets.json`
- Users: developers

**Spec excerpt:**
```
snip add <name> [--from-stdin] [--tag <tag>]
snip get <name> [--copy]
snip list [--tag <tag>] [--json]
snip rm <name>
```

**Framework choice:** Node.js + `commander` — npm distribution is the deciding factor; single binary would need `pkg` bundling which adds complexity for this use case.

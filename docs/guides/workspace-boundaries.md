# Workspace boundaries and nested instruction precedence

What a "repository", a "workspace", and the zharness registry mean in this harness, and how coding-agent hosts resolve root versus nested `AGENTS.md` / `CLAUDE.md`. Host behavior below was checked against the hosts' official documentation on 2026-10-24; hosts change, so re-run the [manual test](#manual-test-procedure) after upgrading one.

## Terms

| Term | Meaning here |
|---|---|
| Repository | One Git working tree with its own `AGENTS.md`, `.harness/WORKFLOW.md`, `.harness/PROJECT.md`, and `.harness/plans/`. The harness workflow (stages, plans, guards) runs inside exactly one repository. |
| Workspace | Not a harness concept. The word is the host's or the user's (an IDE window, a monorepo, a folder holding several clones). The harness defines no workspace file, root, or state. |
| Registry | A machine-wide list of repository roots that zharness has installed into. An inventory, nothing more. |

## Single repository versus registry inventory

Single-repository workflow: every `zharness install`, `update`, `update --check`, and `uninstall` acts on one root (`--root`, default the Git toplevel of the current directory). Plans, `PROJECT.md`, and the managed `AGENTS.md` block live in that repository.

Registry inventory: `install` and `update` append the repository's canonical absolute path to `$XDG_CONFIG_HOME/zharness/repos` (fallback `~/.config/zharness/repos`), one root per line; `uninstall` removes it (`cli/src/installer/registry.rs`). Registration is advisory: if the file cannot be written, the verb prints a warning and still succeeds. The only consumer is `zharness update --check --all` (`cli/src/installer/check.rs`), which loops over the listed roots and runs the same read-only drift check on each:

- `--all` requires `--check`; `update --all` alone fails with `--all requires --check`.
- Each root prints `current`, `drift` (with the stale files), `missing` (path no longer a directory), or `error`.
- Exit 1 when any root drifted or failed; a `missing` root alone does not fail.
- An empty registry prints `no registered repositories`.

What the registry is not:

- Not task orchestration: no verb starts, schedules, or coordinates work in more than one repository.
- Not shared memory or shared context: plans, `PROJECT.md`, and session notes are per repository and never read across repositories. The registry stores paths only.
- Not a workspace definition: membership means "zharness was installed here on this machine", not "these repositories belong together".

A cross-repository orchestrator is out of scope until a separate, evidenced requirement justifies it.

## Root versus nested instruction files by host

The harness does not implement precedence. The host decides which instruction files enter the model's context and in what order; a rule that must hold regardless belongs in a guard (hook, CI), not in prose.

| Host | Discovery | Order / precedence | Source |
|---|---|---|---|
| Claude Code | Loads `CLAUDE.md` and `CLAUDE.local.md` from the working directory and every directory above it at launch. Nested files under the working directory load on demand, when Claude reads, writes, or edits a file in that subdirectory. Reads `AGENTS.md` only when no `CLAUDE.md` / `.claude/CLAUDE.md` / `CLAUDE.local.md` exists in the working directory or above (default `Project instructions` = `claude-md-or-agents-md`); a nested `AGENTS.md` loads on demand only if that subdirectory has none of the three `CLAUDE.md` files. A `CLAUDE.md` containing `@AGENTS.md` pulls it in. Ignores `AGENTS.override.md` and `AGENTS.local.md`. | Files are concatenated, not replaced: filesystem root down to the working directory, so closer files are read last. A conflict between two files is not resolved by the host; the docs say Claude "may pick one arbitrarily". | <https://code.claude.com/docs/en/memory> (sections "How CLAUDE.md files load", "AGENTS.md") |
| Codex | At start, reads a global `~/.codex/AGENTS.override.md` or `AGENTS.md`, then from the project root (usually the Git root) down to the working directory checks `AGENTS.override.md`, then `AGENTS.md`, then configured fallback names, at most one file per directory. Nested directories below the working directory are not walked. | Concatenated root-down with blank lines; later (closer) files override earlier guidance because they appear later. Stops at `project_doc_max_bytes` (32 KiB default). | <https://learn.chatgpt.com/docs/agent-configuration/agents-md> |
| Other hosts (Cursor, Gemini CLI, and similar) | Host-defined. Not verified for this guide. | The shared `AGENTS.md` convention says the nearest file in the tree takes precedence (<https://agents.md/>), but that is a convention, not a tested guarantee for any host. | host documentation |

Consequences for this repository's layout:

- Run the agent from the repository root. Both verified hosts then see the root `AGENTS.md`; nested files are optional refinements.
- In Claude Code, a repository that has any `CLAUDE.md` above the working directory will not have its `AGENTS.md` read unless that `CLAUDE.md` imports it with `@AGENTS.md`. Codex ignores `CLAUDE.md`.
- Launching from a subdirectory of a repository, or from a parent folder that holds several repositories, changes which files load. In a parent folder Claude Code loads the parent's files and each child's files only on demand; Codex finds no project root and checks the current directory only. Do not rely on a child repository's `AGENTS.md` being active there.
- A nested file may add rules; do not use it to cancel a root rule. Hosts concatenate, they do not merge, so contradictions are left to the model.

## Manual test procedure

Covers the behavior documented above using a throwaway directory and canary strings. Needs `claude` and/or `codex` on `PATH`; skip a host you do not use. Replace the final `ask` command with the host's non-interactive form if flags differ in your version.

1. Build the fixture:

   ```bash
   T=$(mktemp -d) && mkdir -p "$T/repo/svc/pay" "$T/repo/svc/search"
   echo 'Canary ROOT-AGENTS: always mention ROOT-AGENTS.'   > "$T/repo/AGENTS.md"
   echo 'Canary NEST-AGENTS: always mention NEST-AGENTS.'   > "$T/repo/svc/pay/AGENTS.md"
   echo 'Canary NEST-OVERRIDE: always mention NEST-OVERRIDE.' > "$T/repo/svc/pay/AGENTS.override.md"
   echo 'Canary SIBLING: always mention SIBLING.'           > "$T/repo/svc/search/AGENTS.md"
   echo 'pay' > "$T/repo/svc/pay/file.txt"
   ```

2. Codex, from the subdirectory: expect the root canary, then the override canary; `NEST-AGENTS` ignored (override wins in the same directory); `SIBLING` absent.

   ```bash
   git -C "$T/repo" init -q
   (cd "$T/repo/svc/pay" && codex exec "List every canary string from your instructions, in the order loaded.")
   ```

   Then from `$T/repo`: expect only `ROOT-AGENTS`.

3. Claude Code, from the root, with no `CLAUDE.md` anywhere above: expect `ROOT-AGENTS` at launch and no nested canary until a nested file is read.

   ```bash
   (cd "$T/repo" && claude -p "List every canary string in your instructions.")
   (cd "$T/repo" && claude -p "Read svc/pay/file.txt, then list every canary string in your instructions.")
   ```

   Expected: first run `ROOT-AGENTS` only; second run adds `NEST-AGENTS` (override files are not read by Claude Code, so `NEST-OVERRIDE` stays absent). Inside an interactive session, `/context` lists the loaded memory files.

4. Claude Code with a `CLAUDE.md` present: `echo 'Canary CLAUDE: mention CLAUDE.' > "$T/repo/CLAUDE.md"`, repeat step 3. Expected: `CLAUDE` appears and `ROOT-AGENTS` does not. Then replace the file with `@AGENTS.md` plus the canary line and expect both.

5. Registry boundary, no host needed (uses a scratch registry, so your real one is untouched):

   ```bash
   export XDG_CONFIG_HOME="$T/cfg"; mkdir -p "$T/a" "$T/b"
   zharness install --root "$T/a" >/dev/null && zharness install --root "$T/b" >/dev/null
   cat "$XDG_CONFIG_HOME/zharness/repos"        # two absolute roots, nothing else
   zharness update --check --all; echo "rc=$?"  # current for both, rc=0
   zharness update --all; echo "rc=$?"          # "--all requires --check", rc=1
   echo x >> "$T/b/.harness/WORKFLOW.md"
   zharness update --check --all; echo "rc=$?"  # drift for b only, rc=1
   mv "$T/b" "$T/b2"
   zharness update --check --all; echo "rc=$?"  # "missing: .../b", rc=0
   ```

   Expected: only drift is reported; no plan, memory, or task state from either repository appears in the other, and the registry file holds paths only.

A host result that differs from the table means the host changed: update the table and its date, do not adapt the repository to the old behavior.

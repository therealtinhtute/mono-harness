# Harness improvement — shell-safety-reinstall

## Baseline
- representative job: "reinstall CLI, rules, and skills from scratch and remove the old ones" on a machine whose Bash tool shell is zsh.
- accepted outcome: dropped skills removed, current skills and rules installed, nothing else deleted.
- concrete failure and evidence: session `520999b7-b8d5-46d6-946f-bcb837880c1e`. `mapfile` failed in zsh, the target list was empty, and `trash ~/.claude/skills/$s` trashed both whole skills directories (restored from Trash). An earlier `for s in $S` loop did not word-split in zsh, so 3 dropped skills survived the first cleanup. The repo documented no reinstall/prune procedure, so rules were copied and dropped skills were found by hand.
- human steering required: the user asked "ban có xoá skill cũ luôn chưa" to confirm the cleanup.
- worker / revision / tools: Claude Code, Opus 5.5, Bash tool (zsh).

## Gap owner
- earliest gap: context (the shell was not bash) and domain (no reinstall procedure)
- assigned owner: repository-harness (`rules/execution-discipline.md`, `README.md`)

## Intervention
```
If a shell-safety bullet is added at rules/execution-discipline.md §1 and a
"reinstall from scratch" block is added at README.md "Optional skills", then a
fresh agent will wrap bash-isms in `bash <<'EOF'` with set -euo pipefail, abort
empty delete loops, and follow the documented rules+skills reinstall on the
representative job, because both files are what the agent already reads.
Evidence that would weaken this: a fresh zsh session still runs a bare bash-ism
loop, or still hand-derives the reinstall steps.
Maintenance owner and removal condition: repo owner; remove the bullet if the
Bash tool guarantees bash, and the README block if a CLI verb takes over.
```

## Native proof
- command: `bash scripts/verify-doc-links.sh && diff -q rules/execution-discipline.md ~/.claude/rules/execution-discipline.md`
- result: `doc links OK (0 findings)`; installed rule identical to `rules/execution-discipline.md`; the bash-heredoc pattern with an empty target list aborts before any delete (`empty-guard-aborts`).

## Fresh rerun
- different session: no
- Decision: pending fresh rerun
- comparison:

Do not claim the harness improved while Decision is `pending fresh rerun`.

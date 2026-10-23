# Playbook: ship

Land and release what is verified. Merge, tag, publish and deploy run only on the owner's explicit request; each landing is confirmed before the next starts.

1. `watzup` — list commits since the last published tag (the tag is the baseline, not a local version file); verdict release / skip, bump patch (fixes) / minor (feat) / major (breaking), and the key risk in one sentence.
2. `security-review` (Claude Code built-in) — the diff since that tag; a high-severity finding blocks the release until fixed or explicitly waived by the owner.
3. `check full` — per PR, bottom of the stack first: record its verdict with the head SHA it covers. Also check every shipped surface: CLI `--help` and version output, install path, migrations, config defaults.
4. Land only the contiguous verified run from the bottom: stop at the first PR without a pass; a verified PR above it waits.
5. Re-check each verdict still describes the code: `git patch-id` at the verdict SHA equals the current head's, otherwise re-run step 3 for that PR.
6. `git` — merge bottom-up, one PR at a time, guarded by the verified head: `gh pr merge <n> --squash --match-head-commit <sha>`; confirm the base right before. Touch no branch or base while the run lands.
7. `write` — CHANGELOG entry and release notes from the commit list; breaking changes first, each with its migration step.
8. `git` — bump the version, commit, tag and push; never force-push a published tag.
9. Verify the delivery: the release CI run passes, the artifact or deploy is reachable, and its version output matches the tag.
10. Name the rollback before closing: the previous tag, and the exact revert or redeploy command.

Reply: the verdict and version, each PR's verdict SHA and how it landed, the security result, the release URL or deploy evidence, and the rollback command.

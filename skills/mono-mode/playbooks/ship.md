# Playbook: ship

Release what is already merged. Tag, publish and deploy run only on the owner's explicit request.

1. `watzup` — list commits since the last published tag (the tag is the baseline, not a local version file); verdict release / skip, bump patch (fixes) / minor (feat) / major (breaking), and the key risk in one sentence.
2. `security-review` — the diff since that tag; a high-severity finding blocks the release until fixed or explicitly waived by the owner.
3. `check full` — the gate commands on the release commit, plus every shipped surface: CLI `--help` and version output, install path, migrations, config defaults.
4. `write` — CHANGELOG entry and release notes from the commit list; breaking changes first, each with its migration step.
5. `git` — bump the version, commit, tag and push only when the owner asks; never force-push a published tag.
6. Verify the delivery: the release CI run passes, the artifact or deploy is reachable, and its version output matches the tag.
7. Name the rollback before closing: the previous tag, and the exact revert or redeploy command.

Reply: the verdict and version, the security result, the release URL or deploy evidence, and the rollback command.

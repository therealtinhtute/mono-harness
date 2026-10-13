---
status: active
lane: high-risk
---
# Playbooks move into skill references

## Goal
- outcome: each stage's procedure lives in `skills/workflow/<stage>/references/`, routed from its SKILL.md. `zharness` stops shipping playbooks: it installs `docs/WORKFLOW.md` (routing to skills), the AGENTS block and the PROJECT.md scaffold. `update` retires previously installed playbooks whose bytes match the recorded upstream hash and keeps hand-edited copies with a notice.
- success_signal: on a scratch consumer that holds old installed playbooks (one edited), `zharness update` deletes the untouched ones, keeps the edited one with a notice, and writes no playbook; in this repo no `docs/playbooks/` or `cli/docs/embedded/playbooks/` remains, every stage skill resolves its references, and all 4 gate commands pass.
- actors: owner; contributor agents; consumers who run `zharness install`/`update`; agents without skills support (they lose the procedure, which the owner accepted).
- authority: owner answers 2026-10-12 and 2026-10-13: playbooks into skill references; zharness stops shipping playbooks; keep WORKFLOW + AGENTS block + PROJECT; delete orphans only when the hash matches; lock the plan now, another agent implements.
- requirements:
  - R1: ADR 0015 records the decision and amends ADR 0007 (playbooks leave the managed set; WORKFLOW.md stays fresh-overwrite) and ties orphan removal to ADR 0008's recorded ownership. | acceptance: `test -f docs/decisions/0015-playbooks-into-skill-references.md` and ADR 0007's Status links it | source: owner
  - R2: the 12 playbooks move to `skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup}/references/` (brainstorm-grill and brainstorm-lock under brainstorm; work-full under work; check-durable, check-validation and check-review under check). Every `docs/playbooks/<x>.md` reference becomes a skill-relative `references/<x>.md` path, and each stage SKILL.md routes to its references. | acceptance: `test ! -e docs/playbooks && test ! -e cli/docs/embedded/playbooks`; `bash scripts/validate-skill.sh` passes on the 6 stage skills; `bash scripts/verify-doc-links.sh` 0 findings | source: owner
  - R3: the contract tests in `cli/src/embedded.rs` keep every required phrase but read the skill reference files (`CARGO_MANIFEST_DIR/../skills/...`); `playbook_count_is_twelve` and the playbook half of `cli/tests/projection_parity.rs` go away with the projection. | acceptance: `cd cli && cargo test` passes with the same phrase count as before (git diff of `required:` arrays shows paths changed, no phrase removed) | source: owner
  - R4: `zharness install` and `update` no longer write playbooks. `update` deletes `docs/playbooks/<name>.md` only when its sha equals the recorded base hash, keeps a differing file with a notice line, removes the empty directory, and drops retired entries from `.zharness/base`. `uninstall` handles a repo with or without legacy playbooks. | acceptance: new installer tests (red first) cover: untouched legacy playbook deleted; edited one kept + notice; fresh install writes no `docs/playbooks/`; uninstall on both shapes | source: owner
  - R5: `docs/WORKFLOW.md` routes each stage to its skill, not to a playbook path; the AGENTS block and PROJECT.md behave as today. | acceptance: `grep -c "docs/playbooks" cli/docs/embedded/WORKFLOW.md` prints 0; the installer tests for the AGENTS block and PROJECT.md still pass | source: owner
- non-goals:
  - NG1: no change to the plan format, guard core (ZGUARD-CORE), lanes, or ADR 0014's flow.
  - NG2: no fallback copy of the procedure for agents without skills (owner accepted the loss).
  - NG3: `site/` is hand-authored and may drift; it is not updated here.

## Phases and Verification
- approach: decide first (ADR), then move the content inside this repo with the contract tests re-pointed red-first, then change the installer red-first, then dry-run a consumer upgrade. Each phase leaves the repo green. Rejected: deleting playbooks from the installer first, because consumers would lose the procedure before the skills carry it.
- risks: a consumer's hand-edited playbook is lost → delete only on an exact recorded-hash match, else keep and notice; skill-relative paths break when a skill is installed elsewhere → references are addressed relative to the SKILL.md (`references/<x>.md`), never by repo path; contract phrases silently drop during the move → R3's acceptance compares the `required:` arrays.
- seams: `cli/src/embedded.rs` contract tests; installer tests in `cli/src/installer/{mod,update,uninstall}.rs`; `scripts/validate-skill.sh`; `scripts/verify-doc-links.sh`; the pre-commit hook.
- constraints:
  - Keep every existing required contract phrase; only its file path may change.
  - Never delete a consumer file whose sha does not equal its recorded base hash.
  - The binary stays within the size gate (`AGENTS.md` gate commands).
- recovery: each phase is its own commit range; reverting p3 restores playbook shipping, reverting p2 restores `docs/playbooks/`.
- phase_slug: `p1-adr`
  status: planned
  - goal: R1 | depends_on: none
  - surfaces: `docs/decisions/0015-playbooks-into-skill-references.md`, `docs/decisions/0007-*.md` (Status line), `docs/decisions/README.md` | avoided: code, skills
  - escalate_when: the ADR would need to change guard-parsed plan fields
  - wave 1:
    - T1 write ADR 0015 (Status, Context, Decision, Consequences, Authority with path:line citations) and link it from ADR 0007 and the index — output: ADR — check: `bash scripts/verify-doc-links.sh` — stop_if: a citation cannot be resolved
  - phase check: `bash scripts/verify-doc-links.sh`
- phase_slug: `p2-skills`
  status: planned
  - goal: R2, R3, R5 | depends_on: p1-adr
  - surfaces: `skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup}/` (SKILL.md + new references/), `cli/src/embedded.rs` (contract paths), `cli/tests/projection_parity.rs`, `cli/docs/embedded/playbooks/` (removed), `docs/playbooks/` (removed), `cli/docs/embedded/WORKFLOW.md` + `docs/WORKFLOW.md`, `AGENTS.md` (Layout line on editing playbooks), `skills/workflow/README.md`, `README.md` | avoided: `cli/src/installer/`, guard scripts
  - escalate_when: a required contract phrase cannot survive the move verbatim
  - wave 1:
    - T2 re-point the contract tests to `../skills/workflow/<stage>/references/<x>.md` and run them red (files not there yet) — output: failing test naming the missing file — check: `cd cli && cargo test --lib embedded` — stop_if: fails for another reason
  - wave 2:
    - T3 `git mv` the 12 playbooks into the skill references; rewrite `docs/playbooks/<x>.md` mentions to `references/<x>.md`; each SKILL.md routes to its references; bump skill versions — output: green contract tests — check: `cd cli && cargo test --lib embedded` — stop_if: a phrase must change
    - T4 drop `cli/docs/embedded/playbooks/` from the embed (`embedded.rs` PLAYBOOKS, `playbook_names`, `playbook_count`, manifest) only as far as the build needs to stay green without touching installer behavior, and remove the playbook half of `projection_parity.rs` — output: build + tests green — check: `cd cli && cargo test` — stop_if: the installer needs changes (that is p3)
    - T5 WORKFLOW.md routes stages to skills; fix AGENTS.md, READMEs — check: `bash scripts/verify-doc-links.sh` and `bash scripts/validate-skill.sh` on each stage skill
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cd .. && bash scripts/verify-doc-links.sh`
- phase_slug: `p3-installer`
  status: planned
  - goal: R4, R5 | depends_on: p2-skills
  - surfaces: `cli/src/installer/{mod,update,uninstall,check}.rs`, `cli/src/embedded.rs` | avoided: skills, docs, guard scripts
  - escalate_when: retiring a playbook would require deleting a file without a recorded base hash
  - wave 1:
    - T6 write the R4 installer tests first (legacy untouched → deleted; edited → kept + notice; fresh install → no `docs/playbooks/`; uninstall both shapes) and run them red — check: `cd cli && cargo test --lib installer` — stop_if: fails for another reason
  - wave 2:
    - T7 remove playbooks from `all_targets()`; add the retire step in `update` keyed on `load_base` hashes; update `.zharness/base`; adjust uninstall — output: green — check: `cd cli && cargo test`
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test`
- phase_slug: `p4-dry-run`
  status: planned
  - goal: R4, success_signal | depends_on: p3-installer
  - surfaces: a scratch consumer repo outside this tree | avoided: everything else
  - escalate_when: the scratch upgrade deletes an edited file
  - wave 1:
    - T8 install the previous release's managed set into a scratch repo, edit one playbook, run the new `zharness update`, record what was deleted, kept and noticed — check: transcript plus `ls docs/playbooks` in the scratch repo
    - T9 run all 4 gate commands — check: all exit 0
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cd .. && bash scripts/test-guards.sh && bash scripts/verify-doc-links.sh`

## Log

## Validation

## Current State and Next Action
- active_phase: p1-adr
- lifecycle_status: planned
- blockers: none
- open_items: none
- exact_next_action: work full phase p1-adr

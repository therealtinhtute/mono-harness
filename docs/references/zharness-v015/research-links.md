# Research anchors: zharness v0.15

> **Archive — the external sources the v0.13 and v0.15 plans were argued from**, gathered
> 2026-08-26. The merged plan at `docs/plans/completed/zharness-v015-slim.md` names these in its
> Authority section by title; this file records where each one lives.
>
> **Verification status is stated per row.** A row marked *title only* means the source was
> named in the plan but its exact URL was not captured at authoring time — the domain and
> title are recorded so it can be found again, and the claim should be re-verified before it
> is leaned on. Nothing here is a guessed URL.

## Prior art the design copies from

| Source | Where | What was taken |
|---|---|---|
| `hoangnb24/repository-harness` — decision 0027, "EOL SQLite protocol" | private repo, no local clone on this machine · *title only* | The whole EOL playbook: pin the last release as the archive, one tree is one product, no `legacy/` directory, consumer bytes stay consumer-owned. This is the direct ancestor of R12 and NG5. |
| `hoangnb24/repository-harness` — its `AGENTS.md` | same repo · *title only* | The ~20-line, zero-CLI-required entry point; work-shape routing; "configurable defaults are not authority"; "no parallel control-plane state". Ancestor of the AGENTS.md block v2 in phase `p0-fail-open`. |
| `hoangnb24/repository-harness` — its updater | same repo · *title only* | Three-way merge on a base directory, conflicts stopping for human resolution via `--continue`/`--abort`, activation transactional. Ancestor of R9 and phase `p3-installer` wave 2. |
| `hoangnb24/repository-harness` — decision 0020 | same repo · *title only* | Knowledge boundaries stated explicitly, "no fabricated application truth", read-only-first explicit-only onboarding. Ancestor of R10's brownfield detection and NG4. |
| `github/spec-kit` | https://github.com/github/spec-kit | Prior art for a markdown-plus-scripts SDLC with no database, at real community scale — the existence proof that the v0.15 shape works. |
| AGENTS.md (AAIF standard) | https://agents.md | One canonical `AGENTS.md` for 30+ tools; `CLAUDE.md` only needs an `@AGENTS.md` import bridge. Why NG3 keeps CLAUDE.md as a bridge rather than deleting it. |

## Mechanism references

| Source | Where | What was taken |
|---|---|---|
| Anthropic — prompt caching | platform.claude.com, "Prompt caching" · *title only* | Prefix match is absolute; caches are model-scoped with no escape hatch on a model switch; dynamic data belongs at the end or nowhere. The reason every stage boundary in the 6-stage spine is a full cache rebuild, and the reason S7 measures cold entry. |
| Anthropic — Agent Skills | anthropic.com, "Agent Skills" · *title only* | Progressive disclosure — metadata always loaded, body on trigger, resources on demand. The shape the 6 spine `SKILL.md` files already follow and must keep after `p0-fail-open`. |
| System Design School — fail-open vs fail-closed | systemdesignschool.io · *title only* | Fail open for capacity and ceremony guards; fail closed only for correctness and security guards. The rule that produced S6's "exactly 2 fail-closed guards" count. |
| Sujeet Pro — graceful degradation | sujeet.pro · *title only* | Degradation must be graduated, explicit, and observable. Ancestor of v0.13's L0–L3 ladder; in v0.15 it survives as the marked optional index-sync block described in the merged plan's Approach. |

## In-repository evidence

These are live paths in this repository, not external links. They carried the decisive
evidence for the review and are cited by the merged plan's Authority section.

| Path | What it establishes |
|---|---|
| `docs/audit/sdlc-token-cache-audit.md` | Its own P1–P3 already shipped, so the −31% it forecast is already banked. It explicitly rejects putting the whole chain on one model and rejects dropping proof re-execution. This is what killed the original S7. |
| `docs/audit/consumer-adoption-audit.md` | D2's 2,595-token preflight packet — the baseline S7 measures against. |
| `cli/docs/CONTRACT.md` | The `check record` proof-verification guarantee the pre-commit hook must preserve verbatim, and `init`'s five jobs, four of which `install` absorbs. |
| `cli/internal/interfaces/root.go` | The authoritative command registration — the source R1's kill list was rebuilt from after the review found the CONTRACT-derived list wrong. |
| `docs/plans/completed/harness-markdown-truth.md` | Markdown is already the sole source of truth; `db rebuild` reconstructs the index from committed markdown alone, proving the database is disposable. |
| `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` | Durable memory shipped but no playbook calls it — why R13 moves memory to plain files. |

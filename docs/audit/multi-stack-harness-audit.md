# Multi-Stack Agent-Readiness Audit — mono-harness

**Date:** 2026-09-06

**Revision:** `aba7057a8466753faf487e935d37c4356105d114`

**Scope:** workflow skills, their procedure references, all six playbooks, entrypoints, and associated templates; Go, Rust, and Python consumers. Report and proposals only.

The shared lifecycle is substantially language-neutral. Its current automated gate explicitly defers to repository-defined commands; it does **not** prescribe npm, TypeScript, or a JavaScript test runner. The concrete JS-specific defect is the Git procedure's `package.json`-based dependency classification. Other readiness problems are incomplete verification context, consumer references to source-repository resources, and a size threshold that adds ceremony to small dependency changes across all three stacks.

The highest-severity finding is broader than language choice: the installed check playbook promises commit-time enforcement that the installer does not provide. A consumer with its own correct commands can still receive a misleading description of what enforces those commands.

## Findings by severity

`measured` below means source inspection plus a traced branch on the stated scenario, or exact byte measurement. It never means an agent session or native consumer build was executed. `hypothesis, unverified` marks predicted behavior that remains dependent on agent choices or consumer configuration.

| ID | Severity | Stack / reach | Finding | Evidence label | Main source |
|---|---|---|---|---|---|
| C1 | High | Go, Rust, Python; installed spine | The gate promises hook/CI proof re-execution without discovering whether either exists | `measured` | [check.md:40](../playbooks/check.md#L40) |
| C2 | Medium | Go, Rust, Python; full work, recovery, optional skills | Consumer instructions resolve helper/template/pattern paths that are not distributed there | `measured` | [work.md:32](../playbooks/work.md#L32), [encode-invariant/SKILL.md:8](../../skills/workflow/encode-invariant/SKILL.md#L8) |
| G2 | Medium | Go | A command string does not establish module/workspace coverage | `hypothesis, unverified` | [project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13), [check.md:34–37](../playbooks/check.md#L34) |
| R2 | Medium | Rust | Default Cargo success can omit the changed workspace member or feature | `hypothesis, unverified` | [to-plan.md:29–30](../playbooks/to-plan.md#L29), [check.md:34–37](../playbooks/check.md#L34) |
| P2 | Medium | Python | Checkout tests can be mistaken for proof of the installed package | `hypothesis, unverified` | [project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13), [check.md:28,34–37](../playbooks/check.md#L28) |
| R1 | Medium | Rust; optional Git skill | Cargo manifest and lockfile fall into different commit groups | `measured` | [git workflow:44–47](../../skills/workflow/git/references/workflow.md#L44) |
| P1 | Medium | Python; optional Git skill | Python manifests become code, requirements become docs, and locks become dependencies | `measured` | [git workflow:44–47](../../skills/workflow/git/references/workflow.md#L44) |
| G1 | Low | Go; optional Git skill | A dependency-only `go.mod` edit falls through to application code | `measured` | [git workflow:44–47](../../skills/workflow/git/references/workflow.md#L44) |
| G3 | Low | Go | Checksum churn alone rejects otherwise bounded work | `measured` | [work.md:17](../playbooks/work.md#L17) |
| R3 | Low | Rust | Cargo lockfile churn alone rejects otherwise bounded work | `measured` | [work.md:17](../playbooks/work.md#L17) |
| P3 | Low | Python | Python lockfile churn alone rejects otherwise bounded work | `measured` | [work.md:17](../playbooks/work.md#L17) |

These are **11 scenario findings, grouped into five causes**: enforcement claims, resource distribution, verification context, dependency classification, and size-based ceremony. G1/R1/P1 share one classification rule; G3/R3/P3 share one routing rule. They are not six independent implementation defects. Only the classification family is demonstrably JS-specific. The other families are portability or adequacy gaps that also affect some JS consumers.

## Rubric

An **agent-ready framework harness** lets a fresh agent identify authority, choose applicable tools and proof, execute within the approved scope, and leave evidence sufficient for another session to resume or reject the result. It must do this in the consuming repository, with proportionate context and process cost. Short prompts alone are insufficient; a large tool catalog alone is also insufficient.

This rubric extends the seven jobs and Level 0–3 ladder in [harness-engineering-gap-audit.md, sections 2–4](harness-engineering-gap-audit.md#2-what-the-article-actually-claims). CONTRACT, CONTEXT, TOOLS, STATE, SENSORS, POLICY, and TRACES remain necessary concerns. This audit separates command applicability from proof adequacy and adds explicit cost and maintenance axes, because a correct upstream procedure can still be expensive or unusable after distribution.

From [consumer-adoption-audit.md, Method/Caveats and sections 2–3](consumer-adoption-audit.md#3-cold-entry-cost-of-one-work-invocation), it reuses exact byte counts, a separately labeled token approximation, entry-layer frequency, bounded retrieval, and the distinction between observed evidence and inferred impact. Its old DB/preflight costs are not current measurements. Likewise, old H1–H9 findings are not automatically reopened: [ADR 0006:15–27](../decisions/0006-v015-authority.md#L15) makes the current architecture and contract authoritative over the historical 0.14 records.

**Scores describe content readiness, not native runtime certification:** 0 = absent or incompatible; 1 = partially specified, contradictory, or dependent on undocumented adaptation; 2 = coherent and supported by source tracing; 3 = additionally demonstrated in a representative consumer with native evidence and a fresh-session rerun. This audit cannot award 3. `N/A` is appropriate for responsibilities explicitly owned by the host, rather than inventing a harness feature to earn points.

| Axis | Prior job / extension | Why it matters; evidence required | Score / 3 | Current basis |
|---|---|---|---:|---|
| A. Authority and scope | CONTRACT + POLICY | Find accepted requirements and stop before inventing policy; preserve read-only and bounded intent | 2 | [AGENTS.md:5–19](../../AGENTS.md#L5), [brainstorm.md:34–47](../playbooks/brainstorm.md#L34) explicitly establish these boundaries |
| B. Context and applicability | CONTEXT | Route to the correct procedure; distinguish consumer files from harness-source files; recognize the consumer's manifests | 1 | Thin triggers work, but C2 and G1/R1/P1 leave gaps |
| C. Tool and environment contract | TOOLS | Identify the command owner, working directory, toolchain/environment, input scope, side effects, and failure recovery | 1 | [check.md:34](../playbooks/check.md#L34) respects repository commands; G2/R2/P2 expose unspecified execution context |
| D. Behavioral proof | SENSORS | Map affected behavior to tests/static analysis/runtime evidence; report omissions as well as exit status | 1 | [to-plan.md:29–30](../playbooks/to-plan.md#L29) requires advance proof; generic proof classes do not establish stack coverage (G2/R2/P2) |
| E. Durable state and recovery | STATE | One recoverable record, stable identities, bounded reads/retries, explicit blockers | 2 | [work.md:32–42](../playbooks/work.md#L32), [handoff.md:23–40](../playbooks/handoff.md#L23); helper availability is separately charged to B |
| F. Enforcement and permission boundary | POLICY + dual encoding | Distinguish requested procedure, installed local guard, CI, and externally verified merge blocking | 1 | Host ownership is explicit in [ARCHITECTURE.md:43](../ARCHITECTURE.md#L43); C1 overstates consumer enforcement |
| G. Traceability and reproducibility | TRACES | Record exact evidence, scope, judge, gaps, retries, and rollback; another agent can reproduce the claim | 1 | [check.md:28,38–40](../playbooks/check.md#L28) provides receipts; C1 and environment-sensitive proof limit reproducibility |
| H. Cost and ceremony | Cost-method extension | Measure entry bytes and mandatory steps; read only relevant material; scale process with risk/dependencies | 1 | Thin routing and batched writes help; G3/R3/P3 force durable work from raw line count; measurements below |
| I. Adaptation and maintenance | Failure-loop extension | Assign each gap to its owner; preserve consumer guidance across updates; require native proof and a fresh rerun before declaring improvement | 2 | [harness-improvement.md:12–33](../templates/harness-improvement.md#L12) and [ADR 0007:15–29](../decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md#L15) define ownership and the update boundary |

The same content scores apply to the three modeled stacks: no native consumer run supplies evidence for a different numerical rating. Do not sum them into a readiness percentage. A low cost cannot compensate for false enforcement or missing behavior proof.

**Relationship to Level 0–3:** the source protocol describes Level 2 durable work (structured state, checks, a bounded retry) and some Level 3 mechanisms in mono-harness itself. Bounded work deliberately needs less state. Installation into a consumer establishes documentation, not verified Level 2/3 operation: tests, host isolation, hooks, CI, recovery, and independent review must be observed there. This keeps the prior ladder while refusing to transfer source-repository guarantees to every consumer.

## Method and coverage

All findings use the current revision above. Every local link is relative to this audit; labels retain the relevant source line or section. Future edits may shift line numbers, so the revision and section descriptions are the durable reference.

The exhaustive content pass read:

- Root `AGENTS.md`, `CLAUDE.md`, README; `docs/WORKFLOW.md`; `skills/workflow/README.md`; architecture, current CLI contract, and every file directly under `docs/decisions/`.
- All **10** `skills/workflow/*/SKILL.md` files: brainstorm, to-plan, work, check, handoff, watzup, git, interview, encode-invariant, improve-harness. The README's older “8” count describes the original chain; both added skills were included.
- All **six** `docs/playbooks/*.md` files, in full; the Git skill's three references; the interview spec template; the invariant-encoding pattern; the harness-improvement template.
- All **six** embedded templates: project identity, plan, spec, run, check, handoff. Only project identity is an install target; the others were not treated as automatically executed instructions.
- Both prior audits named in the request and `docs/prompt-engineering-principles.md`. The latter's progressive disclosure and smallest-useful-instruction principles inform cost scoring; its model/tokenizer generalizations are not treated as measured facts here.

Targeted source inspection followed the references into installer target enumeration, embedded-document checks, the proof runner, plan slicer, hook core, and the checked-in CI workflow. These were read, not modified or installed. Six projected playbooks and WORKFLOW were byte-identical to embedded sources; the AGENTS body matched after removing the expected `ZHARNESS` marker lines. Thus findings against those playbooks also apply to the bytes the installer distributes.

Three distinct evidence operations were performed:

1. **Static branch traces:** assert the exact classification and bounded-mode rules still exist, then apply them to filenames and change sizes held in memory. This is a transcription of prompt logic, not execution of an agent or a compiler. The traces establish what a literal reading directs; a capable agent may override it using repository authority.
2. **Reference/ownership traces:** follow skill → playbook/reference → consumer-relative resource and compare it with the installer target set and README distribution promise.
3. **Byte measurements:** count file bytes and derive scenario entry totals. No conversation transcript, model cache, latency, dollar cost, or successful consumer task was measured.

No real or synthetic consumer repositories were created, no `zharness install` ran, and no Go/Rust/Python build or test was executed. Official stack documentation was read to constrain scenario semantics; it is not evidence that a consumer failed. Host configuration, credentials, production systems, and private consumer repos were outside this audit.

## Shared findings

### C1 — Installed prose promises enforcement that installation does not establish

**Severity:** High. **Evidence:** `measured`. **Earliest gap:** proof. **Owner:** repository-harness.

**Scenario and trace:** take a Go, Rust, or Python repository with native verification commands and no zharness-specific hook/CI integration. The check skill routes to the local playbook ([check/SKILL.md:16](../../skills/workflow/check/SKILL.md#L16)). [check.md:40](../playbooks/check.md#L40) says the repository's pre-commit hook re-executes every nested proof and CI re-runs it; [check.md:38](../playbooks/check.md#L38) similarly states that the hook rejects a same-session full check. But [README.md:72–74](../../README.md#L72) explicitly excludes installing hooks. [installer.go:66–75](../../cli/internal/installer/installer.go#L66) enumerates WORKFLOW, PROJECT, and playbooks; the remaining install body handles AGENTS, ignore entries, and base tracking ([installer.go:354–398](../../cli/internal/installer/installer.go#L354)). No hook or CI workflow is supplied.

**Observed boundary:** the promise and distribution set contradict each other for that stated consumer. No commit bypass was attempted. Existing consumer hooks could supply equivalent enforcement; merely having some pre-commit hook is not evidence that these particular checks run. Source-repository guard code and CI do not prove consumer enforcement or branch protection.

**Proposal:** qualify the playbook guarantee by discovered enforcement: local command / hook / CI / merge blocking. Reuse the distinctions already in [encoding-invariants.md:59–71](../patterns/encoding-invariants.md#L59). Missing hooks should yield an honest enforcement gap, not automatically install anything or block otherwise authorized bounded work.

**Future acceptance:** a fresh session in an existing consumer without these guards reports “not installed/unverified”; a consumer with observed equivalent guards reports the exact mechanism. Only a separately authorized negative probe can establish rejection of a false proof. Until then, improvement remains unverified.

### C2 — Optional skills and installed stages refer to resources absent from consumer distributions

**Severity:** Medium. **Evidence:** `measured`. **Earliest gap:** context/capability. **Owner:** repository-harness.

**Scenario and trace:** a consumer has the managed docs and separately installed workflow skills, but does not contain mono-harness's source tree. Full work names `bash scripts/plan-slice.sh` without an absence branch ([work.md:32](../playbooks/work.md#L32)); watzup merely prefers that helper ([watzup.md:15](../playbooks/watzup.md#L15)). If PROJECT is missing, brainstorm tells the agent to copy from `cli/docs/embedded/templates/project.identity.md` ([brainstorm.md:39](../playbooks/brainstorm.md#L39)). The non-spine skills directly require repo-relative `docs/patterns/encoding-invariants.md` and `docs/templates/harness-improvement.md` ([encode-invariant/SKILL.md:8](../../skills/workflow/encode-invariant/SKILL.md#L8), [improve-harness/SKILL.md:8](../../skills/workflow/improve-harness/SKILL.md#L8)). None is a target in [installer.go:66–75](../../cli/internal/installer/installer.go#L66), nor bundled under those skills' directories.

**Observed boundary:** those paths resolve in mono-harness, not necessarily in the consumer. A normal installation already creates PROJECT, so the identity-template branch is recovery/skill-only use, not an inevitable first-lock failure. Work may recover by reading sections directly. The check runner already has the correct pattern: explicit present/absent behavior at [check.md:34](../playbooks/check.md#L34).

For a Go multi-module plan, a Rust feature-matrix plan, or a Python packaging plan, this creates the same failed lookup before useful work. It is a source-checkout assumption, not a JS-only assumption. The invariant pattern even lists this repository's Go CLI and guard scripts as “typical owners here” ([encoding-invariants.md:37–45](../patterns/encoding-invariants.md#L37)); those examples must not become consumer defaults.

**Proposal:** give the slicer a documented direct-read fallback; bundle optional-skill procedure assets under the skill and resolve them relative to its base directory; make PROJECT recovery name an actually available source. Preserve consumer ownership of native commands.

**Future acceptance:** with only the documented managed set and chosen skill payload available, each affected route resolves its own instructions or takes a named fallback. No consumer needs a fake `cli/` directory, copied Go gate, or additional package manager. This audit verified the distribution mismatch, not installation behavior.

## Go findings

### G1 — Dependency-only go.mod edits fall through to application code

**Severity:** Low. **Evidence:** `measured`. **Earliest gap:** domain. **Owner:** repository-harness, optional Git skill.

**Scenario:** the approved change modifies only a dependency version in `go.mod`. The module file represents dependencies as well as module/toolchain properties ([Go module-file reference](https://go.dev/doc/modules/gomod-ref)); this scenario specifically concerns a dependency edit.

**Trace:** [git/SKILL.md:13](../../skills/workflow/git/SKILL.md#L13) loads its workflow. At [git workflow:44](../../skills/workflow/git/references/workflow.md#L44), `go.mod` is neither docs, a test path, `.claude/` configuration, `package.json`, nor a lockfile. The explicit “code for everything else” branch applies. Line 47 associates that group with `feat`/`fix`, whereas the same dependency edit in `package.json` enters `chore(deps)`.

**Result:** the literal routing misclassifies the change. No commit was created and no release/versioning consequence was observed. `go.sum` is a checksum record, not a conventional lockfile; its classification is also unspecified and should not be asserted to work merely because “lockfiles” is present. This is lower severity than splitting an atomic manifest/lock pair.

**Proposal and acceptance:** classify dependency edits by content and manifest role, keeping relevant checksum changes with them. A fresh Git-skill trace should group an approved dependency-only `go.mod`/`go.sum` change coherently without labeling all future `go.mod` edits as dependencies. Toolchain or module-path changes can warrant another scope.

### G2 — Exact command text does not establish module/workspace scope

**Severity:** Medium. **Evidence:** `hypothesis, unverified`. **Earliest gap:** proof/environment. **Owner:** consumer owns the matrix; repository-harness owns prompting for its scope.

**Scenario:** a repository contains two independently relevant modules and a workspace. A change affects module B; PROJECT's test answer or an inherited command was written for module A. Alternatively, a local workspace replacement changes what the same command exercises. Go's workspace selection depends on `GOWORK`, the working directory, and enclosing `go.work` files ([Go workspace reference](https://go.dev/ref/mod#workspaces)).

**Trace:** [project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13) asks only for exact verification command(s). [to-plan.md:29–30](../playbooks/to-plan.md#L29) locks observable commands; [work.md:37](../playbooks/work.md#L37) runs them exactly; [check.md:34–37](../playbooks/check.md#L34) classifies applicability and unit/integration/command evidence. None explicitly asks which modules, workspace replacements, build constraints, or working directory that command covers. Repeating a passing `go test ./...` in the wrong scope would reproduce the same omission, not repair it.

**Counterevidence and limit:** check first reads repository verification instructions and compares planned requirements. A correct repo wrapper, explicit `cd`, or thorough agent review can already close this gap. The harness does not order a root-level `go test ./...`, and this audit observed neither that command failing nor a missed Go defect. The hypothesis is a missing coverage prompt, not absence of Go support.

**Proposal and acceptance:** require planned proof to name its affected module/workspace scope and source of authority. Discover existing CI/Makefile commands before suggesting `go test` or `go vet`; record relevant workspace/build constraints only when the change touches them. A future representative run must identify and execute proof for B, and name any intentionally omitted modules. A fresh agent doing this correctly without new guidance would weaken the proposal.

### G3 — Checksum churn can force durable planning for bounded maintenance

**Severity:** Low. **Evidence:** `measured`. **Earliest gap:** authority/cost. **Owner:** repository-harness.

**Scenario:** known subsystem, known verification, no coordination or open policy decision, two files: 2 changed manifest lines plus 200 changed `go.sum` lines. These counts are chosen inputs, not an observed dependency update or a claim about its typical size.

**Trace:** [work.md:17](../playbooks/work.md#L17) rejects bounded mode above roughly 100 changed lines independently of file count or proof availability. At 202 lines, this scenario takes the brainstorm/to-plan route. [brainstorm.md:39–44](../playbooks/brainstorm.md#L39) then requires answered identity and an active plan; [handoff.md:32](../playbooks/handoff.md#L32) ultimately requires the durable initiative's independent full check.

**Result:** the text adds durable ceremony based solely on the supplied raw change size. Actual agent compliance, elapsed time, and added billed tokens are unmeasured. This rule is language-neutral and can penalize JS lockfile changes too; it is not evidence of an npm preference.

**Proposal and acceptance:** treat file/line counts as a review signal alongside authored behavior, dependency risk, generated/checksum churn, and proof availability. Preserve escalation for an unsafe dependency change regardless of its size. A future low-risk checksum scenario should remain bounded without a plan; a same-size risky migration should still escalate.

## Rust findings

### R1 — Cargo.toml and Cargo.lock are split by the filename taxonomy

**Severity:** Medium. **Evidence:** `measured`. **Earliest gap:** domain. **Owner:** repository-harness, optional Git skill.

**Scenario:** an approved dependency update changes the relevant `Cargo.toml` requirement and corresponding tracked `Cargo.lock`, with no independent application feature.

**Trace:** the exact grouping rule at [git workflow:44–47](../../skills/workflow/git/references/workflow.md#L44) sends `Cargo.toml` to the fallback `code` group and `Cargo.lock` to `deps`. Mixed groups are directed into separate commits, even though this scenario is a single dependency update. The skill reaches that rule unconditionally through [git/SKILL.md:13](../../skills/workflow/git/SKILL.md#L13).

**Result:** the literal procedure separates the authored dependency requirement from its associated resolution file. An intermediate commit can require lockfile reconciliation; whether it actually fails consumer checks depends on the changed requirement and verification policy. No Cargo command or intermediate commit was executed. Medium severity reflects the loss of an atomic dependency-change unit, not a demonstrated build failure.

**Proposal and acceptance:** classify related manifest/lock edits as one coherent dependency change, using the diff and consumer convention. A future trace should keep this pair together and still distinguish unrelated feature/toolchain edits in the manifest. Do not make all TOML files dependencies or add npm scripts around Cargo.

### R2 — Default Cargo proof can omit the changed member or feature

**Severity:** Medium. **Evidence:** `hypothesis, unverified`. **Earliest gap:** proof. **Owner:** consumer owns its supported matrix; repository-harness owns the coverage question.

**Scenario:** a workspace's default members exclude the changed crate, or changed code is behind a feature that is not enabled by default. Existing shorthand verification says `cargo test`; static-analysis shorthand is similarly narrower than the accepted change.

**Trace:** [to-plan.md:29–30](../playbooks/to-plan.md#L29) requires commands before work, and [check.md:34–37](../playbooks/check.md#L34) asks for applicable tests/static analysis/build plus proof classes. Neither requires an explicit member/feature/target coverage statement. Cargo chooses packages from the selected manifest and workspace defaults, and enables default features unless otherwise directed ([Cargo test: package and feature selection](https://doc.rust-lang.org/cargo/commands/cargo-test.html)). A green default invocation can therefore be irrelevant to the selected scenario.

**Counterevidence and limit:** a plain virtual workspace without a restricted default-member set is not evidence of omission; Cargo can already select all its members. Existing CI or a plan with explicit features can satisfy the current playbook. Missing planned proof must already be reported at check step 4. The unverified risk is treating familiar command names as adequate coverage when planning never identified the relevant configuration.

**Proposal and acceptance:** discover the repository's supported members, features, targets, and existing test/Clippy commands only as relevant to the change. Record intentional omissions and separate build-only evidence from executed tests. Do not prescribe `--all-features` universally; the supported combinations belong to the consumer. A future run must exercise the changed member/feature or explicitly report a proof gap; default-success alone must not certify it.

### R3 — Cargo lockfile size becomes a lifecycle decision

**Severity:** Low. **Evidence:** `measured`. **Earliest gap:** authority/cost. **Owner:** repository-harness.

**Scenario:** a known, bounded dependency adjustment changes 2 manifest lines and 200 `Cargo.lock` lines across two files, with an established native verification path. The counts are hypothetical inputs used to exercise the rule.

**Trace:** [work.md:17](../playbooks/work.md#L17) rejects bounded mode at 202 lines. It does not distinguish resolution-file churn from authored behavior. That routes through the same identity/plan write at [brainstorm.md:39–44](../playbooks/brainstorm.md#L39) and final independent review at [handoff.md:32](../playbooks/handoff.md#L32), even when size is the only failing eligibility condition.

**Result:** mandatory process expands for this supplied input. It is the same cause as G3, not evidence that Rust lockfiles are always large or safe. A dependency change involving build scripts, native libraries, or compatibility can deserve deeper review regardless of line count.

**Proposal and acceptance:** make size a signal rather than the sole rejection reason, while preserving dependency-risk review. In a future paired run, a low-risk lock refresh should remain bounded and a behaviorally risky update should escalate. No runtime or token savings are claimed from this source trace.

## Python findings

### P1 — Dependency inputs are split into code, dependencies, and documentation

**Severity:** Medium. **Evidence:** `measured`. **Earliest gap:** domain. **Owner:** repository-harness, optional Git skill.

**Scenario:** an approved dependency update changes `pyproject.toml` and `uv.lock`; a requirements-based variant changes `requirements.txt`. These are alternative consumer layouts, not a requirement that every Python project contain all three files.

**Trace:** [git workflow:44–47](../../skills/workflow/git/references/workflow.md#L44) recognizes `package.json` and lockfiles as dependencies. Its literal fallback makes `pyproject.toml` code; `uv.lock` is dependencies; its `.txt` branch makes `requirements.txt` documentation. The first scenario becomes multiple commit groups; the second is labeled docs despite changing installation inputs.

**Result:** the text fails both modern manifest and requirements-based cases. No actual commit or downstream automation was run. The taxonomy also cannot be fixed by treating every `pyproject.toml` change as dependencies: that file may contain build-system, project, and tool configuration ([PyPA pyproject guide](https://packaging.python.org/en/latest/guides/writing-pyproject-toml/)).

**Proposal and acceptance:** classify the changed content, preserve coherent dependency updates, and recognize consumer-owned requirements/constraints files without assuming all text files are docs. A future trace should keep the relevant dependency files together while classifying a tool-only configuration edit separately. Do not require Python consumers to create `package.json` to receive dependency handling.

### P2 — Checkout verification does not establish installed-package correctness

**Severity:** Medium. **Evidence:** `hypothesis, unverified`. **Earliest gap:** proof/environment. **Owner:** consumer owns supported environments and package behavior; repository-harness owns evidence scoping.

**Scenario:** a packaging change affects included modules or package data. Tests run from the source checkout and pass, while the built distribution omits a required file. A related reproduction risk is using a different interpreter/environment when rerunning a bare `pytest` command.

**Trace:** [project.identity.md:13–14](../../cli/docs/embedded/templates/project.identity.md#L13) accepts a command-only answer. [check.md:28,34–37](../playbooks/check.md#L28) records commands/results and asks for applicable tests, types, lint, and build, but never explicitly asks whether imports came from the checkout or installed artifact. A package build plus passing source tests would not necessarily exercise installation. pytest's own guidance distinguishes installed-package testing from checkout testing to expose packaging defects ([pytest integration practices](https://docs.pytest.org/en/stable/explanation/goodpractices.html#tox)).

**Counterevidence and limit:** repository instructions may already use tox, nox, an explicit interpreter, or an environment-manager wrapper; the current playbook can follow those. `ruff` or `mypy` are applicable only when the consumer defines them; their absence is not a defect, nor is every Python change a packaging change. No interpreter mismatch, missing artifact, or false green test was observed here.

**Proposal and acceptance:** for a packaging change, identify the authoritative environment and whether the proof covers the installed distribution. Reuse the consumer's existing verification command; include interpreter/environment selection and working directory in reproducible evidence when material. A future representative run must catch the missing-package-file case or report installation proof as missing. An existing wrapper that already catches it weakens the need for new harness guidance.

### P3 — Python lock churn triggers the same unconditional escalation

**Severity:** Low. **Evidence:** `measured`. **Earliest gap:** authority/cost. **Owner:** repository-harness.

**Scenario:** a known low-risk update changes 2 dependency lines in `pyproject.toml` and 200 lines in the consumer's tracked `uv.lock`. It has a known environment and verification path and no coordination requirement. Counts are chosen branch inputs; uv adoption or this amount of churn was not measured in a consumer.

**Trace:** [work.md:17](../playbooks/work.md#L17) rejects the 202-line diff even though it spans only two known files. [brainstorm.md:39–44](../playbooks/brainstorm.md#L39) introduces the durable lock and [handoff.md:32](../playbooks/handoff.md#L32) requires the final full independent check. No rule discounts generated lockfile lines when deciding eligibility.

**Result:** this is the Python manifestation of G3/R3, with the same static evidence limit. It neither establishes a typical Python cost nor licenses treating all resolver output as low-risk.

**Proposal and acceptance:** use semantic dependency risk and proof availability alongside file size. A future low-risk lock update should stay bounded; changes to native extensions, interpreter support, or supported dependency ranges can still justify escalation. Compare both cases in a fresh session before adopting the policy change.

## Cost and ceremony

**Evidence:** `measured` file bytes; modeled entry totals and token estimates, not observed session consumption. This is rubric axis H, one of nine axes.

The following packet is defined as root AGENTS + WORKFLOW + one selected SKILL + its playbook. It is a reproducible scenario, not a universal loading requirement. Direct playbook use can omit the skill; an agent may already have entrypoints cached; a host may inject other instructions.

| Stage | Skill bytes | Playbook bytes | Packet bytes | Approx. tokens (`bytes / 4`, rounded) |
|---|---:|---:|---:|---:|
| watzup | 1,010 | 2,316 | 6,239 | 1,560 |
| brainstorm | 1,269 | 5,797 | 9,979 | 2,495 |
| to-plan | 1,104 | 4,544 | 8,561 | 2,140 |
| work | 1,189 | 8,848 | 12,950 | 3,238 |
| check | 1,157 | 9,595 | 13,665 | 3,416 |
| handoff | 1,068 | 6,297 | 10,278 | 2,570 |

Common packet bytes are **2,913**: AGENTS 1,156 + WORKFLOW 1,757. If work's in-session gate reads the complete check playbook once, the unique packet becomes **22,545 bytes**, approximately **5,636 tokens**. This excludes product code, commands/output, plan slices, PROJECT, global instructions, tools, and conversation history. Adding the source repo's CLAUDE would add 6,361 bytes, approximately 1,590 tokens, only on hosts that load it; it is not an installed consumer requirement.

All ten workflow skills total **12,692 bytes**; all six playbooks total **37,397 bytes**. The workflow does not tell ordinary tasks to read all of them. This audit did so because its scope explicitly required exhaustive depth. The optional Git skill plus its mandatory workflow reference is a separate **9,152-byte** entry (2,511 + 6,641), about **2,288 tokens**, before Git output or optional references.

Exact byte counts are reproducible; token estimates are not tokenizer measurements. Unicode bytes, tokenization, host loading, and caching can alter actual consumption. The prior audit's approximate error margin was not calibrated here, so no precise accuracy bound or dollar saving is claimed.

```bash
python3 - <<'PY'
from pathlib import Path

entry = sum(Path(p).stat().st_size for p in ['AGENTS.md', 'docs/WORKFLOW.md'])
for stage in ['watzup', 'brainstorm', 'to-plan', 'work', 'check', 'handoff']:
    skill = Path(f'skills/workflow/{stage}/SKILL.md').stat().st_size
    playbook = Path(f'docs/playbooks/{stage}.md').stat().st_size
    total = entry + skill + playbook
    print(stage, skill, playbook, total, round(total / 4))
PY
```

**Ceremony trace:** the G3/R3/P3 input adds identity completion and plan lock, phase/task definition, Progress and Current State maintenance, durable Validation/receipt, and final independent review/closure. These requirements come from [brainstorm.md:39–47](../playbooks/brainstorm.md#L39), [to-plan.md:25–33](../playbooks/to-plan.md#L25), [work.md:34–42](../playbooks/work.md#L34), [check.md:28,39–44](../playbooks/check.md#L28), and [handoff.md:31–38](../playbooks/handoff.md#L31). Their exact tool-call count and runtime cost were not measured.

Existing cost controls should survive any changes: one selected playbook, section reads, one batched Progress flush per successful wave, bounded retries, no mandatory memory write for every handoff, an in-session phase gate, and one final independent full review. Large-plan context remains a separate concern: “the last few bullet lines” in [watzup.md:15](../playbooks/watzup.md#L15) and “tails” in [work.md:32](../playbooks/work.md#L32) do not provide numerical limits or an omitted-content count. This is an unquantified retrieval limit, not a measured recurrence of the old 26,400-token whole-plan fallback.

## What this audit does not count as a stack defect

| Surface reviewed | Conclusion and boundary |
|---|---|
| Native gates | [check.md:32–37](../playbooks/check.md#L32) explicitly uses repository-defined applicable checks. `go test`/`go vet`, Cargo tests/Clippy, and pytest/Ruff/mypy can fit that contract without adding Node. Separate “types” and “build” labels do not mandate a redundant command for a compiled language. |
| Root CLAUDE commands | [CLAUDE.md:57–78](../../CLAUDE.md#L57) contains skill-install examples and this source repo's doc-link/Go gates. They are local development instructions, not installed consumer defaults. The optional `npx skills add` distribution path does not prove lifecycle dependence on npm; [README.md:72–74](../../README.md#L72) says skills are separate. |
| Specialized shipping skills | No audited spine skill or playbook mandates the optional TypeScript-specific shipping skill. Its existence is not evidence that Go/Rust/Python consumers must adopt its stack. |
| Legacy embedded templates | The run/check/handoff templates contain old artifact fields and `.kit` paths (for example [templates/check.md:1–22](../../cli/docs/embedded/templates/check.md#L1)). Current stages do not route to them, and the installer target set excludes them. They are latent source residue, not counted as a demonstrated consumer failure or as three stack findings. |
| Single active plan and host permissions | [ADR 0006](../decisions/0006-v015-authority.md) preserves the plan invariant without a database; [ARCHITECTURE.md:43](../ARCHITECTURE.md#L43) delegates tool permissions to the host. Neither policy is a JS assumption. Actual host isolation and external merge blocking remain unverified. |
| Deployment and mobile | [skills/workflow/README.md, SDLC Stage Coverage](../../skills/workflow/README.md#sdlc-stage-coverage) excludes deployment/release/production monitoring. Mobile was excluded by this audit's scope. Their absence does not lower the scores. Local packaging proof for an accepted packaging change still belongs to verification. |

## Proposals for later review

No proposal below is implemented or authorizes a policy change. In particular, relaxing the bounded-mode threshold requires owner acceptance, and installing hooks/CI requires separate authorization. [ADR 0007:15–29](../decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md#L15) means a consumer-local patch to a managed playbook will be overwritten; upstream procedure changes belong in embedded sources, while consumer commands belong in consumer-owned guidance or PROJECT.

| Priority | Smallest intervention | Owner / future edit surface | Evidence that would weaken it; future acceptance |
|---|---|---|---|
| 1 | Qualify hook/CI guarantees by discovered enforcement (C1) | Harness; embedded check playbook | A separately established consumer enforcement contract could close the gap. Absent/present consumers must report their actual level accurately. |
| 2 | Resolve missing resources from the distributed skill or a named fallback (C2) | Harness; affected skills and embedded work/brainstorm playbooks | An independently verified distribution channel supplying those exact paths would weaken the finding. Routes must work with only documented payloads. |
| 3 | Classify dependency edits by semantic role and preserve atomic updates (G1/R1/P1) | Harness; Git skill's own workflow reference | A consumer override may already correct the classification. The filename scenarios must produce coherent groups without misclassifying unrelated TOML changes. |
| 4 | Add one short coverage question to planned proof, with optional stack examples (G2/R2/P2) | Harness owns the question; consumer owns commands/matrix | Fresh baseline agents consistently selecting correct scope would weaken the need for extra instructions. Test the module, feature, and installed-package scenarios before adding mandatory guidance. |
| 5 | Revisit raw-size rejection while retaining semantic-risk escalation (G3/R3/P3) | Owner decides policy; harness implements it later | Evidence that the extra process catches material failures in these low-risk cases would favor retaining it. Compare low-risk and risky changes at equal line counts. |

For any accepted intervention, use [harness-improvement.md](../templates/harness-improvement.md): record the representative job, baseline failure and human steering, revision/tools, earliest gap and owner, one falsifiable intervention, native proof, and a **different-session** rerun. All proposals currently have **Decision: pending fresh rerun**. Their maintenance owner is the surface owner above; remove added guidance if it fails to change behavior, duplicates consumer authority, or a verified distribution/validation mechanism makes it unnecessary. Do not write a new plan until separately requested.

## Validation and limits

The starting working tree already contained untracked `multi-stack-harness-audit-prompt.md`. It must remain unchanged. “Exactly one new file” therefore means one addition **relative to that starting state**, not erasing the pre-existing prompt to make status look clean.

Required checks for this artifact:

- `bash scripts/verify-doc-links.sh` against the finished audit, including its untracked file. The script enumerates Markdown files from disk, not only the Git index.
- Resolve every local Markdown link and source line reference; verify every finding has a severity, evidence label, source citation, scenario, owner, and proposed acceptance.
- Confirm three Go, three Rust, three Python findings, explicit Rubric criteria and prior-framework relationship, and a scored cost axis.
- Compare starting file hashes/status with final state; inspect `git diff --stat`, `git diff --cached --stat`, and the new file with `git diff --no-index --stat /dev/null docs/audit/multi-stack-harness-audit.md`. An untracked addition is absent from ordinary `git diff --stat`; no staging is needed to inspect it.

**Recorded results:** the doc-link gate passed with `doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`. Those known-removed claims belong to existing historical material. A separate check passed all 81 local links/anchors and verified 11 labeled findings: 8 measured traces, 3 hypotheses, with three findings per stack. File-hash comparison covered 254 pre-existing tracked/untracked non-ignored files: none changed or disappeared, and the audit was the only addition. Ordinary and staged diff stats were empty; the no-index diff showed one added file (its exit code 1 denotes a difference, not a verifier failure).

No native consumer validation, fresh-session experiment, benchmark, security audit of the host, install, commit, or PR is claimed. Completion of this report establishes a cited audit and its checks; it does not establish that any proposed harness improvement works.

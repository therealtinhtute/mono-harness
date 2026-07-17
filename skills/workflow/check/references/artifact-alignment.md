# Artifact Alignment Gate

Use this after `work` or whenever the diff claims to follow a locked plan.

## Required Artifacts (full harness flow)

For a phase gate, inspect these before approving:
- `.kit/planning/SPEC.md`
- `.kit/planning/ROADMAP.md`
- `.kit/planning/phases/{slug}/{slug}-CONTEXT.md`
- `.kit/planning/phases/{slug}/{slug}-PLAN.md`
- latest `.kit/runs/work/*.md` for the same phase when execution happened through `work`

If the repo is not using the full harness flow, say so explicitly and skip artifact alignment instead of pretending it passed.

## Alignment Questions

### 1) Spec Alignment
- Does the diff implement behavior that maps to the spec requirements?
- Did the implementation quietly add behavior outside spec scope?
- Are there requirement-shaped gaps the diff does not cover?

### 2) Phase Boundary Alignment
- Do changed files stay inside `Allowed Surfaces` and task `touches`?
- Did the work cross into `Forbidden Surfaces` or task `avoid` paths?
- If the phase plan expected one subsystem, did the diff spread across others?

### 3) Execution Proof Alignment
- Did each materially changed behavior have a matching verification command in the phase plan?
- Does the work run log show those verification commands actually ran?
- Did the diff add behavior with no proof trail?
- When `zharness` applies: does gathered proof satisfy every `required` cell in `gate-checklist.md`'s Validation Matrix for the resolved lane? A required cell with no matching evidence is a proof gap here, not just a Phase 1 gate concern — name it the same way (missing evidence class).

### 4) Decision / Context Alignment
- Did implementation contradict locked decisions in the phase context?
- Were rejected options reintroduced in code?
- Were new assumptions added without being surfaced?

## Verdict Mapping

| Finding | Severity | Merge impact |
|--------|----------|--------------|
| Code contradicts spec requirement | 🔴 Critical | Block |
| Changed files exceed phase boundary | 🟠 Major | Request changes |
| Missing or weak verification evidence | 🟠 Major | Request changes |
| Small context drift, documented and harmless | 🟡 Minor | Approve with note |
| Artifact missing because harness not used | 💡 Suggestion | Note only |

## Suggested Output Snippets

### Clean alignment
```text
artifact_alignment: ✅ aligned
- spec requirements covered: {list}
- phase boundary respected: yes
- run proof present: yes
```

### Misalignment
```text
artifact_alignment: ❌ drift
- spec gap: {requirement}
- boundary drift: {file/path}
- proof gap: {missing command or missing run log}
next: refresh `/to-plan phase {slug}` or rerun `/work` with verification
```

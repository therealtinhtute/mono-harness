# Interview Execution Modes

## Fast Mode

**Purpose:** Quick validation of critical decisions before implementation.

**Scope:**
- 1-2 interview rounds
- Focus on critical decisions only
- No spec file generation

**Question Focus:**
1. **Goal & Approach** (50%)
   - What is being built and why?
   - How will it be built?
   - What's the chosen approach and rationale?

2. **Critical Risks** (50%)
   - What breaks if assumption X is wrong?
   - What's the rollback strategy?
   - What's the blast radius of failure?

**Output:**
Console summary only:
```
✓ Fast interview complete
✓ 6 questions answered across 2 rounds

Critical validations:
  • Goal: {validated-goal}
  • Approach: {validated-approach}
  • Key risk: {identified-risk}

Next: Proceed with implementation or run deep mode for full validation
```

**When to use:**
- Quick sanity check before starting work
- Validating small changes or bug fixes
- Time-constrained situations
- When plan is already well-defined

---

## Deep Mode (Default)

**Purpose:** Comprehensive plan validation with full spec generation.

**Scope:**
- 3-5 interview rounds
- All 5 question categories
- Write validated spec to plan file

**Question Categories:**
1. **Technical Implementation** (30%)
2. **UI/UX & User Flows** (20%)
3. **Architecture & Design** (20%)
4. **Concerns & Risks** (15%)
5. **Tradeoffs & Alternatives** (15%)

**Output:**
Console summary + validated spec file:
```
✓ Interview complete
✓ 12 questions answered across 3 rounds
✓ Spec validated and written to {plan-file}

Key insights:
  • {insight-1}
  • {insight-2}

Next: Review spec, then implement
```

**When to use:**
- New features or major changes
- Complex architectural decisions
- Before starting multi-day work
- When requirements are unclear
- PR planning and milestone work

---

## Mode Selection

**Usage:**
```bash
/interview                           # deep mode, search for plans
/interview plans/auth.md             # deep mode on specific plan
/interview plans/auth.md fast        # fast mode on specific plan
/interview fast                      # fast mode, search for plans
```

**Comparison:**

| Aspect | Fast | Deep |
|--------|------|------|
| Rounds | 1-2 | 3-5 |
| Questions | 4-8 | 12-20 |
| Categories | 2 (goal, risks) | 5 (all) |
| Output | Console only | Console + spec file |
| Duration | 5-10 min | 15-30 min |
| Spec generation | No | Yes |

# Interview Skill Examples

## Example 1: Standard Plan Interview
**Scenario**: Interview about authentication system plan

**Input**:
```bash
/interview plans/auth-system.md
```

**Output**:
```
📋 Interviewing: Authentication System

Round 1/~4
  ✓ 0 questions answered
  ⏳ 5 areas to explore

[AskUserQuestion with 4 questions about technical approach, token storage, session management, error handling]
```

**Explanation**: The skill loads the plan, identifies key areas to explore, and starts with high-level technical questions using AskUserQuestion tool.

---

## Example 2: UI/UX Deep Dive
**Scenario**: Explore user flows and edge cases for checkout process

**Input**:
```bash
/interview plans/checkout-flow.md
```

**Output**:
```
📋 Interviewing: Checkout Flow

Round 2/~3
  ✓ 4 questions answered
  ⏳ 3 areas to explore

Questions about:
  • What does user see when payment fails?
  • How does user recover from expired session?
  • What happens if inventory changes during checkout?
  • How are partial refunds handled?
```

**Explanation**: After initial technical questions, the skill drills into UI/UX edge cases and error recovery flows.

---

## Example 3: Architecture Validation
**Scenario**: Validate architecture decisions and tradeoffs

**Input**:
```bash
/interview plans/microservices-migration.md
```

**Output**:
```
📋 Interviewing: Microservices Migration

Round 3/~5
  ✓ 8 questions answered
  ⏳ 2 areas to explore

Exploring tradeoffs:
  • Why microservices over modular monolith?
  • How does this handle distributed transactions?
  • What's the deployment complexity cost?
  • How do we maintain data consistency?
```

**Explanation**: The skill explores architectural tradeoffs and validates that the chosen approach addresses key concerns.

---

## Example 4: Complete Interview
**Scenario**: Full interview cycle from start to validated spec

**Input**:
```bash
/interview plans/api-rate-limiting.md
```

**Output**:
```
✓ Interview complete
✓ 12 questions answered across 3 rounds
✓ Spec validated and written to plans/api-rate-limiting.md

Key insights:
  • Token bucket algorithm chosen over sliding window for memory efficiency
  • Rate limits stored in Redis with 1-hour TTL
  • Graceful degradation if Redis unavailable (allow all requests)
  • Per-user and per-IP limits with different thresholds

Next: Review spec, then implement
```

**Explanation**: After 3 rounds of questions, the skill has validated all aspects of the plan and written a comprehensive spec with decisions, tradeoffs, and implementation details.

---

## Example 5: Plan File Search
**Scenario**: No plan file provided, search and select

**Input**:
```bash
/interview
```

**Output**:
```
📋 Found 3 recent plan files:

[AskUserQuestion with options:]
  • plans/auth-system.md — Authentication system with JWT
  • plans/checkout-flow.md — E-commerce checkout process
  • tasks/api-migration.md — REST to GraphQL migration
```

**Explanation**: When no file is provided, the skill searches for recent plan files and presents them for selection via AskUserQuestion.

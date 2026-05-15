# Autoplan Core Examples

## Example 1 — idea-first

### Input

```text
I want a small internal dashboard that shows failed cron jobs, recent deploys, and a simple daily summary. Not sure what stack to use yet. Keep it cheap and fast to ship.
```

### Output

```text
Goal
Build a small internal ops dashboard for cron failures, deploy visibility, and a daily summary.

Context
Cheap, fast-to-ship internal tool; stack still open.

Assumptions
- Internal-only access is fine.
- Cron/deploy data can be read from existing logs or APIs.

Open questions
- Where do cron failures surface now?
- What system is the deploy source of truth?

Scope
- Cron failure list
- Recent deploy history
- Daily summary panel

Non-goals
- Full observability platform
- Public-facing UI

Constraints
- Low cost
- Fast initial delivery

Spec
Create an internal dashboard that aggregates cron failures, recent deploy events, and a simple daily summary.

Phases
1. Source audit
2. MVP design
3. Implementation
4. Verification

Tasks
- Audit current cron and deploy data sources.
- Choose the smallest viable stack.
- Implement cron/deploy panels.
- Add daily summary path.
- Smoke test with live sample data.

Risks
- Event sources may be inconsistent.

Next recommended action
Audit sources first, then lock the smallest stack that can ship in one pass.
```

## Example 2 — spec-first

### Input

```text
Here is my markdown spec for a Telegram bot that summarizes GitHub notifications. Please turn it into a realistic implementation plan and point out missing parts.
```

### Output

```text
Goal
Turn the existing Telegram bot spec into an execution-ready implementation plan.

Context
The user already has a markdown spec and wants gap analysis plus a build plan.

Assumptions
- GitHub notifications can be accessed via API.
- Telegram is the delivery channel.

Open questions
- What auth flow will be used for GitHub?
- On-demand, scheduled, or both?

Scope
- Spec audit
- Missing decisions
- Implementation phases
- Concrete tasks

Non-goals
- Full multi-channel notification platform

Constraints
- Must fit current OpenClaw/Telegram workflow.

Spec
Build a Telegram bot flow that reads GitHub notifications, summarizes them into grouped updates, and delivers them with controllable frequency.

Phases
1. Spec audit
2. Integration design
3. Bot implementation
4. Validation

Tasks
- Review the spec and mark unresolved gaps.
- Define GitHub auth and token storage.
- Define fetch cadence and dedupe behavior.
- Design Telegram summary format.
- Implement fetch, summarize, and send loop.
- Test with real notification samples.

Risks
- Notification volume may create noisy summaries.

Next recommended action
Audit the spec first and lock auth + delivery cadence before implementation.
```

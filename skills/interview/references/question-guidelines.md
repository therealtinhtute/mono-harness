# Interview Question Guidelines

## Question Quality Criteria
- Non-obvious (avoid "did you consider X" for obvious X)
- Specific to the plan (not generic best practices)
- Explores tradeoffs (not yes/no)
- Uncovers hidden assumptions
- Validates feasibility

## Question Categories

### 1. Technical Implementation (30%)
- How will X be implemented given constraint Y?
- What happens when Z fails?
- How does this integrate with existing system A?

### 2. UI/UX & User Flows (20%)
- What does the user see when X happens?
- How does the user recover from error Y?
- What's the happy path vs edge cases?

### 3. Architecture & Design (20%)
- Why pattern X over pattern Y?
- How does this scale to N users/requests?
- What's the data flow from A to B?

### 4. Concerns & Risks (15%)
- What breaks if assumption X is wrong?
- What's the rollback strategy?
- What's the blast radius of failure?

### 5. Tradeoffs & Alternatives (15%)
- Why this approach over alternative X?
- What are we sacrificing for benefit Y?
- What's the maintenance cost?

## Interview Flow
1. Start with high-level questions (goal, approach, scope)
2. Drill into technical details
3. Explore edge cases and failure modes
4. Validate tradeoffs and alternatives
5. Confirm final understanding

## Iteration Rules
- Continue until no ambiguities remain
- Track answered vs unanswered questions
- Adjust questions based on previous answers
- Flag contradictions or conflicts

## Completeness Checklist
Before writing spec, confirm:
- [ ] Goal clearly defined
- [ ] Approach validated with rationale
- [ ] Scope boundaries explicit
- [ ] Key decisions documented with tradeoffs
- [ ] Edge cases identified
- [ ] Failure modes addressed
- [ ] Testing strategy defined
- [ ] Rollback plan exists

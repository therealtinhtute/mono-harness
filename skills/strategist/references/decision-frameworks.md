# Decision Frameworks

## Pros/Cons Evaluation

For each option:
| Criterion | Option A | Option B | Option C |
|-----------|----------|----------|----------|
| Complexity | Low | Medium | High |
| Effort | 2h | 4h | 8h |
| Risk | Low | Medium | High |
| Maintainability | High | Medium | Low |

---

## Trade-off Analysis

Key trade-offs to consider:
- **Speed vs Quality**: Can we iterate later?
- **Simplicity vs Features**: What's MVP?
- **Short-term vs Long-term**: Technical debt implications?

---

## Risk Assessment

Questions to ask:
1. What could go wrong?
2. How likely is failure?
3. What's the impact?
4. How can we mitigate?

---

## Effort Estimation

| Size | Effort | Examples |
|------|--------|----------|
| Small | <2h | Bug fix, config change |
| Medium | 2-8h | New component, refactor |
| Large | 8-24h | New feature, architecture |
| XL | >24h | Major rewrite, migration |

---

## YAGNI Checklist

Before adding complexity:
- [ ] Is this required now or "just in case"?
- [ ] Will this be used in the first release?
- [ ] Is there a simpler approach?
- [ ] Can we add this later if needed?

---

## KISS Checklist

Before implementing:
- [ ] Can a junior dev understand this?
- [ ] Are there fewer moving parts possible?
- [ ] Could this be a config instead of code?
- [ ] Is abstraction premature?

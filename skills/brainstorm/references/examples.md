# Brainstorm Examples

## Example 1: Technology Selection for New Project
**Scenario**: User needs to choose between React and Vue for a new dashboard application.

**Input**:
```bash
"We're building an admin dashboard. Should we use React or Vue?"
```

**Output**:
```markdown
# Brainstorm: React vs Vue for Admin Dashboard

## Problem Statement
Choose frontend framework for new admin dashboard with data tables, charts, forms.

## Evaluated Approaches

### Option 1: React
**Pros:**
- Larger ecosystem, more third-party libraries
- Better TypeScript support out of the box
- More job market availability for hiring
- Stronger enterprise adoption

**Cons:**
- More boilerplate for state management
- Steeper learning curve for new devs
- More decisions to make (routing, state, etc.)

### Option 2: Vue
**Pros:**
- Simpler learning curve, faster onboarding
- Built-in routing and state management
- Better documentation and official tooling
- Less boilerplate, faster initial development

**Cons:**
- Smaller ecosystem compared to React
- Fewer enterprise-grade component libraries
- Less common in job market

## Recommendation: React
**Rationale:** For enterprise dashboard with long-term maintenance, React's ecosystem and hiring pool outweigh Vue's simplicity. The initial complexity pays off in scalability and team flexibility.

## Implementation Considerations
- Use Vite for build tooling
- Consider Tanstack Table for data grids
- Plan for TypeScript from day one
- Budget extra time for architecture decisions

## Success Metrics
- First feature delivered within 2 weeks
- New dev productive within 3 days
- Build time under 30 seconds
```

**Explanation**: The brainstorm explores both options objectively, applies YAGNI (don't over-engineer), and makes a clear recommendation based on long-term maintainability and team considerations.

---

## Example 2: Architecture Decision - Monolith vs Microservices
**Scenario**: User wants to build a new API and is considering microservices architecture.

**Input**:
```bash
"Should we build our new API as microservices or a monolith?"
```

**Output**:
```markdown
# Brainstorm: API Architecture Decision

## Problem Statement
Design architecture for new API serving mobile and web clients. Expected 10K users initially.

## Evaluated Approaches

### Option 1: Microservices
**Pros:**
- Independent scaling of services
- Technology flexibility per service
- Team autonomy and parallel development

**Cons:**
- Massive operational overhead (deployment, monitoring, debugging)
- Network latency between services
- Distributed transaction complexity
- Requires DevOps expertise
- Overkill for 10K users

### Option 2: Modular Monolith
**Pros:**
- Simple deployment (single process)
- Easy debugging and testing
- No network overhead
- Shared database transactions
- Can extract services later if needed

**Cons:**
- All code deploys together
- Scaling requires scaling entire app
- Potential for tight coupling if not disciplined

## Recommendation: Modular Monolith
**Rationale:** **YAGNI applies here.** You don't need microservices complexity for 10K users. Start with a well-structured monolith with clear module boundaries. Extract services only when you have real evidence of scaling needs (100K+ users, specific bottlenecks).

## Implementation Considerations
- Use clear module boundaries (auth, users, billing, etc.)
- Keep modules loosely coupled via interfaces
- Use feature flags for gradual rollouts
- Monitor performance to identify future extraction candidates

## Risk Assessment
- **Risk:** Monolith becomes a "big ball of mud"
- **Mitigation:** Enforce module boundaries via linting, code review
- **Risk:** Premature optimization to microservices
- **Mitigation:** Wait for real data showing bottlenecks

## Success Metrics
- Deploy time under 5 minutes
- 99.9% uptime
- API response time under 200ms p95
- Can handle 50K users before re-architecture
```

**Explanation**: This demonstrates brutal honesty about microservices hype. The brainstorm challenges the assumption that microservices are always better, applies YAGNI and KISS principles, and provides a pragmatic path forward with clear migration criteria.

---

## Example 3: Refactoring Approach for Legacy System
**Scenario**: User has a legacy authentication system that needs modernization but can't afford downtime.

**Input**:
```bash
"Our auth system is a mess. How should we refactor it?"
```

**Output**:
```markdown
# Brainstorm: Auth System Refactoring Strategy

## Problem Statement
Legacy auth system with mixed session/JWT, no tests, tightly coupled to user model. Need to modernize without breaking production.

## Evaluated Approaches

### Option 1: Big Bang Rewrite
**Pros:**
- Clean slate, modern architecture
- No legacy code baggage

**Cons:**
- High risk of breaking production
- Months of parallel development
- Difficult to test all edge cases
- Users experience disruption

**Verdict:** ❌ Too risky. Violates KISS principle.

### Option 2: Strangler Fig Pattern (Incremental)
**Pros:**
- Zero downtime, gradual migration
- Can test each piece in production
- Easy rollback if issues arise
- Maintains business continuity

**Cons:**
- Longer timeline (3-6 months)
- Temporary complexity with two systems
- Requires careful routing logic

**Verdict:** ✅ Recommended approach.

## Recommended Solution: Strangler Fig Pattern

### Phase 1: Add Abstraction Layer (Week 1-2)
- Create `AuthService` interface
- Route all auth calls through it
- Current implementation stays unchanged
- Add comprehensive logging

### Phase 2: Implement New Auth (Week 3-4)
- Build new JWT-based auth service
- Add full test coverage
- Deploy but don't route traffic yet

### Phase 3: Gradual Migration (Week 5-8)
- Route 1% of traffic to new service
- Monitor errors, performance
- Increase to 10%, 50%, 100%
- Keep old system as fallback

### Phase 4: Cleanup (Week 9-10)
- Remove old auth code
- Simplify routing layer
- Update documentation

## Implementation Considerations
- Use feature flags for traffic routing
- Maintain session compatibility during transition
- Plan for rollback at each phase
- Monitor auth success rates closely

## Risk Assessment
- **Risk:** Token incompatibility between systems
- **Mitigation:** Support both formats during transition
- **Risk:** Performance degradation
- **Mitigation:** Load test new system before migration

## Success Metrics
- Zero auth-related outages during migration
- Auth latency stays under 100ms
- 100% test coverage on new system
- Complete migration in 10 weeks
```

**Explanation**: This shows how to handle a common real-world scenario where you can't just "rewrite everything." The brainstorm evaluates the risky approach, rejects it, and provides a detailed incremental strategy that applies KISS and minimizes risk. The phased approach with clear metrics demonstrates practical engineering judgment.

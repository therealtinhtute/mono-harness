---
category: reference
type: checklist
tags: [verify, quality-gates]
---

# Verification Checklist

## Pre-Verification

- [ ] Identify scope (what to verify)
- [ ] Find active plan
- [ ] List changed files
- [ ] Check for breaking changes

## Code Quality Gates

### Tests
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] New code has tests
- [ ] Edge cases covered
- [ ] Test coverage acceptable

### Type Safety
- [ ] TypeScript: `tsc --noEmit` passes
- [ ] Python: `mypy` passes
- [ ] Go: `go vet` passes
- [ ] Rust: `cargo check` passes
- [ ] Java: Compilation succeeds

### Linting
- [ ] ESLint / TSLint passes
- [ ] Prettier formatting correct
- [ ] No console.logs left
- [ ] No TODOs without tickets
- [ ] Import order correct

### Build
- [ ] Production build succeeds
- [ ] No bundle size regressions
- [ ] Assets compile correctly

## Plan Alignment

- [ ] All requirements implemented
- [ ] Implementation matches design
- [ ] No scope creep
- [ ] Architecture follows plan
- [ ] API contracts honored

## Documentation

- [ ] CHANGELOG.md updated
- [ ] README.md updated (if needed)
- [ ] API docs updated
- [ ] Code comments added
- [ ] Examples updated

## Security

- [ ] No secrets in code
- [ ] Input validation present
- [ ] SQL injection prevented
- [ ] XSS prevention in place
- [ ] Auth checks present

## Performance

- [ ] No N+1 queries
- [ ] No unnecessary re-renders
- [ ] No memory leaks
- [ ] Bundle size acceptable
- [ ] Lazy loading used where appropriate

## Gap Analysis

### Must Have
- [ ] Feature works as specified
- [ ] Tests exist and pass
- [ ] Types are correct
- [ ] Error handling present

### Should Have
- [ ] Edge cases handled
- [ ] Performance acceptable
- [ ] Documentation complete
- [ ] Monitoring/logging added

### Nice to Have
- [ ] Metrics/analytics
- [ ] Feature flags
- [ ] A/B test setup
- [ ] Migration guide

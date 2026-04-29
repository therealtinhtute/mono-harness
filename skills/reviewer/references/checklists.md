# Review Checklists

## Security Checklist

- [ ] No hardcoded secrets/credentials
- [ ] Input validation on all user data
- [ ] Proper authentication checks
- [ ] Authorization for protected resources
- [ ] No SQL/command injection risks
- [ ] XSS prevention in rendering
- [ ] CSRF protection on forms
- [ ] Sensitive data encrypted

---

## Performance Checklist

- [ ] No loops inside loops (O(n²))
- [ ] No N+1 database queries
- [ ] Proper caching where needed
- [ ] Lazy loading for large data
- [ ] No memory leaks
- [ ] Optimized bundle size
- [ ] Debounced event handlers

---

## Architecture Checklist

- [ ] Single responsibility principle
- [ ] No duplicate code (DRY)
- [ ] No premature optimization (YAGNI)
- [ ] Simple solution preferred (KISS)
- [ ] Proper error boundaries
- [ ] Consistent patterns

---

## Testing Checklist

- [ ] Unit tests for new logic
- [ ] Edge cases covered
- [ ] Error scenarios tested
- [ ] No flaky test patterns

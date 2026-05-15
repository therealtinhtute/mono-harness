# Review Dimensions — Detailed Checklists

## 1. Security (always first)

### Input & Validation
- SQL / command / path injection at every entry point
- XSS vectors in rendered output
- Missing or bypassable input validation
- File upload without type / size constraints

### Auth & Access
- Missing authentication on protected routes
- Authorization bypass (horizontal + vertical privilege)
- Insecure direct object references (IDOR)
- Session token exposure in logs or responses

### Data Exposure
- PII or credentials in logs, errors, or API responses
- Hardcoded secrets, tokens, or keys in code
- Overly permissive CORS or CSP headers
- Sensitive data in URLs (query params, path segments)

---

## 2. Performance

- N+1 query patterns in loops
- Missing database indexes on filtered / sorted columns
- Unbounded queries without pagination or limits
- Memory leaks: event listeners not cleaned up, growing caches
- Blocking I/O in hot paths or request handlers
- Unnecessary recomputation (results not cached)

---

## 3. Architecture

- YAGNI — is this abstraction actually needed now?
- KISS — can this be simpler without losing correctness?
- DRY — is logic duplicated where a single source of truth exists?
- API contract correctness — does the interface match callers?
- Backward-compat — does this break existing consumers?
- Separation of concerns — is business logic mixed with I/O?
- Harness alignment — does the implementation still match the locked spec, phase context, and phase boundaries?

---

## 4. Code Quality

- Naming clarity — does the name say what it does without a comment?
- Error handling at system boundaries (external APIs, DB, file I/O)
- Type safety — are nulls and undefined cases handled?
- Test coverage — does new behavior have a test?
- Dead code — unreachable branches, unused imports, stale comments
- Proof trail quality — can the claimed verification be traced to actual commands or run artifacts?

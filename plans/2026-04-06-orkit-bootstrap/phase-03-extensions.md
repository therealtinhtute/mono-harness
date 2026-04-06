---
title: Phase 3 - Initial Extensions
description: Create 5-10 production-grade extensions as reference implementations
status: draft
created: 2026-04-06
phase: 3
effort: large
---

# Phase 3: Initial Extensions

## Objectives

Create 5-10 production-quality extensions demonstrating best practices:
- Code review agent
- Test execution agent
- Debugging agent
- Documentation agent
- Planning agent
- Git automation hooks
- Enhanced statusline
- API design skill
- Error handling skill
- Security audit skill

## Extension Specifications

### 3.1 Code Reviewer Agent

**Location**: `plugins/code-reviewer/`

**Purpose**: Expert code review covering security, performance, architecture, and maintainability

**Files**:
- `code-reviewer.md`: Agent definition
- `README.md`: Documentation
- `references/review-checklist.md`: Review criteria

**Agent Configuration**:
```yaml
---
name: code-reviewer
description: Expert code review agent for security, performance, and quality
model: sonnet
permissionMode: acceptEdits
maxTurns: 15
tools: [Read, Grep, Glob, Bash]
disallowedTools: [Write, Edit]
effort: medium
memory: project
---

You are an expert code reviewer specializing in security, performance, architecture, and maintainability.

## Review Focus Areas

### Security
- Input validation and sanitization
- Authentication and authorization
- Secret management
- SQL injection and XSS vulnerabilities
- Dependency vulnerabilities

### Performance
- Algorithm complexity
- Database query optimization
- Memory leaks
- Caching opportunities
- Resource management

### Architecture
- SOLID principles
- Design patterns
- Separation of concerns
- Code organization
- Dependency management

### Maintainability
- Code readability
- Documentation quality
- Test coverage
- Error handling
- Naming conventions

## Review Process

1. Read changed files using Read tool
2. Analyze code against review criteria
3. Identify issues with severity (critical, high, medium, low)
4. Provide specific recommendations with code examples
5. Highlight positive patterns worth keeping
6. Generate summary report

## Output Format

### Issue Template
**Severity**: [Critical|High|Medium|Low]
**Category**: [Security|Performance|Architecture|Maintainability]
**Location**: file:line
**Issue**: Brief description
**Recommendation**: Specific fix with code example
**Rationale**: Why this matters

### Summary
- Total issues by severity
- Key recommendations
- Overall assessment
```

### 3.2 Tester Agent

**Location**: `plugins/tester/`

**Purpose**: Run tests, analyze failures, and provide actionable feedback

**Files**:
- `tester.md`: Agent definition
- `README.md`: Documentation
- `scripts/detect-test-framework.sh`: Auto-detect test framework

**Agent Configuration**:
```yaml
---
name: tester
description: Test execution and analysis agent
model: sonnet
permissionMode: acceptEdits
maxTurns: 20
tools: [Read, Grep, Glob, Bash]
effort: medium
memory: project
---

You are a testing specialist who runs tests, analyzes failures, and provides clear guidance.

## Responsibilities

1. Detect test framework (Jest, Vitest, pytest, Go test, etc.)
2. Run appropriate test commands
3. Analyze test output and failures
4. Identify root causes
5. Provide specific fix recommendations
6. Verify fixes by re-running tests

## Test Framework Detection

Use `scripts/detect-test-framework.sh` to identify:
- JavaScript: Jest, Vitest, Mocha, Jasmine
- Python: pytest, unittest
- Go: go test
- Rust: cargo test
- Ruby: RSpec, Minitest

## Failure Analysis

For each failing test:
1. Extract error message and stack trace
2. Identify failing assertion
3. Determine root cause (logic error, missing mock, etc.)
4. Suggest specific fix
5. Check for related failures

## Output Format

### Test Summary
- Total tests: X
- Passed: X
- Failed: X
- Skipped: X
- Duration: Xs

### Failure Details
**Test**: test name
**File**: path:line
**Error**: error message
**Root Cause**: analysis
**Fix**: specific recommendation
**Related**: other affected tests
```

### 3.3 Debugger Agent

**Location**: `plugins/debugger/`

**Purpose**: Systematic debugging for bugs, build errors, and regressions

**Agent Configuration**:
```yaml
---
name: debugger
description: Systematic debugging agent for root cause analysis
model: sonnet
permissionMode: acceptEdits
maxTurns: 25
tools: [Read, Grep, Glob, Bash]
effort: high
memory: project
---

You are a debugging specialist who systematically identifies root causes.

## Debugging Process

### 1. Reproduce
- Understand the issue
- Identify reproduction steps
- Verify the bug exists

### 2. Isolate
- Binary search to narrow scope
- Identify last working version
- Find minimal reproduction case

### 3. Analyze
- Read relevant code
- Check logs and error messages
- Trace execution flow
- Identify root cause

### 4. Fix
- Propose specific fix
- Explain why it works
- Consider edge cases
- Verify fix resolves issue

### 5. Prevent
- Suggest tests to prevent regression
- Identify similar issues
- Recommend architectural improvements

## Debugging Techniques

- Binary search (git bisect)
- Logging and tracing
- Hypothesis testing
- Rubber duck debugging
- Stack trace analysis
- Memory profiling
- Performance profiling

## Output Format

### Bug Report
**Issue**: Description
**Reproduction**: Steps to reproduce
**Root Cause**: Technical explanation
**Fix**: Specific code changes
**Tests**: Regression tests needed
**Prevention**: How to avoid similar issues
```

### 3.4 Documentation Manager Agent

**Location**: `plugins/docs-manager/`

**Purpose**: Maintain documentation, changelogs, and release notes

**Agent Configuration**:
```yaml
---
name: docs-manager
description: Documentation maintenance and generation agent
model: sonnet
permissionMode: acceptEdits
maxTurns: 15
tools: [Read, Grep, Glob, Write, Edit]
effort: medium
memory: project
---

You are a documentation specialist who maintains high-quality docs.

## Responsibilities

1. Update documentation after code changes
2. Generate changelog entries
3. Write release notes
4. Maintain API documentation
5. Ensure docs accuracy and completeness

## Documentation Types

### Code Documentation
- Inline comments for complex logic
- Function/method documentation
- Class/module documentation
- README files

### User Documentation
- Getting started guides
- Tutorials and examples
- API reference
- Configuration guides

### Project Documentation
- Architecture decisions
- Development guides
- Contribution guidelines
- Changelog and release notes

## Update Process

1. Detect changed files
2. Identify documentation impact
3. Update affected docs
4. Generate changelog entry
5. Verify links and examples
6. Check for completeness

## Output Format

### Documentation Update
**Files Changed**: list
**Docs Impact**: [none|minor|major]
**Updates Made**: list of doc changes
**Changelog Entry**: formatted entry
```

### 3.5 Planner Agent

**Location**: `plugins/planner/`

**Purpose**: Create implementation plans with phased approach

**Agent Configuration**:
```yaml
---
name: planner
description: Implementation planning agent
model: sonnet
permissionMode: acceptEdits
maxTurns: 20
tools: [Read, Grep, Glob, Write]
effort: high
memory: project
---

You are a technical planner who creates scalable, maintainable implementation plans.

## Planning Principles

- YAGNI (You Aren't Gonna Need It)
- KISS (Keep It Simple, Stupid)
- DRY (Don't Repeat Yourself)
- Progressive disclosure
- Phased implementation

## Planning Process

1. Understand requirements
2. Research existing solutions
3. Analyze trade-offs
4. Design architecture
5. Break into phases
6. Define success criteria
7. Identify risks

## Plan Structure

### Overview (plan.md)
- Project summary
- Architecture decision
- Implementation phases
- Success criteria
- Dependencies
- Risks

### Phase Details (phase-XX.md)
- Objectives
- Tasks with acceptance criteria
- Dependencies
- Estimated effort
- Next phase

### Research (research/)
- Technical research
- Best practices
- Similar implementations
- Trade-off analysis

## Output Format

Plans saved to `.kit/plans/{date}-{slug}/`
```

### 3.6 Git Hooks

**Location**: `plugins/git-hooks/`

**Purpose**: Pre-commit and pre-push automation

**Files**:
- `hooks/hooks.json`: Hook definitions
- `scripts/pre-commit.sh`: Pre-commit checks
- `scripts/pre-push.sh`: Pre-push checks
- `README.md`: Documentation

**Hook Configuration**:
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "npx prettier --write",
            "timeout": 60,
            "async": false
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/scripts/pre-commit.sh",
            "timeout": 30,
            "async": false
          }
        ]
      }
    ]
  }
}
```

**Pre-commit Script**:
```bash
#!/usr/bin/env bash
set -euo pipefail

# Run linting
if command -v eslint &>/dev/null; then
  eslint --fix .
fi

# Run formatting
if command -v prettier &>/dev/null; then
  prettier --write .
fi

# Check for secrets
if command -v gitleaks &>/dev/null; then
  gitleaks protect --staged
fi

exit 0
```

### 3.7 Enhanced Statusline

**Location**: `plugins/statusline-pro/`

**Purpose**: Rich statusline with git, cost, and context info

**Files**:
- `statusline.sh`: Statusline script
- `README.md`: Documentation

**Statusline Script**:
```bash
#!/usr/bin/env bash
set -euo pipefail

# Read JSON input
input=$(cat)

# Extract fields
model=$(echo "$input" | jq -r '.model.display_name // "Unknown"')
cost=$(echo "$input" | jq -r '.cost.total_cost_usd // 0')
context_pct=$(echo "$input" | jq -r '.context_window.used_percentage // 0')
cwd=$(echo "$input" | jq -r '.workspace.current_dir // ""')

# Git info
git_branch=""
git_status=""
if [[ -d "${cwd}/.git" ]]; then
  git_branch=$(cd "$cwd" && git branch --show-current 2>/dev/null || echo "")
  if [[ -n "$git_branch" ]]; then
    git_status=$(cd "$cwd" && git status --porcelain 2>/dev/null | wc -l | xargs)
  fi
fi

# Color codes
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

# Build statusline
statusline=""

# Model
statusline+="${BLUE}${model}${RESET}"

# Cost
if (( $(echo "$cost > 0" | bc -l) )); then
  statusline+=" ${GREEN}\$${cost}${RESET}"
fi

# Context
if (( $(echo "$context_pct > 75" | bc -l) )); then
  statusline+=" ${RED}${context_pct}%${RESET}"
elif (( $(echo "$context_pct > 50" | bc -l) )); then
  statusline+=" ${YELLOW}${context_pct}%${RESET}"
else
  statusline+=" ${GREEN}${context_pct}%${RESET}"
fi

# Git
if [[ -n "$git_branch" ]]; then
  statusline+=" ${BLUE}${git_branch}${RESET}"
  if [[ "$git_status" != "0" ]]; then
    statusline+=" ${YELLOW}+${git_status}${RESET}"
  fi
fi

echo -e "$statusline"
```

### 3.8 API Conventions Skill

**Location**: `plugins/api-conventions/`

**Purpose**: API design patterns and best practices

**Skill Definition**:
```yaml
---
name: api-conventions
description: API design patterns and REST best practices
user-invocable: true
model: sonnet
effort: low
---

# API Conventions

Apply these conventions when designing or reviewing APIs.

## REST Principles

### Resource Naming
- Use plural nouns: `/users`, `/posts`
- Hierarchical: `/users/{id}/posts`
- Lowercase with hyphens: `/user-profiles`

### HTTP Methods
- GET: Retrieve resource(s)
- POST: Create resource
- PUT: Replace resource
- PATCH: Update resource
- DELETE: Remove resource

### Status Codes
- 200: Success
- 201: Created
- 204: No content
- 400: Bad request
- 401: Unauthorized
- 403: Forbidden
- 404: Not found
- 500: Server error

## Request/Response Format

### Request
```json
{
  "data": {
    "type": "users",
    "attributes": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

### Response
```json
{
  "data": {
    "id": "123",
    "type": "users",
    "attributes": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  },
  "meta": {
    "timestamp": "2026-04-06T15:00:00Z"
  }
}
```

## Best Practices

- Version APIs: `/v1/users`
- Use pagination: `?page=1&limit=20`
- Support filtering: `?status=active`
- Include timestamps
- Provide clear error messages
- Document with OpenAPI/Swagger
```

### 3.9 Error Handling Skill

**Location**: `plugins/error-handling/`

**Purpose**: Error handling patterns across languages

**Skill Definition**:
```yaml
---
name: error-handling
description: Error handling patterns and best practices
user-invocable: true
model: sonnet
effort: low
---

# Error Handling

Comprehensive error handling patterns.

## Principles

1. Fail fast and explicitly
2. Provide context in errors
3. Log errors appropriately
4. Handle errors at the right level
5. Don't swallow errors

## Patterns by Language

### JavaScript/TypeScript
```typescript
try {
  const result = await riskyOperation();
  return result;
} catch (error) {
  logger.error('Operation failed', { error, context });
  throw new AppError('Failed to complete operation', { cause: error });
}
```

### Python
```python
try:
    result = risky_operation()
    return result
except SpecificError as e:
    logger.error(f"Operation failed: {e}", exc_info=True)
    raise AppError("Failed to complete operation") from e
```

### Go
```go
result, err := riskyOperation()
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
return result, nil
```

### Rust
```rust
let result = risky_operation()
    .map_err(|e| AppError::OperationFailed(e))?;
Ok(result)
```

## Error Types

### Validation Errors
- User input errors
- Return 400 Bad Request
- Include field-level details

### Authorization Errors
- Permission denied
- Return 401/403
- Log security events

### Not Found Errors
- Resource doesn't exist
- Return 404
- Don't leak information

### Server Errors
- Unexpected failures
- Return 500
- Log full context
- Alert on-call

## Best Practices

- Use custom error types
- Include error codes
- Provide actionable messages
- Log with context
- Monitor error rates
```

### 3.10 Security Scanner Skill

**Location**: `plugins/security-scanner/`

**Purpose**: Security audit and vulnerability detection

**Skill Definition**:
```yaml
---
name: security-scanner
description: Security audit and vulnerability detection
user-invocable: true
model: sonnet
effort: medium
allowed-tools: [Read, Grep, Glob, Bash]
---

# Security Scanner

Comprehensive security audit for code and configurations.

## Scan Categories

### 1. Secret Detection
- API keys and tokens
- Passwords and credentials
- Private keys
- Database connection strings

### 2. Injection Vulnerabilities
- SQL injection
- Command injection
- XSS (Cross-Site Scripting)
- Path traversal

### 3. Authentication & Authorization
- Weak password policies
- Missing authentication
- Broken access control
- Session management issues

### 4. Dependency Vulnerabilities
- Outdated packages
- Known CVEs
- Malicious dependencies

### 5. Configuration Issues
- Debug mode in production
- Exposed admin panels
- Insecure defaults
- Missing security headers

## Scan Process

1. Run automated scanners (gitleaks, semgrep, etc.)
2. Manual code review for patterns
3. Check dependencies for vulnerabilities
4. Review configuration files
5. Generate prioritized report

## Output Format

### Vulnerability Report
**Severity**: [Critical|High|Medium|Low]
**Category**: [Secrets|Injection|Auth|Dependencies|Config]
**Location**: file:line
**Issue**: Description
**Impact**: Potential consequences
**Fix**: Remediation steps
**References**: CVE/CWE links
```

## Implementation Tasks

### 3.1 Create Extension Structure
- [ ] Create plugin directories
- [ ] Add README files
- [ ] Add LICENSE files

### 3.2 Implement Agents
- [ ] Code reviewer agent
- [ ] Tester agent
- [ ] Debugger agent
- [ ] Docs manager agent
- [ ] Planner agent

### 3.3 Implement Hooks
- [ ] Git hooks configuration
- [ ] Pre-commit script
- [ ] Pre-push script

### 3.4 Implement Skills
- [ ] API conventions skill
- [ ] Error handling skill
- [ ] Security scanner skill

### 3.5 Implement Statusline
- [ ] Enhanced statusline script
- [ ] Configuration

### 3.6 Testing
- [ ] Test each extension locally
- [ ] Validate with orkit CLI
- [ ] Security scan all extensions
- [ ] Document usage examples

### 3.7 Update Marketplace
- [ ] Add extensions to marketplace.json
- [ ] Update main README
- [ ] Create installation guide

## Acceptance Criteria

- [ ] All 10 extensions implemented
- [ ] Each extension passes validation
- [ ] README and documentation complete
- [ ] Local testing successful
- [ ] Added to marketplace.json
- [ ] Usage examples provided

## Dependencies

- Phase 1: Foundation (schemas, CLI)
- Phase 2: Validation pipeline

## Estimated Effort

7-10 days

## Next Phase

Phase 4: CLI Tooling Enhancement

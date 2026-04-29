# Verifier Skill Scripts

Automation scripts for quality verification workflows.

## Scripts

### quality-check.sh

Run all quality checks: tests, types, lint, build.

**Usage:**
```bash
./quality-check.sh
./quality-check.sh --fast
./quality-check.sh --parallel
```

**Features:**
- Runs tests with coverage
- Type checking (TypeScript/Python)
- Linting (ESLint/Biome)
- Build verification
- Parallel execution support
- Summary report

**Options:**
- `-h, --help`: Show help
- `-f, --fast`: Skip slow checks (build)
- `-p, --parallel`: Run checks in parallel

**Exit codes:**
- `0`: All checks passed
- `1`: One or more checks failed

---

### plan-alignment.sh

Verify implementation matches plan.

**Usage:**
```bash
./plan-alignment.sh .kit/plans/20260416-week3-quality/plan.md
./plan-alignment.sh plan.md --output alignment-report.md
```

**Features:**
- Extracts tasks from plan file
- Checks file existence
- Identifies gaps
- Progress tracking
- Generates alignment report

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path

---

### pre-commit-check.sh

Fast pre-commit validation for staged files.

**Usage:**
```bash
./pre-commit-check.sh
```

**Features:**
- Syntax validation (JS/TS/Python/Bash)
- Secret scanning
- Format checking (Prettier)
- Fast execution (staged files only)
- Git hook compatible

**Options:**
- `-h, --help`: Show help

**Exit codes:**
- `0`: All checks passed
- `1`: Checks failed (blocks commit)

**Git hook integration:**
```bash
# .git/hooks/pre-commit
#!/bin/bash
kit/skills/verifier/scripts/pre-commit-check.sh
```

---

### release-ready.sh

Comprehensive release readiness validation.

**Usage:**
```bash
./release-ready.sh v1.2.0
./release-ready.sh v2.0.0 --output release-report.md
```

**Features:**
- All quality checks
- Changelog validation
- Version consistency check
- Documentation check
- Blocker identification
- Go/no-go verdict

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path

**Exit codes:**
- `0`: Ready for release
- `1`: Blockers found

---

## Installation

All scripts are executable and can be run directly:

```bash
cd kit/skills/verifier/scripts
./quality-check.sh --help
```

## Output Location

Reports are saved to `.kit/reports/verify/` by default:
- `YYYYMMDD-release.md` - Release readiness reports

## Integration

### Pre-commit Hook

Install as git hook:

```bash
#!/bin/bash
# .git/hooks/pre-commit
kit/skills/verifier/scripts/pre-commit-check.sh || exit 1
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
- name: Quality Check
  run: kit/skills/verifier/scripts/quality-check.sh

- name: Release Check
  run: kit/skills/verifier/scripts/release-ready.sh ${{ github.ref_name }}
```

## Testing

Run tests with:

```bash
bats tests/
```

from therealTINHTUTE with love

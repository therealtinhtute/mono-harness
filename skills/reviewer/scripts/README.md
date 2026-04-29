# Reviewer Skill Scripts

Automation scripts for code review workflows.

## Scripts

### review-report.sh

Generate structured code review report with automated analysis.

**Usage:**
```bash
./review-report.sh
./review-report.sh HEAD~5..HEAD
./review-report.sh main..feature-branch
./review-report.sh --output custom-review.md
```

**Features:**
- Analyzes git diff for changes
- Security vulnerability detection
- Performance issue detection
- Architecture analysis
- Code quality checks
- Categorizes issues by severity
- Generates markdown report

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path
- `-f, --format`: Output format (markdown|json)

---

### security-scan.sh

Automated security vulnerability scanner.

**Usage:**
```bash
./security-scan.sh
./security-scan.sh src/**/*.js
./security-scan.sh --format json --output scan.json
```

**Features:**
- SQL injection detection
- XSS vulnerability detection
- Auth boundary violation detection
- Hardcoded secret detection
- Line-by-line reporting
- JSON output support

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path
- `-f, --format`: Output format (text|json)

**Exit codes:**
- `0`: No issues found
- `1`: Security issues detected

---

### review-checklist.sh

Interactive code review checklist with guided workflow.

**Usage:**
```bash
./review-checklist.sh
./review-checklist.sh --output custom-checklist.md
```

**Features:**
- Step-by-step interactive checklist
- Security checks
- Performance checks
- Architecture checks
- Code quality checks
- Automatic verdict generation
- Saves results to markdown

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path

**Checklist categories:**
1. Security (7 checks)
2. Performance (5 checks)
3. Architecture (6 checks)
4. Code Quality (5 checks)

---

## Installation

All scripts are executable and can be run directly:

```bash
cd kit/skills/reviewer/scripts
./review-report.sh --help
```

## Output Location

Reports are saved to `.kit/reports/review/` by default:
- `YYYYMMDD-review.md` - Automated review reports
- `YYYYMMDD-checklist.md` - Interactive checklist results

## Integration

Use these scripts in your workflow:

```bash
# Before creating PR
./review-report.sh main..HEAD

# Security scan on staged files
./security-scan.sh

# Interactive review
./review-checklist.sh
```

## Testing

Run tests with:

```bash
bats tests/
```

from therealTINHTUTE with love

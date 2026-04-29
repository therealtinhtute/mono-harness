# Strategist Skill Scripts

Automation scripts for strategic planning and decision-making workflows.

## Scripts

### decision-matrix.sh

Generate decision comparison matrix with weighted scoring.

**Usage:**
```bash
./decision-matrix.sh "REST API" "GraphQL"
./decision-matrix.sh "Monolith" "Microservices" "Serverless"
```

**Features:**
- Interactive criteria input
- Weighted scoring (1-10 scale)
- Trade-off visualization
- Automatic recommendation
- Markdown output

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path

**Workflow:**
1. Enter evaluation criteria
2. Assign weights to each criterion
3. Score each option
4. Get weighted recommendation

---

### plan-generator.sh

Generate structured implementation plan template.

**Usage:**
```bash
./plan-generator.sh "User Authentication"
./plan-generator.sh "Payment Integration" --dir .kit/plans/custom/
```

**Features:**
- Creates plan directory structure
- Generates plan.md template
- Includes standard sections
- Pre-filled frontmatter
- Ready to edit

**Options:**
- `-h, --help`: Show help
- `-d, --dir`: Custom plan directory

**Generated structure:**
```
.kit/plans/YYYYMMDD-slug/
└── plan.md
```

---

### yagni-check.sh

Validate YAGNI/KISS/DRY principles in code.

**Usage:**
```bash
./yagni-check.sh
./yagni-check.sh src/**/*.js
```

**Features:**
- YAGNI: Detects unused abstractions, premature optimization
- KISS: Finds complex conditionals, deep nesting
- DRY: Identifies code duplication
- File-by-file reporting

**Options:**
- `-h, --help`: Show help

**Exit codes:**
- `0`: No violations or warnings only

---

## Installation

All scripts are executable and can be run directly:

```bash
cd kit/skills/strategist/scripts
./decision-matrix.sh --help
```

## Output Location

Reports are saved to `.kit/reports/brainstorm/` and `.kit/plans/` by default:
- `YYYYMMDD-decision.md` - Decision matrices
- `YYYYMMDD-slug/plan.md` - Implementation plans

## Integration

Use these scripts in your workflow:

```bash
# Compare architectural options
./decision-matrix.sh "Option A" "Option B" "Option C"

# Generate new plan
./plan-generator.sh "New Feature"

# Check code quality
./yagni-check.sh src/
```

## Testing

Run tests with:

```bash
bats tests/
```

from therealTINHTUTE with love

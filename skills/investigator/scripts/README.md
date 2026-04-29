# Investigator Skill Scripts

Automation scripts for codebase discovery and research workflows.

## Scripts

### smart-search.sh

Intelligent multi-strategy codebase search.

**Usage:**
```bash
./smart-search.sh "authentication"
./smart-search.sh "user login" --limit 5
```

**Features:**
- File name search
- Content search with context
- Symbol search (functions/classes)
- Ranked results
- Context preview

**Options:**
- `-h, --help`: Show help
- `-l, --limit`: Max results per strategy (default: 10)

**Search strategies:**
1. File name matches
2. Content matches with line numbers
3. Symbol definitions (functions/classes)

---

### pattern-library.sh

Pre-defined search patterns for common code elements.

**Usage:**
```bash
./pattern-library.sh api-endpoints
./pattern-library.sh database-queries
./pattern-library.sh env-vars
./pattern-library.sh imports
./pattern-library.sh exports
./pattern-library.sh todos
```

**Features:**
- Pre-defined regex patterns
- Language-specific searches
- Result filtering
- Export to file

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output file path

**Available patterns:**
- `api-endpoints`: Find API route definitions
- `database-queries`: Find database queries
- `env-vars`: Find environment variable usage
- `imports`: Find import statements
- `exports`: Find export statements
- `todos`: Find TODO/FIXME comments

---

### context-gather.sh

Gather comprehensive context for a feature or topic.

**Usage:**
```bash
./context-gather.sh "user authentication"
./context-gather.sh "payment processing" --output .kit/context/custom/
```

**Features:**
- Find related files by name and content
- Extract key functions and classes
- Build dependency graph
- Generate context report
- Save to structured directory

**Options:**
- `-h, --help`: Show help
- `-o, --output`: Output directory path

**Generated structure:**
```
.kit/context/YYYYMMDD-slug/
└── context.md
```

**Report sections:**
1. Related files (by filename and content)
2. Key functions (definitions)
3. Dependencies (imports)
4. Summary and next steps

---

## Installation

All scripts are executable and can be run directly:

```bash
cd kit/skills/investigator/scripts
./smart-search.sh --help
```

## Output Location

Reports are saved to `.kit/context/` by default:
- `YYYYMMDD-slug/context.md` - Context reports

## Integration

Use these scripts in your workflow:

```bash
# Quick search
./smart-search.sh "authentication"

# Find specific patterns
./pattern-library.sh api-endpoints

# Gather full context
./context-gather.sh "user authentication"
```

## Testing

Run tests with:

```bash
bats tests/
```

from therealTINHTUTE with love

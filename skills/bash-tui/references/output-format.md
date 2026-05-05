# Output Format

## For reusable scripts
Save to: `scripts/{script-name}.sh`

## For documentation
Save to: `.kit/reports/bash-tui/{YYYYMMDD}-{slug}.md`

Frontmatter:
```yaml
---
title: Bash TUI - {slug}
description: {one-line summary}
status: active | archived
created: YYYY-MM-DD
tags: [bash-tui, {slug}]
---
```

Include:
- Script purpose and usage
- Dependencies (gum, dialog, pure bash)
- Installation instructions
- Example invocations

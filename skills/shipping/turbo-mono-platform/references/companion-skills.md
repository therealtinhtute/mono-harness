# Companion Skills

## Concept

This skill (turbo-mono-platform) handles the **monorepo layer**.
Other skills handle specialized layers. When a task overlaps with a companion skill,
check if it's installed and invoke it — or suggest the user install it.

---

## shadcn/ui Official Skill

**What it does:** Injects live project context via `shadcn info --json`, enforces
component composition rules (FieldGroup, ToggleGroup, semantic colors), provides
full CLI docs, theming guide, and registry authoring.

**When it's needed:** Any time you're adding/modifying components, building forms,
customizing themes, or working with shadcn CLI.

### Check if installed

```bash
ls ~/.claude/skills/ 2>/dev/null | grep -i "shadcn" \
  || ls .claude/skills/ 2>/dev/null | grep -i "shadcn" \
  || echo "NOT_INSTALLED"
```

### If NOT installed — suggest to user

```
⚠️ The official shadcn/ui skill is not installed.
It provides live component context, composition rules, and CLI docs.

Install it:
  bunx skills add shadcn/ui

Or if you have pnpm:
  pnpm dlx skills add shadcn/ui

After installing, re-run your request — the skill will auto-detect
your packages/ui/components.json and inject the correct config.
```

### If installed — how it helps

The shadcn/ui skill automatically:
- Runs `shadcn info --json` to read your `components.json`
- Knows your framework, aliases, base library (radix vs base), icon library
- Enforces composition rules: `FieldGroup` for forms, `ToggleGroup` for options, etc.
- Provides `shadcn docs <component>` and `shadcn search` for discovery

**Division of responsibility when both skills are active:**

| Task | Handle with |
|------|-------------|
| Add component (`shadcn add`) | shadcn/ui skill |
| Component composition rules | shadcn/ui skill |
| Form patterns (FieldGroup, Field) | shadcn/ui skill |
| Icon usage rules | shadcn/ui skill |
| Theming (OKLCH, CSS vars) | shadcn/ui skill |
| `@source` in globals.css | turbo-mono-platform (this skill) |
| Import from `@rp/ui` | turbo-mono-platform (this skill) |
| `packages/ui` package structure | turbo-mono-platform (this skill) |
| Monorepo turbo config | turbo-mono-platform (this skill) |

---

## Other Recommended Companion Skills

Check and suggest these based on what the project needs:

### turborepo (Vercel official)

Deep turbo.json config knowledge, CI patterns, filtering, caching.

```bash
ls ~/.claude/skills/ 2>/dev/null | grep -i "turborepo" || echo "NOT_INSTALLED"
```

Install: `bunx skills add vercel/turborepo --skill turborepo`
When to suggest: User asks about turbo pipeline, remote cache, `--affected` flag, CI config.

---

## Suggestion Template

Use this format when a skill is missing:

```
💡 This task would benefit from the **[skill-name]** skill.

It provides: [one line of what it does]

Install:
  bunx skills add [source]

After installing, your AI assistant will automatically use it
when working on [topic].

Want me to continue without it? I'll use the reference docs I have.
```

---

## Detection Logic in SKILL.md

When the turbo-mono-platform skill loads, it should check at runtime:

```
!`ls ~/.claude/skills/ .claude/skills/ 2>/dev/null | grep -E "shadcn|turborepo" | sort || echo "no companion skills found"`
```

Use result to:
- If `shadcn` found → defer component work to it
- If not found → use `shadcn-rules.md` as fallback, suggest install
- If `turborepo` found → defer turbo config questions to it

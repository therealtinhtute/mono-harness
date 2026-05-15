# Design System Guide

Present these options during planning. Ask user to describe aesthetic first,
then map to a concrete choice.

---

## Step 1 — Aesthetic Interview

Ask: *"Describe the visual feel you want (pick words that resonate):"*

| Cluster | Words | → Maps to |
|---------|-------|-----------|
| **Minimal / Clean** | Simple, lots of whitespace, focused, calm | Neutral + b1tMcUv91 |
| **Professional / Enterprise** | Trustworthy, structured, data-dense, serious | Slate + default shadcn |
| **Bold / Vibrant** | Energetic, colorful, playful, expressive | Custom OKLCH palette |
| **Dark / Hacker** | Dark-first, high contrast, technical, terminal | Zinc dark + b1tMcUv91 |
| **Warm / Human** | Friendly, approachable, organic, cozy | Stone + custom radius |

---

## Step 2 — shadcn Preset Options

### ✅ Recommended: `b1tMcUv91` (default for this stack)
- Style: new-york
- Base: neutral
- CSS variables: OKLCH
- Radius: 0.5rem (balanced)
- Best for: SaaS, internal tools, marketplaces, content

```bash
bunx --bun shadcn@latest init --preset b1tMcUv91 --template next --monorepo
```

### Option: Default shadcn (no preset)
- Style: default
- Base: slate
- Best for: enterprise / data-heavy dashboards
- Trade-off: less polished out of the box

```bash
bunx --bun shadcn@latest init --template next --monorepo
```

### Option: Custom palette from scratch
- Start with b1tMcUv91 then override `@theme` in `globals.css`
- Best for: strong brand identity

---

## Step 3 — Color Palette by Aesthetic

### Neutral (default — clean SaaS)
```css
@layer base {
  :root {
    --background: oklch(1 0 0);
    --foreground: oklch(0.145 0 0);
    --primary: oklch(0.205 0 0);
    --primary-foreground: oklch(0.985 0 0);
  }
  .dark {
    --background: oklch(0.145 0 0);
    --foreground: oklch(0.985 0 0);
  }
}
```

### Blue / Professional (B2B, enterprise)
```css
@layer base {
  :root {
    --primary: oklch(0.546 0.245 262.881);          /* blue-600 */
    --primary-foreground: oklch(0.985 0 0);
  }
}
```

### Violet / Modern SaaS
```css
@layer base {
  :root {
    --primary: oklch(0.585 0.265 301.9);             /* violet-600 */
    --primary-foreground: oklch(0.985 0 0);
  }
}
```

### Emerald / Growth / Fintech
```css
@layer base {
  :root {
    --primary: oklch(0.609 0.179 155.569);           /* emerald-600 */
    --primary-foreground: oklch(0.985 0 0);
  }
}
```

### Rose / Consumer / Lifestyle
```css
@layer base {
  :root {
    --primary: oklch(0.645 0.246 16.439);            /* rose-600 */
    --primary-foreground: oklch(0.985 0 0);
  }
}
```

---

## Step 4 — Typography

All options use Geist by default (included via Next.js). Override in `@theme`:

| Font pair | Best for | CSS |
|-----------|----------|-----|
| Geist + Geist Mono | SaaS, tools (default) | already set |
| Inter + JetBrains Mono | Developer tools | `--font-sans: "Inter"` |
| Cal Sans + Inter | Marketing, landing | `--font-heading: "Cal Sans"` |
| Instrument Serif + Inter | Editorial, content | mixed serif/sans |

---

## Step 5 — Radius (border-radius feel)

| Value | Feel | CSS |
|-------|------|-----|
| `0rem` | Brutal / sharp | `--radius: 0rem` |
| `0.3rem` | Tight / minimal | `--radius: 0.3rem` |
| `0.5rem` | Balanced (default) | `--radius: 0.5rem` |
| `0.75rem` | Friendly / rounded | `--radius: 0.75rem` |
| `1rem` | Playful / soft | `--radius: 1rem` |

---

## Quick Decision Matrix

| Project type | Recommended preset | Palette | Radius |
|-------------|-------------------|---------|--------|
| B2B SaaS | b1tMcUv91 | Neutral | 0.5rem |
| Developer tool | b1tMcUv91 | Zinc dark | 0.3rem |
| Marketplace | b1tMcUv91 | Violet | 0.5rem |
| E-commerce | b1tMcUv91 | custom brand | 0.75rem |
| Internal tool | default shadcn | Slate | 0.3rem |
| Content / blog | b1tMcUv91 | Stone warm | 0.5rem |

---

## After Design Confirmed

Update `packages/ui/src/styles/globals.css` with chosen palette.
Note preset ID for scaffold command:
```bash
bunx --bun shadcn@latest init --preset b1tMcUv91 --template next --monorepo
```

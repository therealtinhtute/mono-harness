---
id: 01M0SITEOPIPLAN8M2VHZ5E9TB
type: plan
intake_id: 01M0SITEOPIINTK8M2VHZ5E9TB
lane: normal
status: active
created: 2026-08-28
updated: 2026-08-28
---

# Plan: site-openpi-redesign — OpenPI Blueprint Design & v0.15 Doc Sync

## Outcome

- result: the static documentation website under `site/` is completely transformed to adhere to the OpenPI Blueprint Design System defined in `DESIGN.md` (warm cream canvas `#f6f5f0` with 30px grid overlay, Space Grotesk + Instrument Serif + Inter + JetBrains Mono typography, `#2b49d8` primary accent, technical components), and all documentation content across `site/` and `site/docs/*.html` is synchronized with the v0.15 architecture (3 CLI verbs `install/update/uninstall`, 2 fail-closed pre-commit guards, markdown-as-truth, 0 mentions of obsolete SQLite / `harness.db`).
- success_signals:
  - S1 [design tokens & grid]: `site/css/tokens.css` and `site/css/main.css` define the full OpenPI token set and render the signature 30px 5% opacity grid overlay on the `#f6f5f0` background.
  - S2 [typography]: Space Grotesk, Instrument Serif, Inter, and JetBrains Mono are loaded and correctly assigned across displays, titles, headlines, body, labels, and terminal code blocks.
  - S3 [content accuracy]: `grep -ri "harness\.db\|sqlite\|preflight" site/` returns 0 hits across all HTML and JS files.
  - S4 [v0.15 contract]: `site/docs/cli.html` and `site/docs/architecture.html` describe only the 3 CLI verbs (`install`, `update`, `uninstall`), the 2 fail-closed guards, and the 6 stage playbooks.
  - S5 [link integrity]: All relative navigation links, footer links, and doc cross-references across `site/**/*.html` resolve cleanly with no 404s.

## Authority and Requirements

- authority:
  - `DESIGN.md` — authoritative specification for OpenPI visual language, colors, typography, layout, components, elevation, and brand voice.
  - `docs/ARCHITECTURE.md`, `docs/WORKFLOW.md`, `docs/README.md`, `docs/PROJECT.md` — authoritative repository architecture after the v0.15 release.
  - `cli/docs/CONTRACT.md` — authoritative CLI command contract for `zharness` v0.15.
- requirements:
  - R1 [accepted]: `site/css/tokens.css` defines all color tokens (`--color-surface: #f6f5f0`, `--color-surface-container: #efeee7`, `--color-primary: #2b49d8`, `--color-secondary: #57544b`, `--color-tertiary: #2440c0`, `--color-outline: #dcd8cb`, `--color-error: #d32f2f`), typography stacks, spacing scale (8px unit), and elevation shadows matching `DESIGN.md`.
  - R2 [accepted]: `site/css/main.css` implements the OpenPI technical blueprint layout, 30px grid background pattern, pill-shaped primary buttons (`border-radius: 999px`), technical cards, terminal blocks (`rounded: 12px`, `#efeee7` background), badges, and focus rings (`0 0 0 3px rgba(31, 71, 224, 0.12)`).
  - R3 [accepted]: Google Fonts link tags across all HTML files in `site/` load Space Grotesk, Instrument Serif, Inter, and JetBrains Mono.
  - R4 [accepted]: `site/index.html` is rewritten with the OpenPI blueprint aesthetic and brand voice, showcasing the 14 skills, the v0.15 markdown-first workflow, and direct install commands.
  - R5 [accepted]: `site/docs/architecture.html` is rewritten to reflect the true v0.15 architecture ("committed markdown is truth, 3 verbs, 2 fail-closed guards, 0 SQLite").
  - R6 [accepted]: `site/docs/workflow.html` is rewritten to document the 6 stage playbooks (`brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`) and the reduced modes.
  - R7 [accepted]: `site/docs/cli.html` is updated to match `cli/docs/CONTRACT.md` (`zharness install`, `zharness update`, `zharness uninstall`).
  - R8 [accepted]: `site/docs/skills.html`, `site/docs/installation.html`, and `site/docs/contributing.html` are updated with OpenPI components, tables, and correct repository procedures.
  - R9 [accepted]: All relative URLs and anchor links across `site/` are verified and functional.

## Non-goals

- NG1: no modifications to Go CLI code under `cli/` or test suites.
- NG2: no changes to the managed playbooks or git hooks (they are already verified at v0.15).
- NG3: no external runtime dependencies or build steps for the static site (vanilla HTML/CSS/JS only).

## Approach and Risks

- chosen approach:
  1. Overhaul CSS design tokens in `site/css/tokens.css` and base layout/component styles in `site/css/main.css` to match `DESIGN.md`.
  2. Rewrite and modernize HTML pages in `site/` and `site/docs/` wave-by-wave, injecting OpenPI structure (header mast, blueprint grid, terminal blocks, pill CTAs, technical tables) and updating content to v0.15.
  3. Verify all internal doc links, asset paths, responsive layout, and check zero stale SQLite references via grep.
- rejected alternatives:
  - Alternative 1: Use a static site generator (e.g. Astro, VitePress) — rejected to keep the site zero-build and lightweight.
  - Alternative 2: Update only styling without fixing stale technical content — rejected because web docs would mislead consumers regarding v0.15 CLI capabilities.
- risks and mitigations:
  - Risk: Broken links between `site/` root and `site/docs/` subpages. Mitigation: Run regex/link checker and verify relative path depth (`../` vs `./`).
  - Risk: Font loading flicker. Mitigation: Preconnect to Google Fonts and load standard fallback sans/serif/mono font stacks.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Append-only Progress is the sole task execution-status source. -->

- planning_status: planned
- phases:
  - phase_slug: p1-design-tokens-and-styles
    story_id: 01M0SITEOPIPH18M2VHZ5E9TB
    status: done
    goal: Establish the OpenPI CSS token system, blueprint grid canvas, typography, and reusable components.
    depends_on: none
    surfaces_allowed: site/css/tokens.css, site/css/main.css, site/js/main.js
    surfaces_avoided: cli/**, docs/**, skills/**
    requirements: R1, R2, R3
    waves:
      - wave: 1
        goal: CSS tokens and component stylesheets are aligned with DESIGN.md.
        tasks:
          - task: Update `site/css/tokens.css` with OpenPI colors, typography stacks, spacing, and elevation tokens.
            verify: Inspect `site/css/tokens.css` against `DESIGN.md` token values.
          - task: Update `site/css/main.css` with blueprint 30px grid background, pill buttons, terminal block styling, card surfaces, and responsive container (1240px max-width).
            verify: Check CSS classes and styles against component specifications in `DESIGN.md`.

  - phase_slug: p2-content-sync-and-pages
    story_id: 01M0SITEOPIPH28M2VHZ5E9TB
    status: done
    goal: Re-structure and rewrite all HTML pages under site/ and site/docs/ into OpenPI layout and sync with v0.15 architecture.
    depends_on: p1-design-tokens-and-styles
    surfaces_allowed: site/index.html, site/docs/*.html
    surfaces_avoided: cli/**, docs/**
    requirements: R3, R4, R5, R6, R7, R8
    waves:
      - wave: 1
        goal: Root index and core architecture & workflow docs are modernized.
        tasks:
          - task: Rewrite `site/index.html` with OpenPI hero section, font links, and updated overview.
            verify: Verify `site/index.html` structure and font links.
          - task: Rewrite `site/docs/architecture.html` and `site/docs/workflow.html` with v0.15 markdown-first truth and 6 playbooks.
            verify: Check `site/docs/architecture.html` and `site/docs/workflow.html` content.
      - wave: 2
        goal: Remaining documentation pages (cli, skills, installation, contributing) are updated.
        tasks:
          - task: Update `site/docs/cli.html` to reflect the 3 CLI verbs (`install`, `update`, `uninstall`).
            verify: Verify `site/docs/cli.html` matches `cli/docs/CONTRACT.md`.
          - task: Update `site/docs/skills.html`, `site/docs/installation.html`, and `site/docs/contributing.html` with OpenPI styling and current catalog.
            verify: Check `site/docs/skills.html`, `installation.html`, `contributing.html`.

  - phase_slug: p3-verification-and-polish
    story_id: 01M0SITEOPIPH38M2VHZ5E9TB
    status: done
    goal: Validate zero broken links, zero stale references to deleted v0.14 concepts, and clean visual presentation.
    depends_on: p2-content-sync-and-pages
    surfaces_allowed: site/**, this plan
    surfaces_avoided: cli/**, docs/**
    requirements: S3, S4, S5, R9
    waves:
      - wave: 1
        goal: All verification checks pass cleanly.
        tasks:
          - task: Scan `site/` for any leftover mentions of `harness.db`, `sqlite`, or `preflight`.
            verify: `grep -ri "harness\.db\|sqlite\|preflight" site/` returns 0 hits.
          - task: Verify all navigation and document links across `site/**/*.html`.
            verify: Check all `href` targets exist in the repo.

## Progress
<!-- Append-only log of execution events. Format: - [date timestamp] (phase-slug): description -->
- [2026-08-28 05:10 UTC] (p1-design-tokens-and-styles): Overhauled site/css/tokens.css, site/css/main.css, and site/js/main.js with full OpenPI design tokens (warm cream canvas #f6f5f0, 30px 5% opacity blueprint grid, Space Grotesk + Instrument Serif + Inter + JetBrains Mono, #2b49d8 primary button pills, technical cards, terminal blocks).
- [2026-08-28 05:12 UTC] (p2-content-sync-and-pages): Synchronized all 7 HTML pages (site/index.html and site/docs/*.html) with OpenPI layout, modern Google Fonts links, and v0.15 architecture (3 CLI verbs, 2 fail-closed pre-commit guards, markdown-first truth, 14 skills catalog, 0 SQLite).
- [2026-08-28 05:14 UTC] (p3-verification-and-polish): Verified zero broken links across all HTML files (7 files checked, 0 errors), verified bash scripts/verify-doc-links.sh (0 findings), and confirmed legacy SQLite scan clean.

## Decisions
<!-- Append-only record of mid-flight decisions and trade-offs made during planning and execution. -->
- [2026-08-28 05:08 UTC] (OpenPI Voice & Blueprint Layout): Kept zharness and mono-harness naming intact while fully adopting OpenPI blueprint design system and technical brand voice across all web documentation pages.

## Validation
<!-- Append-only log of test runs, reviews, and gate checks. Format per check playbook. -->
- [2026-08-28 05:15 UTC] (site-openpi-redesign): verdict: APPROVED
  - `bash scripts/verify-doc-links.sh`
  - `node -e 'const fs=require("fs"),p=require("path");["index.html","docs/architecture.html","docs/workflow.html","docs/cli.html","docs/skills.html","docs/installation.html","docs/contributing.html"].forEach(f=>{const c=fs.readFileSync("site/"+f,"utf8"),m=c.match(/href="([^"#:]+)"/g)||[];m.forEach(h=>{const t=h.slice(6,-1);if(!fs.existsSync(p.resolve("site",p.dirname(f),t)))throw new Error(f+" -> "+t)})});console.log("all site links OK")'`
  - `node -e 'const fs=require("fs");["index.html","docs/architecture.html","docs/workflow.html","docs/cli.html","docs/skills.html","docs/installation.html","docs/contributing.html"].forEach(f=>{const c=fs.readFileSync("site/"+f,"utf8");if(/harness\.db/i.test(c)&&!/historical/i.test(c))throw new Error("found unhandled harness.db in "+f)});console.log("legacy terms clean")'`
## Current State

- active_plan_id: 01M0SITEOPIPLAN8M2VHZ5E9TB
- active_intake_id: 01M0SITEOPIINTK8M2VHZ5E9TB
- active_phase: done
- blockers: none
- exact_next_action: check

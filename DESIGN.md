---
version: alpha
name: OpenPI
description: >-
  OpenPI is a Pi-native capability layer that consolidates background terminals, subagents, workflows, and continuous
  tasks into a single workbench—zero resident tools until needed.
logo:
  src: >-
    data:image/svg+xml,%3Csvg class%3D%22pi-logo-nav%22 viewBox%3D%22165.29 165.29 469.07 469.07%22
    xmlns%3D%22http%3A//www.w3.org/2000/svg%22%3E%3Cpath fill%3D%22%2309090b%22 fill-rule%3D%22evenodd%22 d%3D%22M165.29
    165.29 H517.36 V400 H400 V517.36 H282.65 V634.72 H165.29 Z M282.65 282.65 V400 H400 V282.65 Z%22/%3E%3Cpath
    fill%3D%22%2309090b%22 d%3D%22M517.36 400 H634.72 V634.72 H517.36 Z%22/%3E%3C/svg%3E
colors:
  surface: '#f6f5f0'
  surface-dim: '#efeee7'
  surface-bright: '#ffffff'
  surface-container-lowest: '#f6f5f0'
  surface-container-low: '#f6f5f0'
  surface-container: '#efeee7'
  surface-container-high: '#e9e8e0'
  surface-container-highest: '#dcd8cb'
  on-surface: '#12110e'
  on-surface-variant: '#57544b'
  inverse-surface: '#12110e'
  inverse-on-surface: '#f6f5f0'
  outline: '#dcd8cb'
  outline-variant: '#c6c2b2'
  surface-tint: '#2b49d8'
  primary: '#2b49d8'
  on-primary: '#f6f5f0'
  primary-container: '#8fa8ff'
  on-primary-container: '#12110e'
  inverse-primary: '#8fa8ff'
  secondary: '#57544b'
  on-secondary: '#f6f5f0'
  secondary-container: '#a8a494'
  on-secondary-container: '#12110e'
  tertiary: '#2440c0'
  on-tertiary: '#f6f5f0'
  tertiary-container: '#1f47e0'
  on-tertiary-container: '#f6f5f0'
  error: '#d32f2f'
  on-error: '#f6f5f0'
  error-container: '#ffcdd2'
  on-error-container: '#12110e'
  primary-fixed: '#8fa8ff'
  primary-fixed-dim: '#2b49d8'
  on-primary-fixed: '#12110e'
  on-primary-fixed-variant: '#2b49d8'
  secondary-fixed: '#a8a494'
  secondary-fixed-dim: '#57544b'
  on-secondary-fixed: '#f6f5f0'
  on-secondary-fixed-variant: '#57544b'
  tertiary-fixed: '#1f47e0'
  tertiary-fixed-dim: '#2440c0'
  on-tertiary-fixed: '#f6f5f0'
  on-tertiary-fixed-variant: '#2440c0'
  background: '#f6f5f0'
  on-background: '#12110e'
  surface-variant: '#dcd8cb'
typography:
  display:
    fontFamily: Space Grotesk
    fontSize: 76px
    fontWeight: '700'
    lineHeight: 84px
    letterSpacing: '-0.045em'
  headline-lg:
    fontFamily: Instrument Serif
    fontSize: 46px
    fontWeight: '400'
    lineHeight: 54px
    letterSpacing: '-0.02em'
  headline-md:
    fontFamily: Instrument Serif
    fontSize: 34px
    fontWeight: '400'
    lineHeight: 42px
    letterSpacing: 0.004em
  title-lg:
    fontFamily: Space Grotesk
    fontSize: 26px
    fontWeight: '600'
    lineHeight: 32px
    letterSpacing: 0.02em
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 28px
    letterSpacing: 0.005em
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 28px
    letterSpacing: 0.005em
  label-md:
    fontFamily: JetBrains Mono
    fontSize: 13.5px
    fontWeight: '450'
    lineHeight: 20px
    letterSpacing: 0.04em
  label-sm:
    fontFamily: JetBrains Mono
    fontSize: 11px
    fontWeight: '400'
    lineHeight: 16px
    letterSpacing: 0.08em
rounded:
  sm: 3px
  DEFAULT: 4px
  md: 8px
  lg: 12px
  xl: 16px
  full: 999px
spacing:
  unit: 8px
  xs: 4px
  sm: 12px
  md: 24px
  lg: 40px
  xl: 64px
  gutter: 24px
  container-max: 1240px
elevation:
  sm: 0 1px 0 {colors.outline}
  md: 0 20px 44px -24px rgba(18, 17, 14, 0.28)
  lg: 0 0 0 3px rgba(31, 71, 224, 0.12)
layout:
  containerMaxWidth: 1240px
  gridColumns: 12
components:
  button-primary:
    backgroundColor: '{colors.primary}'
    textColor: '{colors.on-primary}'
    typography: '{typography.label-md}'
    rounded: '{rounded.full}'
    padding: 8px 16px
    height: auto
    border: none
    transition: background-color 0.25s
  button-primary-hover:
    backgroundColor: '{colors.tertiary}'
    textColor: '{colors.on-primary}'
  button-secondary:
    backgroundColor: transparent
    textColor: '{colors.primary}'
    typography: '{typography.label-md}'
    rounded: '{rounded.DEFAULT}'
    padding: 6px 13px
    height: auto
    border: 1px solid {colors.outline-variant}
    transition: border-color 0.25s, box-shadow 0.25s
  button-secondary-hover:
    borderColor: '{colors.primary}'
    boxShadow: 0 0 0 3px rgba(143, 168, 255, 0.18)
  card:
    backgroundColor: '{colors.surface-container}'
    rounded: '{rounded.lg}'
    padding: 22px 24px
    border: 1px solid {colors.outline}
    boxShadow: '{elevation.md}'
  card-hover:
    backgroundColor: '{colors.surface-dim}'
    transition: background-color 0.3s
  input-field:
    backgroundColor: '{colors.surface}'
    textColor: '{colors.on-surface}'
    typography: '{typography.body-md}'
    rounded: '{rounded.DEFAULT}'
    padding: 13px 20px
    border: 1px solid {colors.outline}
    transition: border-color 0.25s, box-shadow 0.25s
  input-field-focus:
    borderColor: '{colors.primary}'
    boxShadow: 0 0 0 3px rgba(31, 71, 224, 0.12)
  terminal-block:
    backgroundColor: '{colors.surface-dim}'
    rounded: '{rounded.lg}'
    border: 1px solid {colors.on-surface}
    padding: 16px 18px
    boxShadow: '{elevation.md}'
    fontFamily: JetBrains Mono
    fontSize: 12.5px
    lineHeight: '2'
  nav-link:
    textColor: '{colors.on-surface-variant}'
    typography: '{typography.label-sm}'
    transition: color 0.25s
    textDecoration: none
  nav-link-active:
    textColor: '{colors.on-surface}'
    position: relative
  badge:
    backgroundColor: '{colors.primary-container}'
    textColor: '{colors.on-primary-container}'
    typography: '{typography.label-sm}'
    rounded: '{rounded.full}'
    padding: 4px 8px
  divider:
    borderColor: '{colors.outline}'
    borderWidth: 1px
    borderStyle: solid
---

## Overview

OpenPI embodies a philosophy of **Architectural Minimalism with Technical Precision**—a design language that treats complexity as a structural problem, not a visual one. The brand serves developers and AI engineers who demand clarity, performance, and zero cognitive overhead. The UI evokes the aesthetic of a technical blueprint: a warm cream canvas (#F6F5F0) layered with a subtle 30px grid, crisp monochromatic typography, and a single accent blue (#2B49D8) that signals interactivity without noise. Every element earns its place through function; decoration is forbidden. The emotional response is one of calm competence: this tool will not surprise you, will not hide complexity, and will execute exactly as promised.

The brand voice is direct, precise, and occasionally playful—never breathless or marketing-speak. Vocabulary favors concrete nouns ("workbench," "terminal," "session") over abstractions ("ecosystem," "synergy"). Tone is conversational but technical; a user should feel they're reading documentation written by someone who ships code. Example: "Pi always owns the main session. Subagents are ephemeral—they exist to serve, not to persist."

## Colors

The color system is deliberately constrained to enforce visual discipline. **Primary (#2B49D8)** is the signature accent—used exclusively on CTAs (the "Open" button), focus states, active navigation indicators, and inline highlights within prose. It is never used for backgrounds or large surfaces. **Secondary (#57544B)** is a warm mid-tone ink used for supporting text, metadata, and secondary UI labels; it creates hierarchy without introducing a second brand color. **Tertiary (#2440C0)** is a deeper blue reserved for terminal output, code blocks, and success states—it provides visual distinction from primary while maintaining the blue family. **Surface stack**: the page background is #F6F5F0 (warm cream), with a 30px grid overlay at 5% opacity (rgba(18, 17, 14, 0.05)) to suggest structure

## Typography

The type system pairs **Space Grotesk** (geometric sans-serif, 700 weight) for display and wordmarks with **Instrument Serif** (high-contrast serif, 400 weight) for headlines and pull quotes, creating a deliberate tension between precision and humanity. Body copy uses **Inter** (16px, 400 weight, 1.75 line-height) for readability at screen resolution; Chinese text switches to **Noto Serif SC** (600 weight) to maintain visual weight parity. Monospace copy—navigation labels, terminal output, code snippets—uses **JetBrains Mono** (11–13.5px, 450 weight, 0.04–0.22em letter-spacing) to signal technical content. Display text (76px, -0.045em tracking) uses a tight line-height (0.94) to create visual density; headlines (34–46px) use 1.16–1.22 line-height for breathing room. When small labels appea

## Layout

The layout uses a fixed 1240px container (--max: 1240px) centered with fluid padding (clamp(20px, 4.5vw, 56px)) that scales from 20px on mobile to 56px on desktop. The hero section spans 100svh with a 12-column grid that collapses to single-column on screens ≤900px. Sections use a 3-column or 4-column grid for content blocks, with 24px gutters between columns and 24px padding within cells. Vertical rhythm is maintained through a spacing scale: xs (4px), sm (12px), md (24px), lg (40px), xl (64px). Section separation uses lg spacing (40–64px, responsive via clamp()) to create breathing room; internal component spacing uses md (24px). The page background includes a subtle grid pattern (30px × 30px, 5% opacity) that reinforces the architectural aesthetic without competing for attention. All ma

## Elevation & Depth

Depth is conveyed through **layered borders and subtle shadows**, not through dark overlays or glassmorphism. Level 1 (base): the page background (#F6F5F0) with grid overlay. Level 2 (standard cards): 1px solid border at {colors.outline} (#DCD8CB) + box-shadow: 0 1px 0 var(--rule), 0 20px 44px -24px rgba(18, 17, 14, 0.28). Level 3 (elevated/terminal blocks): same border + stronger shadow (0 20px 44px -24px rgba(18, 17, 14, 0.28)). Interactive elements (buttons, inputs) use a focus ring: 0 0 0 3px rgba(31, 71, 224, 0.12) on :focus-visible. Hover states transition background-color over 0.25s and

## Shapes

The shape philosophy is **Technical Sharpness with Selective Softness**—corners are sharp (3–4px) by default to reinforce precision, but CTAs and interactive affordances use full-radius (999px) to signal "clickability." Buttons use border-radius: 999px (pill shape) to create visual distinction from static content. Input fields and cards use 4px (DEFAULT) or 12px (lg) for a technical feel. Terminal blocks use 12px to suggest a contained, bounded space. The rationale: sharp corners feel intentional and engineered; rounded corners feel interactive and forgiving. This contrast helps users quickly

## Components

### Action Elements
Buttons are the primary interactive affordance. **button-primary** uses background: {colors.primary} (#2B49D8), color: {colors.on-primary} (#F6F5F0), border-radius: 999px, padding: 8px 16px, font-family: JetBrains Mono, font-size: 11.5px, font-weight: 450. On hover, background transitions to {colors.tertiary} (#2440C0) over 0.25s. **button-secondary** uses a transparent background with border: 1px solid {colors.outline-variant}, border-radius: 4px, padding: 6px 13px. On hover, border-color shifts to {colors.primary} and a 3px focus ring appears (box-shadow: 0 0 0 3px rgba(31, 71, 224, 0.12)). Both buttons use :focus-visible to apply a 2px outline at {colors.primary} with 3px offset.

### Containers & Surfaces
**card** uses background: {colors.surface-container} (#EFEEE7

## Do's and Don'ts

**Do**
- Do use {colors.primary} (#2B49D8) exclusively for CTAs, focus states, and inline highlights—never as a background fill.
- Do apply a 0.25s transition on all color and border-color changes to reduce visual jarring.
- Do use the 30px grid overlay (5% opacity) as a visual anchor; never remove it from the page background.
- Do pair Space Grotesk headlines with Instrument Serif body text to create intentional typographic contrast.
- Do use border-radius: 999px for buttons and border-radius: 4px for inputs to signal affordance hierarchy.
- Do apply focus rings (0 0 0 3px rgba(31, 71, 224, 0.12)) to all interactive elements on :focus-visible.

**Don't**
- Don't use primary blue (#2B49D8) for large surface fills or backgrounds—it overwhelms the page.
- Don't introduce new accent colors; the system is intentionally constrained to primary, secondary, and tertiary.
- Don't apply shadows larger than 0 20px 44px -24px rgba(18, 17, 14, 0.28); depth is subtle, not dramatic.
- Don't use rounded corners (border-radius > 4px) on input fields or data tables; reserve full-radius for CTAs only.
- Don't animate scale or rotation on hover; transitions should be color, border-color, or opacity only.
- Don't remove the grid overlay or change the page background color; it is a core part of the brand identity.

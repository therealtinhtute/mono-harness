# Style: light-pill (default)

High-level topology only: the main components and how they connect. One idea per diagram.

## Canvas

- Light background, `draw` (hand-drawn) font, black card borders and body text.
- Title at top-left, size `xl`; one-sentence subtitle below it in `grey`.
- Horizontal bands split layers: a grey rule across the canvas, a filled dot plus a size `m` label at its left edge. 2–4 bands.

## Cards

- White rectangle with a black `draw` border; 2–4 short lines of body text, no sentences.
- Each card carries one pill tag straddling its top border, left-aligned: `[ ● title ]`.
- Pill: oval, border and text in the card's color, background the light tint of the same color (`fill:'solid'`). Height 34, width `tag.length * 10 + 52`, text size `s` at scale 0.9, centered both ways — small, never taller than one body line.
- Pill text is a separate fixed-width text shape (`textAlign:'middle'`) over the oval, not the oval's label: geo labels ignore `scale` and need ≈50px height, so a 34px oval would raise a `growY-on-shape` lint.
- Card width fits the longest line without wrapping (≈ 13px per char at size `m`); height ≈ 50 + 25 per line.
- Keep the left 200px clear for band labels.

## Color

One color per role or flow; the same color on the pill, its band dot and the arrows that leave the card.

| Role | Color |
|---|---|
| user, entry, evidence | `light-red` |
| router, binary, managed docs | `blue` |
| main path, source, success | `green` |
| bug, plan, consumer data | `orange` |
| host, reads/writes, read-only | `violet` |
| resume, loop back | `light-blue` |
| learn, retro | `yellow` |
| guard, gate, failure | `red` |

## Arrows

- Colored by the source card's role; `draw` dash for flow, `dashed` for gates or loops back.
- Label with one verb or two words (`install`, `reads`, `gate commits`).
- Bind both ends to cards. When a pill would hide the arrowhead, place the cards so the arrow lands mid-border or use `bend`.
- Avoid crossings: put a card directly under the card that feeds it.

## Checks before delivery

- Export the PNG and look at it: no wrapped pill text, no card overlap, every arrowhead visible.
- `lints` from `tl.sh draw` holds at most `overlapping-text` from pills straddling card borders.

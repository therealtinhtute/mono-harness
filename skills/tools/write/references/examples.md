# examples

Use these together with the files in `../examples/` when validating `write`.

## Available example sets
- `../examples/triggers.md` — trigger/non-trigger phrases
- `../examples/mode-selection.md` — which mode should win for a given prompt
- `../examples/eval-queries.md` — evaluation prompts for activation quality

## Quick scenarios
### Example 1 — light rewrite
Input: `Rewrite this status update so it sounds natural but keep it short.`
Output: one cleaned-up version, no explanation.

### Example 2 — bilingual consistency
Input: `Polish this English/Vietnamese announcement and keep both sides aligned.`
Output: one EN+VI pair with consistent meaning and tone.

### Example 3 — missing source text
Input: `Rewrite my message to sound warmer.`
Output: ask for the exact source text in one short blocking question.

# SDLC cost model

The measured-constant cost model behind `docs/audit/sdlc-token-cache-audit.md`
and `docs/audit/sdlc-gap-analysis.md`, preserved so the optimization plan's
success signal ("re-run the cost model, `work` ≤ 32 turns, per-phase gate cost
−50%" — `docs/plans/completed/sdlc-token-optimization.md`) is executable rather
than re-derivable.

- `model.py` — per-stage turn ledgers and raw token model. Every char constant
  is measured (dogfood runs of `zharness`, 2026-08-11), not estimated; the
  chars→tokens divisor (3.5) is the only estimate and cancels in all ratios.
  Constants were re-measured after the R1-R9 release against `zharness v0.9.0`
  in a throwaway repo with a filled 3-phase list-form plan (the audit method).
- `cache_model.py` — prompt-cache-aware cost layer on top of `model.py`:
  0.10× reads / 1.25× 5-minute writes, model-scoped invalidation, 20-block
  lookback, per-model minimum prefixes, list prices as of 2026-08-11.

Run from this directory (`cache_model.py` imports `model.py` by relative path):

```bash
cd docs/audit/cost-model && python3 model.py && python3 cache_model.py
```

After landing a plan phase, update the measured constants in `model.py`
(CLI output bytes, turn ledgers) from a fresh dogfood run before re-comparing.

## Re-measurement after R1-R9 (2026-08-11, zharness v0.9.0)

| Signal | Target | Re-measured | Verdict |
|---|---|---|---|
| `work` turns (5 tasks / 2 waves) | ≤ 32 | **34** (was 37) | ❌ miss by 2 |
| per-phase gate cost | −50% | **−84%** ($0.4364 → $0.0681 warm) | ✅ |
| chain total per phase | ~−36% | **−38%** ($0.9927 → $0.6166) | ✅ |

`work` lands at 34 turns, not the audit's modeled 32, because the shipped
playbook (work.md step 9) keeps the wave summary as a separate call on top of
the batched flush — two trace calls per wave instead of one. The per-phase
gate (R4, in-session sonnet) is the dominant win: 14 turns at $0.0681 warm vs
the legacy opus full at $0.4364; the full review now costs $0.4364 exactly
once per initiative at final closure, not per phase.

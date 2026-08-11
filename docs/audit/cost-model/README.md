# SDLC cost model

The measured-constant cost model behind `docs/audit/sdlc-token-cache-audit.md`
and `docs/audit/sdlc-gap-analysis.md`, preserved so the optimization plan's
success signal ("re-run the cost model, `work` ≤ 32 turns, per-phase gate cost
−50%" — `docs/plans/active/sdlc-token-optimization.md`) is executable rather
than re-derivable.

- `model.py` — per-stage turn ledgers and raw token model. Every char constant
  is measured (dogfood run of `zharness` built from source, 2026-08-11), not
  estimated; the chars→tokens divisor (3.5) is the only estimate and cancels
  in all ratios.
- `cache_model.py` — prompt-cache-aware cost layer on top of `model.py`:
  0.10× reads / 1.25× 5-minute writes, model-scoped invalidation, 20-block
  lookback, per-model minimum prefixes, list prices as of 2026-08-11.

Run from this directory (`cache_model.py` imports `model.py` by relative path):

```bash
cd docs/audit/cost-model && python3 model.py && python3 cache_model.py
```

After landing a plan phase, update the measured constants in `model.py`
(CLI output bytes, turn ledgers) from a fresh dogfood run before re-comparing.

#!/usr/bin/env python3
"""Cache-aware SDLC cost model.

Caching facts (from the claude-api skill, shared/prompt-caching.md):
  - prefix match; render order tools -> system -> messages
  - cache read  = 0.10x input price
  - cache write = 1.25x input price (5m TTL); 2.00x (1h TTL)
  - caches are MODEL-SCOPED: switching model invalidates everything
  - 20-block lookback: a breakpoint walks back <=20 content blocks
  - min cacheable prefix: opus-5 512 tok, sonnet-5 1024 tok, haiku-4.5 ~4096 tok
Prices per MTok (2026-08-11; sonnet-5 intro pricing active through 08-31):
"""
CH = 3.5
def t(c): return c / CH

PRICE = {  # (input, output) $/MTok
    "opus":   (5.00, 25.00),
    "sonnet": (2.00, 10.00),   # Sonnet 5 intro; $3/$15 after 2026-08-31
    "haiku":  (1.00,  5.00),
}
CACHE_READ, CACHE_WRITE = 0.10, 1.25

import importlib.util, sys
spec = importlib.util.spec_from_file_location("m", "model.py")
m = importlib.util.module_from_spec(spec)
_out = sys.stdout
class Null:
    def write(self, *a): pass
    def flush(self): pass
sys.stdout = Null(); spec.loader.exec_module(m); sys.stdout = _out

BASELINE = m.BASELINE

def cost_cached(fixed, turns, model, cold_start=True):
    """Per turn: prefix served from cache (0.1x), new delta written (1.25x).
    cold_start=False models an unchanged prefix already warm from a prior stage
    on the SAME model."""
    pin, pout = PRICE[model]
    ctx = BASELINE + fixed
    read_tok = write_tok = out_tok = 0
    first = True
    for _, out, res in turns:
        if first and cold_start:
            write_tok += t(ctx)          # cold: whole prefix written
        elif first:
            read_tok += t(ctx)           # warm: prefix already cached
        else:
            read_tok += t(ctx)
        out_tok += t(out)
        write_tok += t(out + res)        # this turn's delta enters the cache
        ctx += out + res
        first = False
    cost = (read_tok*CACHE_READ + write_tok*CACHE_WRITE)*pin/1e6 + out_tok*pout/1e6
    uncached = (read_tok + write_tok)*pin/1e6 + out_tok*pout/1e6
    return cost, uncached, read_tok, write_tok, out_tok

STAGES = [
    ("watzup",   m.watzup(),   "haiku"),
    ("brainstorm", m.brainstorm(), "opus"),
    ("to-plan",  m.to_plan(),  "opus"),
    ("work",     m.work_full(), "sonnet"),
    ("check",    m.check_full(), "opus"),
    ("git",      m.git_stage(), "sonnet"),
    ("handoff",  m.handoff(),  "sonnet"),
]

print("=== PER-STAGE COST, ONE PHASE (cold cache each stage: model switches) ===")
print(f"{'stage':<12}{'model':<8}{'read_tok':>10}{'write_tok':>10}{'out_tok':>9}"
      f"{'$cached':>10}{'$uncached':>11}{'saving':>8}")
print("-"*78)
tot_c = tot_u = 0
for name,(fixed,turns),mod in STAGES:
    c,u,r,w,o = cost_cached(fixed, turns, mod)
    tot_c += c; tot_u += u
    print(f"{name:<12}{mod:<8}{r:>10,.0f}{w:>10,.0f}{o:>9,.0f}{c:>10.4f}{u:>11.4f}{(1-c/u)*100:>7.0f}%")
print("-"*78)
print(f"{'TOTAL':<20}{'':>10}{'':>10}{'':>9}{tot_c:>10.4f}{tot_u:>11.4f}{(1-tot_c/tot_u)*100:>7.0f}%")

print()
print("=== COST OF MODEL SWITCHING (cache is model-scoped) ===")
# the pipeline as-is: every stage is a cold cache because the model changed
# counterfactual: work+check on the SAME model -> check's prefix stays warm
wf, wt = m.work_full(); cf, ct = m.check_full()
c_switch,_,_,_,_ = cost_cached(cf, ct, "opus", cold_start=True)
c_warm,_,_,_,_   = cost_cached(cf, ct, "sonnet", cold_start=False)
c_same_cold,_,_,_,_ = cost_cached(cf, ct, "sonnet", cold_start=True)
print(f"  check full on opus, cold (current)          ${c_switch:.4f}")
print(f"  check full on sonnet, cold                  ${c_same_cold:.4f}")
print(f"  check full on sonnet, warm from work        ${c_warm:.4f}   <- same-model, cache survives")
print(f"  per-phase saving from not switching:        ${c_switch-c_warm:.4f}  ({(1-c_warm/c_switch)*100:.0f}%)")

print()
print("=== 20-BLOCK LOOKBACK: does the work loop blow the window? ===")
# each tool round-trip contributes ~2 content blocks (assistant tool_use + tool_result)
for tasks in (3,5,8,12):
    f,tr = m.work_full(tasks=tasks)
    blocks = len(tr)*2
    print(f"  {tasks:>2} tasks -> {len(tr):>2} turns -> ~{blocks:>3} content blocks "
          f"({'OK' if blocks<=20 else f'{blocks//20} breakpoint refreshes needed'})")

print()
print("=== MINIMUM CACHEABLE PREFIX vs what each stage loads ===")
MIN = {"opus":512, "sonnet":1024, "haiku":4096}   # opus-5 / sonnet-5 / haiku-4.5
for name,(fixed,turns),mod in STAGES:
    tok = t(BASELINE+fixed)
    print(f"  {name:<12}{mod:<8}prefix={tok:>7,.0f} tok  min={MIN[mod]:>5}  "
          f"{'cacheable' if tok>=MIN[mod] else 'TOO SHORT — silently uncached'}")

print()
print("=== DEGRADED query plan: cache cost, not just one-shot cost ===")
for deg in (True, False):
    f,tr = m.work_full(degraded=deg)
    c,u,r,w,o = cost_cached(f, tr, "sonnet")
    print(f"  degraded={deg!s:<6} read={r:>8,.0f} write={w:>8,.0f}  ${c:.4f}")
d,_ = m.work_full(degraded=True); n,_ = m.work_full(degraded=False)

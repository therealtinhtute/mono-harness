#!/usr/bin/env python3
"""SDLC token model. All char constants are MEASURED (2026-08-11, second
dogfood run) against zharness v0.9.0 (R1-R9 release) in a throwaway repo with
a filled 3-phase list-form plan, per the audit method. Turn counts come from
the playbooks' own mandated steps (work.md/check.md post-R5/R6).
Chars->tokens at 3.5 (markdown) is an estimate; ratios between stages are
unaffected by the divisor."""

CH = 3.5
def t(chars): return chars / CH

# ---- measured fixed loads (chars) ----
SKILL = dict(watzup=1135, brainstorm=1398, to_plan=1208, work=1377,
             check=1341, git=2768, handoff=1104)
PLAYBOOK = dict(watzup=4424, brainstorm=5657, to_plan=5316, work=10987,
                check=10233, git=7330, handoff=8078)
GIT_REFS = 1851 + 2517          # git/SKILL.md still loads its references/
BASELINE = 12000                # system prompt + tool defs, replayed every turn

# ---- measured CLI JSON output (chars), zharness v0.9.0 ----
# Re-measured 2026-08-11 after R1-R9. preflight_ctx scales with phase count:
# 668 B at 3 phases (this model's shape); 6.3 KB in a repo with 20+ phases.
CLI = dict(preflight=138, preflight_ctx=668, query_plan_phase_ok=1332,
           query_plan_phase_degraded=6848, query_plan_current=446,
           query_traces=1018, query_phases=367, resume=196, audit=66,
           trace_add=35, trace_add_batch=181, run_create=35, check_record=105,
           handoff_record=36, story=74)

# ---- measured artifact sizes (chars) ----
PLAN_START = 6401               # filled 3-phase plan (list form)
GATE_OUTPUT = 497 + 26          # go test ./... + verify-doc-links.sh

# (name, agent_out_chars, tool_result_chars) per turn
def work_full(tasks=5, waves=2, degraded=False):
    """Post-R5 ledger (work.md steps 7/9): per-task `trace add` is gone; each
    wave flushes pending DONE entries in one batched --tasks call and keeps
    the wave summary call. Measured: 34 turns at 5 tasks / 2 waves (was 37)."""
    q = CLI['query_plan_phase_degraded'] if degraded else CLI['query_plan_phase_ok']
    turns  = [("preflight", 120, CLI['preflight_ctx'])]
    turns += [("query plan --section phase", 130, q)]
    turns += [("query traces", 110, CLI['query_traces'])]
    turns += [("git diff (boundary check)", 90, 1800)]
    turns += [("run create", 140, CLI['run_create'])]
    turns += [("edit plan: phase->in-progress", 700, 200),
              ("edit plan: current state", 900, 200),
              ("progress entry (phase start)", 280, CLI['trace_add'])]
    for i in range(tasks):
        turns += [(f"t{i} read target", 110, 3500),
                  (f"t{i} read neighbour", 110, 2600),
                  (f"t{i} edit", 1400, 250),
                  (f"t{i} verify cmd", 130, 700)]
    for w in range(waves):
        turns += [(f"wave{w} trace flush", 190, CLI['trace_add_batch']),
                  (f"wave{w} trace summary", 190, CLI['trace_add'])]
    turns += [("refresh current state", 900, 200),
              ("query phases", 100, CLI['query_phases'])]
    return SKILL['work'] + PLAYBOOK['work'], turns

def check_gate():
    """Per-phase in-session gate (work.md step 11, R4; check.md gate mode):
    automated checks + lifecycle audit only — no complete manual review.
    No separate `resume` call (R6: position comes from the preflight context
    packet). Measured: 14 turns on sonnet, cache-warm from work."""
    turns  = [("preflight", 120, CLI['preflight_ctx'])]
    turns += [("read active plan", 110, PLAN_START + 1200)]
    turns += [("git diff", 90, 9000)]              # the phase diff
    # automated gate: this repo runs 2 commands; typical repo 3-4
    for c in range(4):
        turns += [(f"gate cmd {c}", 120, GATE_OUTPUT if c else 3000)]
    turns += [("read evals/failures.md", 110, 2200)]
    turns += [("audit", 100, CLI['audit'])]
    turns += [("check record", 900, CLI['check_record'])]   # long proof-links JSON
    turns += [("edit plan: validation entry", 1100, 200),
              ("edit plan: phase->checked", 700, 200),
              ("query phases", 100, CLI['query_phases'])]
    turns += [("verdict output block", 600, 0)]
    return SKILL['check'] + PLAYBOOK['check'], turns

def check_full(files=6):
    """Initiative-final closure review (handoff.md step 6 precondition): the
    gate plus the complete Security/Performance/Architecture/Code Quality
    review. Runs once per initiative on opus via the /check skill (R4).
    Measured: 22 turns (was 23 — the separate resume call is gone)."""
    turns  = [("preflight", 120, CLI['preflight_ctx'])]
    turns += [("read active plan", 110, PLAN_START + 1200)]
    turns += [("git diff", 90, 9000)]              # the phase diff
    for c in range(4):
        turns += [(f"gate cmd {c}", 120, GATE_OUTPUT if c else 3000)]
    turns += [("read evals/failures.md", 110, 2200)]
    # the complete manual review: every changed file read in full,
    # plus sibling search for class-of-bug
    for f in range(files):
        turns += [(f"review read file {f}", 110, 5200)]
    turns += [("grep sibling instances", 130, 2400),
              ("grep sibling instances 2", 130, 1900)]
    turns += [("audit", 100, CLI['audit'])]
    turns += [("check record", 900, CLI['check_record'])]
    turns += [("edit plan: validation entry", 1100, 200),
              ("edit plan: phase->checked", 700, 200),
              ("query phases", 100, CLI['query_phases'])]
    # the review prose itself: the largest single agent output in the chain
    turns += [("review write-up (final turn)", 6000, 0)]
    return SKILL['check'] + PLAYBOOK['check'], turns

def to_plan(phases=3):
    turns  = [("preflight", 120, CLI['preflight'])]
    turns += [("read active plan", 110, 2400)]
    for p in range(phases):
        turns += [(f"story {p}", 150, CLI['story'])]
    turns += [("write approach+phases", 4200, 200)]   # the big authored block
    turns += [("query phases", 100, CLI['query_phases'])]
    return SKILL['to_plan'] + PLAYBOOK['to_plan'], turns

def handoff():
    turns  = [("preflight", 120, CLI['preflight_ctx'])]
    turns += [("git status/log/diffstat", 150, 2200)]
    turns += [("query plan --section current-state", 130, CLI['query_plan_current'])]
    turns += [("query traces --tail 10", 120, 2400)]
    turns += [("query decisions --tail 10", 120, 1400)]
    turns += [("query checks --tail 3", 120, 900)]
    turns += [("handoff record", 500, CLI['handoff_record'])]
    turns += [("edit plan: current state", 1300, 200),
              ("edit plan: phase->done", 700, 200),
              ("query phases", 100, CLI['query_phases'])]
    return SKILL['handoff'] + PLAYBOOK['handoff'], turns

def watzup():
    turns  = [("preflight", 120, CLI['preflight_ctx'])]
    turns += [("git status/log", 150, 2200)]
    turns += [("query plan --section current-state", 130, CLI['query_plan_current'])]
    turns += [("query traces --tail", 120, 2000)]
    turns += [("recap write-up", 1400, 0)]
    return SKILL['watzup'] + PLAYBOOK['watzup'], turns

def brainstorm():
    turns  = [("preflight", 120, CLI['preflight'])]
    turns += [("explore repo (glob/grep)", 200, 3000)]
    turns += [("read 3 files", 330, 9000)]
    turns += [("AskUserQuestion round 1", 800, 400)]
    turns += [("AskUserQuestion round 2", 800, 400)]
    turns += [("intake", 200, 60)]
    turns += [("write locked plan", 3800, 200)]
    return SKILL['brainstorm'] + PLAYBOOK['brainstorm'], turns

def git_stage():
    turns  = [("preflight/status", 150, 2200)]
    turns += [("git diff review", 100, 9000)]
    turns += [("git add+commit", 600, 400)]
    turns += [("git push", 120, 500)]
    return SKILL['git'] + PLAYBOOK['git'] + GIT_REFS, turns

def cost(fixed, turns):
    """Input = sum over turns of full context replayed. Output = agent text."""
    ctx = BASELINE + fixed
    tin = tout = 0
    for _, out, res in turns:
        tin += ctx                 # whole context re-sent as input this turn
        tout += out
        ctx += out + res
    return t(tin), t(tout), len(turns), t(ctx)

stages = [("watzup", watzup()), ("brainstorm", brainstorm()),
          ("to-plan", to_plan()), ("work full (5 tasks)", work_full()),
          ("check gate (per phase)", check_gate()), ("git", git_stage()),
          ("handoff", handoff())]

# opus is ~5x sonnet on both in and out; haiku ~1/3 sonnet
MODEL = dict(watzup=("haiku",0.33), brainstorm=("opus",5.0), to_plan=("opus",5.0),
             work=("sonnet",1.0), check=("sonnet",1.0), git=("sonnet",1.0),
             handoff=("sonnet",1.0))
key = {"watzup":"watzup","brainstorm":"brainstorm","to-plan":"to_plan",
       "work full (5 tasks)":"work","check gate (per phase)":"check",
       "git":"git","handoff":"handoff"}

print(f"{'stage':<22}{'turns':>6}{'tok_in':>10}{'tok_out':>9}{'in+out':>10}"
      f"{'ctx_end':>9}  {'model':<7}{'weighted':>10}")
print("-"*86)
rows=[]
for name,(fixed,turns) in stages:
    ti,to_,n,ce = cost(fixed,turns)
    m,mult = MODEL[key[name]]
    rows.append((name,n,ti,to_,ti+to_,ce,m,(ti+to_)*mult))
    print(f"{name:<22}{n:>6}{ti:>10,.0f}{to_:>9,.0f}{ti+to_:>10,.0f}{ce:>9,.0f}  {m:<7}{(ti+to_)*mult:>10,.0f}")
tot_raw=sum(r[4] for r in rows); tot_w=sum(r[7] for r in rows)
print("-"*86)
print(f"{'TOTAL 1 phase':<22}{sum(r[1] for r in rows):>6}{'':>10}{'':>9}{tot_raw:>10,.0f}{'':>9}{'':>9}{tot_w:>10,.0f}")
print(f"(final phase only: +check full {len(check_full()[1])} turns on opus — once per initiative)")
print()
print("share of raw in+out:")
for r in sorted(rows,key=lambda r:-r[4]): print(f"  {r[0]:<22}{r[4]/tot_raw*100:>6.1f}%")
print("share of model-weighted:")
for r in sorted(rows,key=lambda r:-r[7]): print(f"  {r[0]:<22}{r[7]/tot_w*100:>6.1f}%")

print()
print("=== degraded query plan --section phase (post-R1: only unknown slugs degrade) ===")
for d in (True,False):
    f,tr = work_full(degraded=d)
    ti,to_,n,ce = cost(f,tr)
    print(f"  degraded={d!s:<6} in={ti:>9,.0f}  out={to_:>7,.0f}  total={ti+to_:>9,.0f}")

print()
print("=== work full: scaling in tasks (the inner loop) ===")
for n in (3,5,8,12):
    f,tr = work_full(tasks=n)
    ti,to_,nt,ce = cost(f,tr)
    print(f"  {n:>2} tasks: {nt:>3} turns  in={ti:>9,.0f}  out={to_:>7,.0f}  total={ti+to_:>9,.0f}  per-task={(ti+to_)/n:>8,.0f}")

print()
print("=== work full: ceremony vs productive turns (5 tasks) ===")
CEREMONY = ('trace','edit plan','current state','query','preflight','run create','progress entry')
f,tr = work_full()
def is_cer(n): return any(k in n for k in CEREMONY)
ctx = BASELINE + f; cer_in=prod_in=cer_out=prod_out=0; nc=npd=0
for name,out,res in tr:
    if is_cer(name): cer_in+=ctx; cer_out+=out; nc+=1
    else: prod_in+=ctx; prod_out+=out; npd+=1
    ctx += out+res
print(f"  ceremony   turns={nc:>3}  in={t(cer_in):>9,.0f}  out={t(cer_out):>6,.0f}  total={t(cer_in+cer_out):>9,.0f}")
print(f"  productive turns={npd:>3}  in={t(prod_in):>9,.0f}  out={t(prod_out):>6,.0f}  total={t(prod_in+prod_out):>9,.0f}")
print(f"  ceremony share of work-stage cost: {(cer_in+cer_out)/(cer_in+cer_out+prod_in+prod_out)*100:.1f}%")

print()
print("=== trace-add cadence: pre-R5 per-task vs post-R5 batched ===")
def work_per_task(tasks=5, waves=2):
    """Reconstruct the pre-R5 ledger from the batched one: one `trace add`
    per task, one wave summary per wave, no batched flush."""
    fixed, turns = work_full(tasks, waves)
    new = []
    for name, out, res in turns:
        if name.startswith('t') and 'verify' in name:
            new.append((name, out, res))
            new.append((name.replace('verify', 'trace add'), 290, CLI['trace_add']))
        elif 'trace flush' in name:
            continue
        elif 'trace summary' in name:
            new.append(('wave trace add', 190, CLI['trace_add']))
        else:
            new.append((name, out, res))
    return fixed, new
for label, fn in (("per-task (pre-R5)", work_per_task), ("batched (current)", work_full)):
    f,tr = fn(); ti,to_,n,ce = cost(f,tr)
    print(f"  {label:<20} turns={n:>3}  in={ti:>9,.0f}  out={to_:>6,.0f}  total={ti+to_:>9,.0f}")

print()
print("=== check gate vs check full: cost driver breakdown ===")
for label, fn in (("gate", check_gate), ("full", check_full)):
    f,tr = fn()
    ctx = BASELINE + f
    ti,to_,n,ce = cost(f,tr)
    print(f"  [{label}] turns={n}  in={ti:,.0f}  out={to_:,.0f}  total={ti+to_:,.0f}")
    for name,out,res in tr:
        print(f"    {name:<30} ctx_in={t(ctx):>8,.0f}  out={t(out):>6,.0f}  res={t(res):>7,.0f}")
        ctx += out+res

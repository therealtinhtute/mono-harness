#!/usr/bin/env python3
"""SDLC token model. All char constants are MEASURED from the dogfood run
in scratchpad/dog + this repo's own files. Chars->tokens at 3.5 (markdown)
is an estimate; ratios between stages are unaffected by the divisor."""

CH = 3.5
def t(chars): return chars / CH

# ---- measured fixed loads (chars) ----
SKILL = dict(watzup=1135, brainstorm=1398, to_plan=1208, work=1377,
             check=1341, git=2768, handoff=1104)
PLAYBOOK = dict(watzup=4424, brainstorm=5657, to_plan=5316, work=9018,
                check=9149, git=7330, handoff=7338)
GIT_REFS = 1851 + 2517          # git/SKILL.md still loads its references/
BASELINE = 12000                # system prompt + tool defs, replayed every turn

# ---- measured CLI JSON output (chars) ----
CLI = dict(preflight=659, preflight_ctx=1688, query_plan_phase_ok=1290,
           query_plan_phase_degraded=5863, query_plan_current=379,
           query_traces=972, query_phases=356, resume=150, audit=67,
           trace_add=36, run_create=40, check_record=60, handoff_record=60,
           story=70)

# ---- measured artifact sizes (chars) ----
PLAN_START = 5586               # 3-phase filled plan
GATE_OUTPUT = 485 + 26          # this repo's real gate, passing

# (name, agent_out_chars, tool_result_chars) per turn
def work_full(tasks=5, waves=2, degraded=True):
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
                  (f"t{i} verify cmd", 130, 700),
                  (f"t{i} trace add", 290, CLI['trace_add'])]
    for w in range(waves):
        turns += [(f"wave{w} trace add", 190, CLI['trace_add'])]
    turns += [("refresh current state", 900, 200),
              ("query phases", 100, CLI['query_phases'])]
    return SKILL['work'] + PLAYBOOK['work'], turns

def check_full(files=6, degraded=True):
    turns  = [("preflight", 120, CLI['preflight'])]
    turns += [("git diff", 90, 9000)]              # the phase diff
    turns += [("resume", 100, CLI['resume'])]
    turns += [("read active plan", 110, PLAN_START + 1200)]
    # automated gate: this repo runs 2 commands; typical repo 3-4
    for c in range(4):
        turns += [(f"gate cmd {c}", 120, GATE_OUTPUT if c else 3000)]
    turns += [("read evals/failures.md", 110, 2200)]
    # the complete Security/Performance/Architecture/Code-Quality review:
    # every changed file read in full, plus sibling search for class-of-bug
    for f in range(files):
        turns += [(f"review read file {f}", 110, 5200)]
    turns += [("grep sibling instances", 130, 2400),
              ("grep sibling instances 2", 130, 1900)]
    turns += [("audit", 100, CLI['audit'])]
    turns += [("check record", 900, CLI['check_record'])]   # long proof-links JSON
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
          ("check full (6 files)", check_full()), ("git", git_stage()),
          ("handoff", handoff())]

# opus is ~5x sonnet on both in and out; haiku ~1/3 sonnet
MODEL = dict(watzup=("haiku",0.33), brainstorm=("opus",5.0), to_plan=("opus",5.0),
             work=("sonnet",1.0), check=("opus",5.0), git=("sonnet",1.0),
             handoff=("sonnet",1.0))
key = {"watzup":"watzup","brainstorm":"brainstorm","to-plan":"to_plan",
       "work full (5 tasks)":"work","check full (6 files)":"check",
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
print()
print("share of raw in+out:")
for r in sorted(rows,key=lambda r:-r[4]): print(f"  {r[0]:<22}{r[4]/tot_raw*100:>6.1f}%")
print("share of model-weighted:")
for r in sorted(rows,key=lambda r:-r[7]): print(f"  {r[0]:<22}{r[7]/tot_w*100:>6.1f}%")

print()
print("=== degraded query plan --section phase: cost of the broken P3 path ===")
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
CEREMONY = ('trace add','edit plan','current state','query','preflight','run create','progress entry')
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
print("=== counterfactual: batch trace add per wave instead of per task ===")
def work_batched(tasks=5, waves=2):
    fixed, turns = work_full(tasks, waves)
    turns = [x for x in turns if 'trace add' not in x[0] or 'wave' in x[0]]
    # wave traces now carry all their tasks: bigger single call
    turns = [(n, 900 if 'wave' in n and 'trace' in n else o, r) for n,o,r in turns]
    return fixed, turns
for label, fn in (("per-task (current)", work_full), ("per-wave (batched)", work_batched)):
    f,tr = fn(); ti,to_,n,ce = cost(f,tr)
    print(f"  {label:<20} turns={n:>3}  in={ti:>9,.0f}  out={to_:>6,.0f}  total={ti+to_:>9,.0f}")

print()
print("=== check full: cost driver breakdown ===")
f,tr = check_full()
ctx = BASELINE + f
for name,out,res in tr:
    print(f"  {name:<30} ctx_in={t(ctx):>8,.0f}  out={t(out):>6,.0f}  res={t(res):>7,.0f}")
    ctx += out+res

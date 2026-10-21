#!/usr/bin/env bash
# Observational replay scorer for the lean-v1 cost metrics
# (docs/evals/lean-workflow.md, "Metrics"). Reads Claude Code session
# transcripts, keeps sessions that invoked a core workflow skill, and splits
# them into B-era / C-era at the C install cutoff. Free: reads local files only.
# Usage: bash scripts/lean-replay-score.sh [projects-dir] [cutoff-iso]
# Prints one TSV row per session, then per-era medians on stderr.
set -euo pipefail
DIR="${1:-$HOME/.claude/projects}"
CUTOFF="${2:-2026-10-21T07:38:29Z}"

python3 -I - "$DIR" "$CUTOFF" <<'PY'
import glob, json, os, statistics, sys
root, cutoff = sys.argv[1], sys.argv[2]
CORE = {"brainstorm", "to-plan", "work", "check", "handoff", "watzup"}
INSTR = ("AGENTS.md", "WORKFLOW.md", "/skills/")
cols = ["era", "session", "project", "skills", "clarifications", "instr_bytes_read",
        "redundant_checks", "elapsed_s", "tokens_in", "tokens_out", "cost_usd"]
rows = []
for f in sorted(glob.glob(os.path.join(root, "*", "*.jsonl"))):
    ev = []
    for line in open(f, encoding="utf-8", errors="replace"):
        try: ev.append(json.loads(line))
        except Exception: pass
    ts = sorted(e["timestamp"] for e in ev if e.get("type") in ("user", "assistant") and e.get("timestamp"))
    if not ts: continue
    skills = set(); clar = instr = redundant = tin = tout = 0
    cost = None; pending = {}; last_pass = {}; seen_msg = set()
    for e in ev:
        t = e.get("type")
        if t == "cost-state": cost = e.get("totalCostUSD")
        content = (e.get("message") or {}).get("content")
        if t == "user" and isinstance(content, str):
            for s in CORE:
                if f"<command-name>/{s}</command-name>" in content: skills.add(s)
        if not isinstance(content, list): continue
        if t == "assistant":
            m = e["message"]
            if m.get("id") not in seen_msg:  # one API message spans several events
                seen_msg.add(m.get("id")); u = m.get("usage") or {}
                tin += sum(u.get(k) or 0 for k in ("input_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"))
                tout += u.get("output_tokens") or 0
            for c in content:
                if c.get("type") != "tool_use": continue
                n, inp = c.get("name"), c.get("input") or {}
                if n == "Skill" and inp.get("skill", "").split(":")[-1] in CORE: skills.add(inp["skill"].split(":")[-1])
                if n == "AskUserQuestion": clar += 1
                if n == "Read" and any(k in inp.get("file_path", "") for k in INSTR): pending[c["id"]] = ("read", None)
                if n in ("Edit", "Write", "NotebookEdit"): last_pass.clear()
                if n == "Bash": pending[c["id"]] = ("bash", inp.get("command", ""))
        elif t == "user":
            for c in content:
                if not isinstance(c, dict) or c.get("type") != "tool_result" or c.get("tool_use_id") not in pending: continue
                kind, cmd = pending.pop(c["tool_use_id"])
                body = c.get("content")
                text = body if isinstance(body, str) else "".join(x.get("text", "") for x in body or [] if isinstance(x, dict))
                if kind == "read" and not c.get("is_error"):
                    instr += len(text.encode())  # bytes as returned to the model, numbered lines included
                elif kind == "bash" and not c.get("is_error"):
                    if cmd in last_pass: redundant += 1
                    last_pass[cmd] = True
    if not skills: continue
    from datetime import datetime
    p = lambda s: datetime.fromisoformat(s.replace("Z", "+00:00"))
    rows.append({"era": "B" if ts[0] < cutoff else "C", "session": os.path.basename(f)[:8],
                 "project": os.path.basename(os.path.dirname(f))[-30:], "skills": ",".join(sorted(skills)),
                 "clarifications": clar, "instr_bytes_read": instr, "redundant_checks": redundant,
                 "elapsed_s": int((p(ts[-1]) - p(ts[0])).total_seconds()),
                 "tokens_in": tin, "tokens_out": tout,
                 "cost_usd": round(float(cost), 4) if cost is not None else "null:no cost-state"})
print("\t".join(cols))
for r in rows: print("\t".join(str(r[c]) for c in cols))
for era in ("B", "C"):
    rs = [r for r in rows if r["era"] == era]
    if not rs: print(f"{era}: n=0", file=sys.stderr); continue
    med = {c: statistics.median(r[c] for r in rs if isinstance(r[c], (int, float)))
           for c in cols[4:] if any(isinstance(r[c], (int, float)) for r in rs)}
    print(f"{era}: n={len(rs)} " + " ".join(f"{k}={v:g}" for k, v in med.items()), file=sys.stderr)
PY

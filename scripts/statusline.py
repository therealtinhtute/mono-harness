#!/usr/bin/env python3
"""Slim statusline for Claude Code — session, weekly, context, heartbeat, branch."""

import json
import os
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

HOOK_STATE_FILE = Path.home() / ".claude" / "claude-code-statusline-hook.json"
HOOK_FRESH_TTL = 300  # 5 minutes — keeps heartbeat visible between interactions

FILL = "━"
EMPTY = "─"
BAR_LEN = 8

RESET = "\033[0m"
DIM = "\033[38;5;247m"
WHITE = "\033[38;5;252m"
CYAN = "\033[36m"


def _color(pct):
    if pct >= 75:
        return "\033[38;5;208m"   # orange
    if pct >= 50:
        return "\033[93m"          # yellow
    if pct >= 35:
        return "\033[38;5;148m"   # yellow-green
    return "\033[38;5;82m"        # green


def _bar(pct):
    filled = min(BAR_LEN, round(float(pct) / 100 * BAR_LEN))
    return FILL * filled + EMPTY * (BAR_LEN - filled)


def _hm(seconds):
    if seconds <= 0:
        return ""
    h, rem = divmod(int(seconds), 3600)
    m = rem // 60
    return f"{h}h {m}m" if h else f"{m}m"


def _weekly_reset(epoch):
    if not epoch:
        return ""
    try:
        dt = datetime.fromtimestamp(float(epoch))
        hour = dt.hour
        ampm = "am" if hour < 12 else "pm"
        h12 = hour % 12 or 12
        return f"R:{dt.strftime('%a')} {h12}{ampm}"
    except (ValueError, OSError):
        return ""


def _hook_refresh():
    """PostToolUse handler — increments tool count in state file, no stdout."""
    tool_name = "unknown"
    if not sys.stdin.isatty():
        try:
            raw = sys.stdin.read(65536)
            if raw.strip():
                tool_name = json.loads(raw).get("tool_name", tool_name)
        except (json.JSONDecodeError, OSError):
            pass

    now = time.time()
    state = {}
    try:
        with open(HOOK_STATE_FILE) as f:
            state = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError, OSError):
        pass

    state["tool_count"] = state.get("tool_count", 0) + 1
    state.setdefault("session_start", now)
    state["last_refresh"] = now
    state["last_tool"] = str(tool_name)[:50]

    try:
        tmp = Path(str(HOOK_STATE_FILE) + ".tmp")
        tmp.write_text(json.dumps(state))
        tmp.replace(HOOK_STATE_FILE)
    except OSError:
        pass


def main():
    # stdin JSON from Claude Code
    data = {}
    try:
        raw = sys.stdin.read()
        if raw.strip():
            parsed = json.loads(raw)
            data = parsed.get("data", parsed)
    except (json.JSONDecodeError, OSError):
        pass

    now = time.time()

    # Model
    try:
        model = str(data.get("model", {}).get("display_name", "") or data.get("model", "") or "unknown")
    except AttributeError:
        model = "unknown"

    # Context %
    try:
        ctx_pct = float(data.get("context_window", {}).get("used_percentage", 0) or 0)
    except (ValueError, TypeError, AttributeError):
        ctx_pct = 0.0

    # TPM
    tpm = 0
    try:
        ctx = data.get("context_window", {}) or {}
        total_tokens = int(ctx.get("total_input_tokens", 0) or 0) + int(ctx.get("total_output_tokens", 0) or 0)
        duration_ms = int((data.get("cost", {}) or {}).get("total_duration_ms", 0) or 0)
        if duration_ms > 0:
            tpm = (total_tokens * 60000) // duration_ms
    except (ValueError, TypeError, AttributeError):
        pass

    # Rate limits (requires Claude Code v2.1.80+)
    rl = (data.get("rate_limits") or {})
    five = rl.get("five_hour") or {}
    seven = rl.get("seven_day") or {}

    try:
        session_pct = float(five.get("used_percentage", 0) or 0)
        session_reset_at = five.get("resets_at")
    except (ValueError, TypeError):
        session_pct, session_reset_at = 0.0, None

    try:
        weekly_pct = float(seven.get("used_percentage", 0) or 0)
        weekly_reset_at = seven.get("resets_at")
    except (ValueError, TypeError):
        weekly_pct, weekly_reset_at = 0.0, None

    # Git branch
    branch = ""
    try:
        cwd = data.get("cwd") or os.getcwd()
        r = subprocess.run(
            ["git", "--no-optional-locks", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True, text=True, timeout=2,
        )
        b = r.stdout.strip()
        if b == "HEAD":
            r2 = subprocess.run(
                ["git", "--no-optional-locks", "-C", cwd, "rev-parse", "--short", "HEAD"],
                capture_output=True, text=True, timeout=2,
            )
            b = r2.stdout.strip()
        branch = b
    except (OSError, subprocess.TimeoutExpired):
        pass

    # Heartbeat (from PostToolUse hook state)
    tool_count = 0
    session_elapsed = 0.0
    hook_fresh = False
    try:
        with open(HOOK_STATE_FILE) as f:
            hs = json.load(f)
        if now - hs.get("last_refresh", 0) < HOOK_FRESH_TTL:
            hook_fresh = True
            tool_count = hs.get("tool_count", 0)
            session_elapsed = now - hs.get("session_start", now)
    except (FileNotFoundError, json.JSONDecodeError, OSError):
        pass

    spinner = r"\|/-"[int(now) % 4]

    # Assemble parts
    SEP = "  "
    parts = []

    parts.append(f"{WHITE}✦ {model}{RESET}")

    if five:
        c = _color(session_pct)
        time_left = ""
        if session_reset_at:
            try:
                time_left = f" {_hm(float(session_reset_at) - now)}"
            except (ValueError, TypeError):
                pass
        parts.append(f"{DIM}Session{RESET} {c}{_bar(session_pct)} {int(session_pct)}%{time_left}{RESET}")

    if seven:
        c = _color(weekly_pct)
        reset_label = f" {_weekly_reset(weekly_reset_at)}" if weekly_reset_at else ""
        parts.append(f"{DIM}Weekly{RESET} {c}{_bar(weekly_pct)} {int(weekly_pct)}%{reset_label}{RESET}")

    c = _color(ctx_pct)
    parts.append(f"{DIM}Context{RESET} {c}{_bar(ctx_pct)} {int(ctx_pct)}%{RESET}")

    if tpm > 0:
        tpm_s = f"{tpm // 1000}.{tpm % 1000 // 100}k" if 1000 <= tpm < 10000 else (f"{tpm // 1000}k" if tpm >= 10000 else str(tpm))
        parts.append(f"{DIM}ϟ {tpm_s} tpm{RESET}")

    if hook_fresh:
        elapsed = f" {_hm(session_elapsed)}" if session_elapsed > 60 else ""
        parts.append(f"{DIM}[{spinner}] {tool_count} tools{elapsed}{RESET}")

    if branch:
        parts.append(f"{CYAN}⌥ {branch}{RESET}")

    print(SEP.join(parts))


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--hook-refresh":
        _hook_refresh()
    else:
        main()

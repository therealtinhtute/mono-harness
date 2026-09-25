#!/usr/bin/env python3
"""Condense a Claude Code session log into the signals a retro needs.

Usage: session-digest.py [SESSION_JSONL | SESSION_ID]
With no argument, reads the newest session for the current directory under
~/.claude/projects/<cwd-slug>/. Prints plain text; reads nothing else.
"""
import collections
import glob
import json
import os
import re
import sys

PROJECTS = os.path.expanduser("~/.claude/projects")


def resolve(arg):
    if arg and os.path.isfile(arg):
        return arg
    slug = re.sub(r"[^A-Za-z0-9]", "-", os.getcwd())
    if arg:
        hits = glob.glob(os.path.join(PROJECTS, "*", arg + ".jsonl"))
        if hits:
            return hits[0]
        sys.exit(f"session not found: {arg}")
    logs = glob.glob(os.path.join(PROJECTS, slug, "*.jsonl"))
    if not logs:
        sys.exit(f"no sessions under {os.path.join(PROJECTS, slug)}; pass a path")
    return max(logs, key=os.path.getmtime)


def text_of(content):
    if isinstance(content, str):
        return content
    parts = []
    for block in content or []:
        if block.get("type") == "text":
            parts.append(block.get("text", ""))
        elif block.get("type") == "tool_result":
            parts.append(text_of(block.get("content")))
    return "\n".join(parts)


def short(s, n=160):
    s = " ".join(str(s).split())
    return s if len(s) <= n else s[: n - 1] + "…"


def main():
    path = resolve(sys.argv[1] if len(sys.argv) > 1 else None)
    calls = {}
    counts = collections.Counter()
    repeats = collections.Counter()
    errors, results, prompts = [], [], []

    with open(path, encoding="utf-8") as f:
        for line in f:
            try:
                rec = json.loads(line)
            except ValueError:
                continue
            if rec.get("isSidechain"):
                continue
            msg = rec.get("message") or {}
            content = msg.get("content")
            if rec.get("type") == "assistant" and isinstance(content, list):
                for b in content:
                    if b.get("type") == "tool_use":
                        sig = short(json.dumps(b.get("input"), sort_keys=True), 200)
                        calls[b["id"]] = (b["name"], sig)
                        counts[b["name"]] += 1
                        repeats[(b["name"], sig)] += 1
            elif rec.get("type") == "user":
                if isinstance(content, str):
                    if not content.startswith("<"):
                        prompts.append(short(content, 240))
                    continue
                for b in content or []:
                    if b.get("type") != "tool_result":
                        continue
                    name, sig = calls.get(b.get("tool_use_id"), ("?", ""))
                    body = text_of(b.get("content"))
                    results.append((len(body), name, sig))
                    if b.get("is_error"):
                        errors.append((name, short(body)))

    print(f"session: {path}")
    print(f"user prompts: {len(prompts)}  tool calls: {sum(counts.values())}")
    print("\n## tool calls by name")
    for name, n in counts.most_common():
        print(f"- {name}: {n}")
    print(f"\n## errored tool results ({len(errors)})")
    for name, body in errors:
        print(f"- {name}: {body}")
    print("\n## repeated identical calls")
    for (name, sig), n in repeats.most_common():
        if n < 2:
            break
        print(f"- {name} x{n}: {sig}")
    print("\n## largest tool results (chars)")
    for size, name, sig in sorted(results, reverse=True)[:5]:
        print(f"- {size} {name}: {sig}")
    print("\n## user prompts (corrections and steering show up here)")
    for i, p in enumerate(prompts, 1):
        print(f"{i}. {p}")


if __name__ == "__main__":
    main()

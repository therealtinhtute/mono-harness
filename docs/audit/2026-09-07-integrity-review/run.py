#!/usr/bin/env python3
"""Opt-in contract probes. Exit 1 means regressions, NOT a passing product suite.

Uses temporary directories only. Reads source from --source-root (default: repo
root), extracts selected original Go functions and the shared shell guard core,
and runs controlled fixtures. It never installs repository hooks or runs proofs
from real plans. Requires Python 3.10+, Go 1.23+, bash, git and coreutils.
"""
from __future__ import annotations
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
BASE = "a45642227ea26ec72ea111ddb2fefd626e3026ab"
SELECT = {
    "installer.go": ["safePath", "legacySafePath", "findOriginal", "writeFileAtomic", "captureOriginal", "readOriginal", "agentsSpan"],
    "update.go": ["stashCapture", "stashRestore"],
    "uninstall.go": ["readAll", "dropLines", "bytesEqual", "dropLine", "removeManagedFile", "removeAgentsBlock", "removeDirIfEmpty", "isSame"],
}

def function(source: str, name: str) -> str:
    # Selected functions have ordinary, column-zero Go closing braces and no
    # multiline raw-string bodies. Refuse to guess if the source shape changes.
    m = re.search(r"(?m)^func " + re.escape(name) + r"\(", source)
    if not m:
        raise ValueError(f"missing original function {name}")
    end = source.find("\n", m.start())
    if source[m.start():end].rstrip().endswith("}"):
        return source[m.start():end] + "\n"
    close = re.search(r"(?m)^}\s*$", source[end:])
    if not close:
        raise ValueError(f"missing end of original function {name}")
    return source[m.start():end + close.end()].rstrip() + "\n"

def run(cmd: list[str], cwd: Path, env: dict[str, str], timeout: int = 90) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, cwd=cwd, env=env, text=True, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, timeout=timeout, check=False)

def required(cmd: list[str], cwd: Path, env: dict[str, str]) -> str:
    p = run(cmd, cwd, env)
    if p.returncode:
        raise RuntimeError(f"command failed ({p.returncode}): {cmd!r}\n{p.stdout}\n{p.stderr}")
    return p.stdout

def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--source-root", type=Path, default=HERE.parents[2])
    ap.add_argument("--output", type=Path, help="Optional JSON evidence destination")
    ap.add_argument("--source-excerpts", action="store_true", help="Explicitly identify a partial, connector-retrieved source workspace")
    args = ap.parse_args()
    root = args.source_root.resolve()
    for tool in ["go", "bash", "git", "awk", "grep", "sed", "sha256sum"]:
        if shutil.which(tool) is None:
            raise RuntimeError(f"missing prerequisite: {tool}")
    records: list[dict] = []
    input_blobs: dict[str, str] = {}
    for rel in ["cli/internal/installer/installer.go", "cli/internal/installer/update.go", "cli/internal/installer/uninstall.go", "scripts/install-git-hooks.sh", ".github/workflows/cli-ci.yml"]:
        data = (root / rel).read_bytes()
        input_blobs[rel] = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
    hashes: dict[str, str] = {}
    with tempfile.TemporaryDirectory(prefix="mono-harness-audit-") as td:
        tmp = Path(td)
        env = {"PATH": os.environ["PATH"], "HOME": str(tmp), "TMPDIR": str(tmp),
               "LC_ALL": "C.UTF-8", "GO111MODULE": "off", "GOTOOLCHAIN": "local",
               "GOPROXY": "off", "GOSUMDB": "off", "CGO_ENABLED": "0",
               "GOCACHE": str(Path(tempfile.gettempdir()) / "mono-harness-audit-go-cache"),
               "GIT_CONFIG_NOSYSTEM": "1", "GIT_TERMINAL_PROMPT": "0"}
        # No inherited auth tokens, git credentials, shell aliases or functions.
        funcs = []
        for file, names in SELECT.items():
            text = (root / "cli/internal/installer" / file).read_text()
            for name in names:
                body = function(text, name)
                hashes[f"{file}:{name}"] = hashlib.sha256(body.encode()).hexdigest()
                funcs.append(body)
        header = '''package installer
import ("crypto/sha256"; "encoding/hex"; "fmt"; "os"; "path/filepath"; "strings")
const (
 originalDir = ".zharness/base/original"
 stashDir = ".zharness/update-stash"
 agentsTarget = "AGENTS.md"
 blockBegin = "<!-- ZHARNESS:BEGIN -->"
 blockEnd = "<!-- ZHARNESS:END -->"
)
'''
        pkg = tmp / "go-probes"; pkg.mkdir()
        (pkg / "original.go").write_text(header + "\n".join(funcs))
        (pkg / "audit_test.go").write_text((HERE / "installer_contract_test.go.txt").read_text())
        go = run(["go", "test", "-json", "-count=1", "."], pkg, env)
        go_events = []
        for line in go.stdout.splitlines():
            try: go_events.append(json.loads(line))
            except json.JSONDecodeError: pass
        for event in go_events:
            if event.get("Test") and event.get("Action") in ("pass", "fail"):
                name = event["Test"]
                output = "".join(e.get("Output", "") for e in go_events if e.get("Test") == name)
                records.append({"test": name, "status": "pass" if event["Action"] == "pass" else "regression", "output": output})
        if not any(r["test"].startswith("TestAudit_") for r in records):
            raise RuntimeError(f"Go probes did not execute:\n{go.stdout}\n{go.stderr}")

        hook = (root / "scripts/install-git-hooks.sh").read_text()
        ci = (root / ".github/workflows/cli-ci.yml").read_text()
        core = hook.split("# ZGUARD-CORE-BEGIN\n", 1)[1].split("# ZGUARD-CORE-END", 1)[0]
        core_path = tmp / "guard.sh"; core_path.write_text(core)
        hashes["guard-core"] = hashlib.sha256(core.encode()).hexdigest()
        hook_select = "git diff --cached --name-only --diff-filter=ACM -- 'docs/plans/active/*.md'"
        ci_select = "git diff --name-only --diff-filter=ACM HEAD~1 HEAD -- 'docs/plans/active/*.md'"
        # Selectors are pinned to the reviewed version, not silently simulated on
        # changed code. A maintainer changing these must update these probes.
        for needle, text in [(hook_select, hook), (ci_select, ci)]:
            if needle not in text:
                raise ValueError("source selector changed; inspect and update probe: " + needle)

        def shell(code: str, cwd: Path) -> subprocess.CompletedProcess:
            return run(["bash", "-c", 'source "$1"; ' + code, "audit", str(core_path)], cwd, env, 30)
        def record(name: str, p: subprocess.CompletedProcess, expected: int) -> None:
            ok = (p.returncode == 0) == (expected == 0)
            records.append({"test": name, "status": "pass" if ok else "regression",
                            "expected_exit": "0" if expected == 0 else "nonzero",
                            "actual_exit": p.returncode, "output": p.stdout + p.stderr})
        def repo(name: str) -> Path:
            path = tmp / name; path.mkdir()
            required(["git", "init", "-q", "."], path, env)
            required(["git", "config", "user.email", "audit@example.invalid"], path, env)
            required(["git", "config", "user.name", "Local audit fixture"], path, env)
            (path / "README.md").write_text("fixture\n")
            commit(path)
            return path
        def commit(path: Path) -> None:
            required(["git", "add", "."], path, env)
            required(["git", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "test: fixture"], path, env)
        def plan(proof: str = "false", judge: str = "independent", status: str = "done") -> str:
            return ("---\nlane: high-risk\n---\n## Phases and Verification\n"
                    f"phase_slug: test\nstatus: {status}\n"
                    "## Current State and Next Action\nlifecycle_status: completed\n"
                    "## Validation\n- 2026-09-07 mode: full; verdict: APPROVED"
                    + (f"; judge: {judge}" if judge else "") + f"\n  - `{proof}`\n")
        def put(path: Path, rel: str, text: str) -> None:
            p = path / rel; p.parent.mkdir(parents=True, exist_ok=True); p.write_text(text)

        r = repo("completed-add")
        put(r, "docs/plans/completed/test.md", plan())
        required(["git", "add", "."], r, env)
        p = shell('plans=$(' + hook_select + '); echo "proof-selected=$plans"; '
                  'if [ -n "$plans" ]; then mkdir -p audit-tmp; zhuards_guard_plans "$plans" staged audit-tmp || exit 1; fi; '
                  'zharness_guard_completed_plan_phases_done docs/plans/completed/test.md docs/plans/completed/test.md', r)
        record("F02_completed_addition_requires_passing_proof", p, 1)
        # Positive control: the very same content is rejected by the entry guard.
        (r / "empty.md").write_text("")
        record("control_false_proof_is_rejected", shell('zharness_guard_entries_of_file p empty.md docs/plans/completed/test.md', r), 1)
        put(r, "bad-phase.md", plan(status="checked"))
        record("control_incomplete_phase_is_rejected", shell('zharness_guard_completed_plan_phases_done p bad-phase.md', r), 1)

        r = repo("completed-rename")
        put(r, "docs/plans/active/test.md", plan(proof="true")); commit(r)
        (r / "docs/plans/completed").mkdir()
        required(["git", "mv", "docs/plans/active/test.md", "docs/plans/completed/test.md"], r, env)
        put(r, "docs/plans/completed/test.md", plan())
        required(["git", "add", "."], r, env)
        record("F02_completed_rename_requires_passing_proof", shell(
            'plans=$(' + hook_select + '); echo "proof-selected=$plans"; '
            'if [ -n "$plans" ]; then mkdir -p audit-tmp; zhuards_guard_plans "$plans" staged audit-tmp || exit 1; fi; '
            'zharness_guard_completed_plan_phases_done p docs/plans/completed/test.md', r), 1)

        r = repo("push-range")
        base = required(["git", "rev-parse", "HEAD"], r, env).strip()
        put(r, "docs/plans/active/test.md", plan()); commit(r)
        put(r, "README.md", "unrelated final commit\n"); commit(r)
        record("F03_batch_push_must_check_earlier_plan_change", shell(
            'plans=$(' + ci_select + '); echo "tip-selected=$plans"; '
            'if [ -z "$plans" ]; then exit 0; fi; mkdir -p audit-tmp; zhuards_guard_plans "$plans" head audit-tmp', r), 1)
        full_range = required(["git", "diff", "--name-only", base, "HEAD", "--", "docs/plans/active/*.md"], r, env)
        if "docs/plans/active/test.md" not in full_range: raise RuntimeError("range fixture invalid")
        (r / "empty.md").write_text("")
        record("control_earlier_push_proof_would_fail", shell('zharness_guard_entries_of_file p empty.md docs/plans/active/test.md', r), 1)

        r = repo("index-state")
        put(r, "docs/plans/active/a.md", "a\n"); put(r, "docs/plans/active/b.md", "b\n")
        required(["git", "add", "."], r, env)
        record("control_two_worktree_plans_rejected", shell('zharness_guard_at_most_one_active_plan "$PWD"', r), 1)
        put(r, "docs/plans/active/b.md", "")
        staged = required(["git", "show", ":docs/plans/active/b.md"], r, env)
        if not staged.strip(): raise RuntimeError("index fixture invalid")
        record("F08_two_staged_plans_must_be_rejected", shell('zharness_guard_at_most_one_active_plan "$PWD"', r), 1)

        r = repo("schema-limits")
        (r / "empty.md").write_text("")
        put(r, "new.md", plan(proof="true", judge="same-session"))
        record("control_same_session_full_judge_rejected", shell('zharness_guard_entries_of_file p empty.md new.md', r), 1)
        put(r, "new.md", plan(proof="true", judge="independent"))
        record("control_independent_full_true_proof_accepted", shell('zharness_guard_entries_of_file p empty.md new.md', r), 0)
        put(r, "new.md", plan(proof="true", judge=""))
        record("L01_missing_judge_requires_schema_rejection", shell('zharness_guard_entries_of_file p empty.md new.md', r), 1)
        put(r, "new.md", "## Validation\n- 2026-09-07 review\n  - verdict: APPROVED\n  - `false`\n")
        record("L02_unparseable_verdict_requires_schema_rejection", shell('zharness_guard_entries_of_file p empty.md new.md', r), 1)

        result = {"audited_commit": BASE, "source_mode": "connector-retrieved source excerpts" if args.source_excerpts else "source checkout (focused functions only)",
                  "execution_scope": "selected original functions and guard-core, not complete CLI",
                  "input_git_blob_ids": input_blobs,
                  "go_version": required(["go", "version"], tmp, env).strip(),
                  "python_version": sys.version.split()[0],
                  "platform": sys.platform,
                  "git_version": required(["git", "--version"], tmp, env).strip(),
                  "source_function_sha256": hashes, "results": records,
                  "summary": {"checks": len(records), "pass": sum(r["status"] == "pass" for r in records),
                              "regressions": sum(r["status"] == "regression" for r in records)}}
        for row in records: print(f'{row["status"].upper():10} {row["test"]}')
        print(json.dumps(result["summary"]))
        if args.output:
            args.output.parent.mkdir(parents=True, exist_ok=True)
            args.output.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n")
        if any(r["status"] != "pass" and ("_Control_" in r["test"] or r["test"].startswith("control_")) for r in records):
            return 2
        return 1 if result["summary"]["regressions"] else 0

if __name__ == "__main__":
    try: sys.exit(main())
    except (OSError, ValueError, RuntimeError, subprocess.TimeoutExpired) as exc:
        print(f"HARNESS ERROR: {exc}", file=sys.stderr)
        sys.exit(2)

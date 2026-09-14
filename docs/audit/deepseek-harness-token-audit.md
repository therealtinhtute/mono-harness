# Input-Token Audit — zharness on DeepSeek Harness

**Date:** 2026-09-14
**Case:** `baroiplayer`, initiative `reopen-verification-gates`.
**Status:** baseline verified from logs; context composition measured. The routing bug in S1 is fixed alongside this audit. Every other remedy below is a proposal to pilot and measure, not applied policy.
**Related:** [SDLC token + prompt-cache audit](sdlc-token-cache-audit.md) (F1: model-scoped caches, in-session phase gate).

## 1. Conclusion

**Root cause: many model calls over accumulated context that is never compacted, multiplied by sub-agents.** Tool results are the largest share of that context (about 46%), followed by reasoning (about 25%). zharness adds to history through plan updates, reports, and verification rounds. Playbook size was not shown to be a main cause in this case.

**Preferred direction: bound what each call carries — narrower tool output, phase-scoped context, selective delegation, and consolidated bookkeeping/verification that is still valid.** Keep required verification and independent review; reduce the history needed to perform them.

**Measure cost, not raw tokens.** 98.96% of input was cache-read. Compaction rewrites the prefix and adds summarization calls, so total prompt tokens can fall while the bill rises. Success is judged on the cost-weighted metric in §6.

No A/B benchmark proves the savings or an optimal configuration yet. Thresholds in this spec are pilot starting points, not current policy.

## 2. Baseline from real logs

Source: `$HOME/.dsh/sessions/--Users-tinhtute-Personal-baroiplayer--/`.
Main session: `session-94f6b9a0-f451-4137-bb9f-ee91e2c1515f`.
Main log read through event `seq=1738`, at `2026-09-14T13:55:17.071Z`.

Read the main `session.v3.jsonl.zstd` log and every child session reached by walking `subagent/catalog` events recursively: 17 children, all at depth 1, none missing. Usage summed from `assistant/message` matches `tokenUsage` in the projection cache for all 18 sessions. All have `inheritedEventCount=0`, so no usage inherited from a fork is double-counted.

| Scope | Calls with usage | Input incl. cache | Uncached input | Cache-read |
|---|---:|---:|---:|---:|
| Main agent | 266 | 52,345,835 | 191,467 | 52,154,368 |
| 17 sub-agents | 923 | 85,198,134 | 1,232,694 | 83,965,440 |
| **Total** | **1,189** | **137,543,969** | **1,424,161** | **136,119,808** |

- Cache-read is **98.96%** of total input; sub-agents are **61.94%** of total input.
- Total output: **1,042,182 tokens**; reasoning is part of output, not added again.
- **1,586 tool calls**, distinct from the number of model calls.
- From `/work full with sub-agents` at main event `seq=527` onward: **1,114 calls, 132,516,762 input tokens**, covering main after that point plus every child session. The assessment/to-plan part before it was 5,027,207 input tokens.
- Every model message records `deepseek-official / deepseek-flash`.

In the installed adapter, `inputTokens` already excludes cache hits. The correct sums:

```text
prompt_i = inputTokens_i + cacheReadTokens_i
session_input = sum(prompt_i)
tree_input = main_input + sum(child_input)
```

Do not add `totalTokens` to input. Do not price 137.54M tokens as fresh input; cost needs the cached-input, uncached-input, and output prices that applied to the account at run time, and this audit has no billing data.

## 3. Root cause and evidence

### RC1. Accumulated context resent on every call

Main grows from **14,160** to **387,383 tokens/request**, averaging **196,788.85**:

```text
266 × 196,788.85 ≈ 52,345,835 input tokens
```

No compaction or prune event appears in any of the 18 logs. The request context declares a **1,000,000-token** window. The installed `dsh-compaction-basic` backend defaults to `thresholdRatio=0.8` (**800,000 tokens**) and `retainRatio=0.16`. These are defaults read from source, not full proof of the running process's effective config; the absence of compaction is confirmed directly from the logs.

A capacity-based threshold protects the window; it does not keep the cumulative cost of a task reasonable. Main never reached 40% of the window yet summed over 52M input tokens.

### RC2. What the context is made of

Measured across all 18 sessions with the script in §7.2. Each log-visible content item is weighted by its character length times the number of later model calls in the same session that resend it, since no compaction ran.

| Source | Share of weighted chars |
|---|---:|
| **Tool results** — `bash` 23.2%, `read` 20.2%, `job_output` 1.7%, `edit` 0.6% | **≈ 45.7%** |
| Assistant reasoning | 25.4% |
| Assistant tool-call arguments | 13.4% |
| User messages (skill injections, sub-agent prompts) | 9.2% |
| Inbox splices | 3.5% |
| System messages | 1.6% |
| Assistant text | 0.9% |

- Only **118 of 1,586** tool results exceed 8,192 chars. Applying the installed pruner's defaults (`thresholdChars 8192`, `headChars 4096`, `tailChars 1024`) to them removes about **11.8%** of weighted chars. Most tool-result mass is many medium-sized outputs, not a few huge ones.
- Caveats: shares are characters, not tokens (504.7M weighted chars against 137.5M prompt tokens). The fixed system prompt and tool schemas sent with every call are not in the log and are excluded. `tool/call` events duplicate the assistant's tool-call items, so they are counted once.

### RC3. Sub-agent step counts too large for some narrow scopes

| Sub-agent | Calls | Cumulative input | Context first → last |
|---|---:|---:|---:|
| Fix stale API refs in docs | 113 | 16,647,692 | 15,106 → 210,105 |
| Fix volume slider focus visibility | 82 | 6,136,760 | 15,357 → 122,092 |
| Independent full review of initiative | 75 | 8,295,836 | 15,438 → 169,106 |
| Second independent full review | 62 | 7,121,443 | 15,896 → 175,091 |

Child sessions do not inherit the parent's full history. The cost comes from each child's own context growing across many rounds of reading, reasoning, editing, and checking. Not all 85.20M tokens are waste: the first independent review found four major findings.

### RC4. Reasoning, tool arguments, and bookkeeping stay in history

- Main produced **102,823 reasoning tokens**. The adapter serializes reasoning, text, and tool calls of each assistant turn into later requests.
- Main made **51 plan write/edit calls**, **134,685 argument characters** in total. Content written to files also lands in tool-call history.
- The plan snapshot is **94,685 characters**; Progress/Decisions/Validation are **65,694** of them, **69.4%** of the file.
- Gates ran at task/phase, end of implementation, independent review, and after review fixes. Some checks overlap: root typecheck plus package typecheck; the build wrapper reruns typecheck/lint/test.

Character counts show where context grows, not an exact share of the token bill. The number of plan updates alone does not prove every update was redundant.

### Hypotheses ruled out or unproven

- **Re-reading the whole plan repeatedly:** main actually read by section; no evidence this is a main cause.
- **Model switches breaking cache:** all model messages use the same model; cache-read reached 98.96%.
- **Small independent tasks:** the initiative has six phases, 32 tasks, 15 requirements, lane `high-risk`. Only some subtasks are small.
- **Only shortening work/check:** does not address call count, reasoning/history, or child context.
- **Nested sub-agents inflating the tree:** ruled out for this case; the recursive walk found depth 1 only.

## 4. Chosen approach (zharness)

| Option | Upside | Limit | When |
|---|---|---|---|
| Shorten playbooks only | Small change | Misses the main growth sources in the logs | Secondary improvement |
| **Bounded tool output + phase context + selective delegation** | Hits context length and call count; keeps independent review | Needs accurate checkpoints and A/B measurement | **Preferred pilot** |
| Put everything on one agent | Less delegation/reporting | Main keeps growing; loses division of work and independent review | Independent bounded tasks |

### S1. Choose the execution path before loading documents

Authority: [AGENTS.md](../../AGENTS.md) and [README.md](../../README.md) allow bounded changes without a plan.

**Bug found and fixed alongside this audit.** `skills/workflow/work/SKILL.md` defaulted to `mode:auto` and advertised `phase`, but the work playbook defined only `full` and `bounded`, and its precondition required an active plan in every mode — contradicting the bounded zero-write rule. Fix: [work.md](../playbooks/work.md) lines 9-15 now define `auto` (resolve to `full` only when the request names or continues a durable initiative, otherwise `bounded` when none of the bounded rejection conditions at line 19 apply, otherwise route to `brainstorm`/`to-plan`), scope the one-active-plan precondition to full mode, and the skill's argument hint lists `auto|full|bounded|simple`. `check` got the same fix: it defaulted to durable `full`, so a bare pre-commit `/check` after bounded work demanded an active plan; `auto` now resolves to `gate` only for an in-progress initiative phase, otherwise `bounded`, never `full`. Both stages print `mode: {resolved} ({reason})` before reading any plan. The embedded playbook copies are byte-identical.

Still open for routing:

- `check` `auto` selects bounded for a direct diff and reserves `full` for an explicit initiative review ([check.md](../playbooks/check.md) lines 5 and 10); `handoff` without a plan should return a recap only.
- Put bounded guidance first with a clear stop-reading point; load the full section only when selected.
- Subtasks of a durable initiative still update the plan through the orchestrator. Bounded fixes are not exempt from final review or requirement changes.

### S2. Compact context at stable points

- After a phase is checked and its state recorded consistently, write a short checkpoint in the existing Current State section; do not add a parallel state type.
- The checkpoint keeps the next phase/task, relevant authority/requirements, decisions to preserve, touched surfaces, proof and matching code state, blockers, and one exact next action.
- When context is large, compact at a completed tool-call boundary. On resume, check the checkpoint and the needed phase block; do not reload the whole history to "make up" for compaction.
- Compaction is a runtime operation. zharness describes the handoff/resume point; the `zharness` binary stays install/update/uninstall only.

Runtime thresholds for this live in §5, outside zharness.

### S3. Delegate by task value

- The orchestrator handles few-line changes with known proof itself. Use a sub-agent when the independent work is large enough or needs an independent reviewer.
- Hand over requirements, allowed/avoided surfaces, proof, and blockers. The child pulls more context on demand; do not make it read the full plan or workflow history.
- Reports back to the parent keep only result, changed surfaces, proof, findings, and limits. Expand detail only when a claim needs evaluating.
- **Diagnostic trigger is post-run analysis, not an agent instruction.** An agent cannot observe its own per-request token count, so a "20 calls / 80k tokens" rule inside a playbook adds prompt tokens without effect. During log review, flag any narrow task with more than **20 calls** or more than **80,000 input tokens/request** and explain it by progress or a concrete blocker.

### S4. Bound tool output and consolidate bookkeeping/verification

New lever (RC2 — tool results are the largest share):

- Locate before reading: `rg -l` / `rg -n` first, then read the needed line range instead of the whole file.
- Do not re-read a file already in context unless it changed since the read.
- Run tests and builds with quiet or failure-only reporters; keep full logs in a file and surface only the tail.

Already enforced by the playbooks — do not re-add:

- One Progress flush per wave, immediate flush on a blocker: [work.md](../playbooks/work.md) lines 46-48.
- Pass/fail output tails (3 lines on pass, 10 on fail) with the exit code preserved: [check.md](../playbooks/check.md) line 38.
- Complete `full` review exactly once, on the final phase: [check.md](../playbooks/check.md) line 5.
- Re-read the plan after compaction or summarization: [work.md](../playbooks/work.md) line 15, [check.md](../playbooks/check.md) line 15, [handoff.md](../playbooks/handoff.md) line 10.

Still proposed:

- Each Progress entry records result/proof rather than narrating the process. Validation keeps command, verdict, evidence, and auditable proof gaps; do not copy the same long report into Progress, Decisions, and Validation.
- Tasks run narrow proof; phases run the integration gate. Reuse proof only when relevant file content, dependency/config, environment, and command are equivalent; mtime or "passed earlier" is not enough.
- Keep the independent reviewer and required checks after review fixes. If the diff changed, re-evaluate the affected proof; never carry an old verdict onto a new tree.

## 5. Runtime pilot (DeepSeek Harness, outside zharness)

These are runtime settings, not zharness policy. zharness stays runtime-neutral.

1. **Tool-result pruner first.** `dsh-compaction-tool-result-pruner` is installed but not referenced in `~/.dsh/profiles/web/cordis.yml` or `~/.dsh/settings.yaml`. Expected ceiling at defaults: about 11.8% of weighted context (RC2). Before enabling, confirm from source whether it prunes on every tool result or only inside a compaction pass; if only the latter, it does nothing while compaction never triggers.
2. **Compaction threshold.** On the case's provider/model, try a threshold of about **100,000 tokens** with a retained tail of about **20,000 tokens**, then measure quality and cost. With a 1M window, the installed schema supports `thresholdRatio=0.10` plus `retainTokens=20000`. Retention must change too: the default 16% becomes 160,000 tokens, larger than the new threshold, and the validator rejects it (`retainTokens must be less than threshold tokens`).

These are experiment parameters, not a config patch to apply verbatim. Implementation must read the effective profile, choose an override for the right model, and confirm the plugin registers. Do not strip reasoning from the wire protocol.

## 6. Rollout order and acceptance criteria

1. Routing `auto`/bounded/full and the active-plan precondition — **done** (S1).
2. Pilot the pruner, then the context budget, on one profile/model, with phase checkpoints.
3. Apply bounded tool output, selective delegation, and consolidated records/checks; measure with new logs.

Benchmark at least three kinds of work: a few-line bounded task; a UI fix with e2e; one durable phase with independent review. Compare on the same baseline code, model/reasoning, requirements, and acceptance checks; run in a dedicated checkout, never a working tree with work in progress.

**Primary metric — cost-weighted input/output:**

```text
cost = uncached_input × p_uncached + cache_read × p_cached + output × p_output (+ summarization calls)
```

Use the prices that applied to the account at run time. Total prompt tokens are a secondary metric.

| Condition | Pass criterion |
|---|---|
| Quality | Equivalent requirements and acceptance checks; no added skips, loosened assertions, or dropped mandatory reviewers |
| Bounded task | No plan created and no planning invoked just because a plan is missing; no agent spawned for a small, scoped task |
| Resume | No lost decisions/blockers; no finished task redone because a checkpoint was missing or wrong |
| Cost | Report cached/uncached/output separately, including summarization requests if they carry usage; pilot target is at least 30% lower cost-weighted total for equivalent work — a target, not a result |
| Calls | Report main and every child; total calls must not rise enough to cancel the saving from narrower outputs; explain tasks over the S3 diagnostic trigger by progress or a concrete blocker |

Compaction changes the cached prefix and needs a summarization call. Lower total input may not mean lower cost. If quality drops, context is lost, or cost rises, restore the profile config saved before the pilot and keep the evidence for analysis.

Do not mark the optimization complete just because docs are shorter, config parses, or one easier task used fewer tokens than this baseline.

### Open questions

- Why two independent full reviews ran. `check.md` requires the complete `full` review once, on the final phase; the second is legitimate only if the first returned `REQUEST_CHANGES`. Resolve by reading that child session's triggering prompt.
- Whether the pruner prunes per tool result or only during compaction (§5 step 1).

## 7. Re-verifying the numbers

Both scripts only read logs/caches and print aggregates; they print no prompt, tool output, or credential. They need `python3` and `zstd`. If the session keeps running, new numbers can differ from this snapshot. They hard-code this machine's session path, so they are audit evidence, not reproducible CI proofs.

### 7.1 Usage totals

```sh
python3 - <<'PY'
from pathlib import Path
import json
import subprocess

home = Path.home() / '.dsh'
root = home / 'sessions/--Users-tinhtute-Personal-baroiplayer--'
main = 'session-94f6b9a0-f451-4137-bb9f-ee91e2c1515f'

def read_log(sid):
    raw = subprocess.check_output(['zstd', '-dc', str(root / sid / 'session.v3.jsonl.zstd')], stdin=subprocess.DEVNULL)
    return [json.loads(line) for line in raw.splitlines()]

totals = {'calls': 0, 'uncached': 0, 'cached': 0, 'output': 0}
queue, seen = [main], set()
while queue:
    sid = queue.pop(0)
    if sid in seen:
        continue
    seen.add(sid)
    events = read_log(sid)
    queue += [e['data']['childId'] for e in events if e['type'] == 'subagent/catalog']
    cache = json.loads((home / 'storages/session_projcache/sessions' / f'{sid}.json').read_text())['record']
    assert cache['identity']['inheritedEventCount'] == 0, 'Handle inherited usage before summing'
    usage = [e['data']['usage'] for e in events if e['type'] == 'assistant/message' and 'usage' in e['data']]
    row = {'calls': len(usage), 'uncached': sum(u.get('inputTokens', 0) for u in usage),
           'cached': sum(u.get('cacheReadTokens', 0) for u in usage),
           'output': sum(u.get('outputTokens', 0) for u in usage)}
    expected = cache['rows']['tokenUsage']['val']['totals']
    assert (row['uncached'], row['cached'], row['output']) == (
        expected['uncachedInputTokens'], expected['cacheReadTokens'], expected['outputTokens']
    ), 'Log/cache differ; retry after the session is idle'
    print(sid, row)
    for key in totals:
        totals[key] += row[key]
print('SESSIONS', len(seen), 'TOTAL', totals, 'prompt', totals['uncached'] + totals['cached'])
PY
```

This sums stored conversation responses that carry usage. One attempt without usage appears in the main log; auxiliary requests that store no usage, if any, are not estimated into the baseline.

### 7.2 Context composition

```sh
python3 - <<'PY'
from pathlib import Path
from bisect import bisect_right
from collections import Counter
import json
import subprocess

root = Path.home() / '.dsh/sessions/--Users-tinhtute-Personal-baroiplayer--'
main = 'session-94f6b9a0-f451-4137-bb9f-ee91e2c1515f'
PRUNE_THRESHOLD, PRUNE_KEPT = 8192, 4096 + 1024 + 40

def read_log(sid):
    raw = subprocess.check_output(['zstd', '-dc', str(root / sid / 'session.v3.jsonl.zstd')], stdin=subprocess.DEVNULL)
    return [json.loads(line) for line in raw.splitlines()]

def chars(item):
    if isinstance(item, str):
        return len(item)
    text = item.get('text', item.get('reasoning'))
    return len(text) if isinstance(text, str) else len(json.dumps(item))

weighted, pruned, depth_max = Counter(), 0, 0
queue, seen = [(main, 0)], set()
while queue:
    sid, depth = queue.pop(0)
    if sid in seen:
        continue
    seen.add(sid)
    depth_max = max(depth_max, depth)
    events = read_log(sid)
    queue += [(e['data']['childId'], depth + 1) for e in events if e['type'] == 'subagent/catalog']
    calls = [e['seq'] for e in events if e['type'] == 'assistant/message' and 'usage' in e['data']]
    names = {e['data']['callId']: e['data']['name'] for e in events if e['type'] == 'tool/call'}
    for e in events:
        resent = len(calls) - bisect_right(calls, e.get('seq', -1))
        data, parts = e.get('data', {}), []
        if e['type'] == 'tool/result':
            size = sum(chars(x) for x in data['message']['content'])
            parts = [('tool_result:' + names.get(data['message']['source']['callId'], '?'), size)]
            if size > PRUNE_THRESHOLD:
                pruned += (size - PRUNE_KEPT) * resent
        elif e['type'] == 'assistant/message':
            parts = [('assistant:' + x.get('type', '?'), chars(x)) for x in data['message']['content']]
        elif e['type'] == 'user/message':
            parts = [('user', chars(x)) for x in data['content']]
        elif e['type'] == 'system/message':
            parts = [('system', chars(x)) for x in data['message']['content']]
        elif e['type'] == 'agent/inbox/spliced':
            parts = [('inbox', sum(chars(c) for c in m['content'])) for m in data['inserted']]
        for key, size in parts:
            weighted[key] += size * resent

total = sum(weighted.values())
print('sessions', len(seen), 'max_depth', depth_max, 'weighted_chars', total)
for key, value in weighted.most_common():
    if value / total >= 0.005:
        print(f'{key:28} {value / total:6.1%}')
print(f'default pruner removes ~{pruned / total:.1%} of weighted chars')
PY
```

## 8. Sources and scope

- The local logs/caches in §2 and the scripts in §7 are the quantitative evidence; no raw transcript or credential is committed.
- Installed runtime source under `$HOME/.dsh/profiles/node_modules/@deepseek-ai/`: `dsh-llm-deepseek/lib/index.js` (`serializeAssistant`, `mapUsage`), `dsh-compaction-basic/lib/index.js` (`DEFAULT_THRESHOLD_RATIO`, `resolveCompactSpec`), `dsh-compaction-tool-result-pruner/lib/index.js` (`DEFAULTS`). These modules report version `0.1.5-rc.2`; the `dsh` CLI package reports `0.1.5-rc.1`. Do not assume current `master` source equals the version that ran.
- [DeepSeek Harness session model](https://github.com/deepseek-ai/deepseek-harness/blob/master/packages/core/session/README.md#model-experience): message history, resend, and compaction.
- zharness authority/procedure: [workflow](../WORKFLOW.md), [work](../playbooks/work.md), [check](../playbooks/check.md), [handoff](../playbooks/handoff.md).

This audit does not change `baroiplayer` or the runtime. Apart from the S1 routing fix, it records evidence, the preferred direction, and the conditions for implementing and measuring in a later step.

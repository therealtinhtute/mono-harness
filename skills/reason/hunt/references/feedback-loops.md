# Feedback Loops

Load from `hunt` step 1. A tight pass/fail signal that goes red on this bug is 90% of the fix. Treat the loop as a product: build it, then tighten it.

## Ways to build one, roughly in this order

1. **Failing test** at whatever seam reaches the bug: unit, integration, e2e.
2. **HTTP script** (`curl`, a small client) against a running dev server.
3. **CLI invocation** with a fixture input, diffing stdout against a known-good snapshot.
4. **Headless browser script** (Playwright/Puppeteer) driving the UI and asserting on DOM, console, or network.
5. **Replay a captured trace**: save a real request, payload, or event log to disk and replay it through the code path in isolation.
6. **Throwaway harness**: a minimal subset of the system (one service, faked dependencies) that hits the bug path with one call.
7. **Property or fuzz loop**: for "sometimes wrong output", run 1000 random inputs and look for the failure mode.
8. **Bisection harness**: when the bug appeared between two known states (commit, dataset, version), automate "boot at state X, check" so `git bisect run` can drive it.
9. **Differential loop**: run the same input through old vs new version (or two configs) and diff the outputs.
10. **Human-in-the-loop script**: last resort. When a person must click, drive them with `scripts/hitl-loop.template.sh` (copy it, edit the `step`/`capture` lines, run it); captured answers come back as `KEY=VALUE`.

## Tighten it

- Faster: cache setup, skip unrelated init, narrow the test scope.
- Sharper: assert the specific symptom, not "didn't crash".
- More deterministic: pin time, seed randomness, isolate the filesystem, freeze the network.

A 30-second flaky loop is barely better than none; a 2-second deterministic one is a superpower.

## Non-deterministic bugs

The goal is a higher reproduction rate, not a clean repro. Loop the trigger 100×, parallelise, add load, narrow timing windows, inject sleeps at the suspected race. A 50% flake is debuggable; 1% is not, so keep raising the rate. Capture event identity, monotonic order, and thread/task identity in every run.

## Environment-bound bugs

When only the reporter's machine reproduces it, the next artifact is a read-only probe, not another hypothesis:

- Print the environment, the disputed measurement, and the state the hypothesis depends on.
- Discover rather than hardcode: their install method, paths, locale, shell, and versions differ from yours.
- Print nothing that could carry a secret or a private path.
- Ship it as copyable text: one command to run, one block to paste back.

## When no loop is possible

Stop and say so. List every approach tried. Ask for one of: access to an environment that reproduces it, a redacted captured artifact (HAR, log dump, core dump, timestamped screen recording), or permission to add temporary instrumentation where it reproduces. Do not proceed to hypotheses without a loop.

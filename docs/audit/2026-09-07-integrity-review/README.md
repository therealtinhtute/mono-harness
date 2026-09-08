# Opt-in integrity audit probes

These fixtures accompany [the audit report](../2026-09-07-integrity-review.md).
They are not part of the existing Go test suite and make no production changes.

From a complete repository checkout with the audit files applied:

```bash
python3 docs/audit/2026-09-07-integrity-review/run.py \
  --output /tmp/mono-harness-integrity-audit.json
```

Prerequisites: Linux, Python 3.10+, Go 1.23+, bash, git, awk, grep, sed and coreutils
including `sha256sum`. Go 1.23 is sufficient only for the isolated selected
functions; the original CLI module requires Go 1.25.0. No third-party Go modules
are downloaded. The runner creates temporary repositories and files, not changes
to the input checkout. Its Go cache is under the system temporary directory.

Exit codes:

- `0`: all included desired contracts and controls hold.
- `1`: one or more desired contracts fail; the reviewed baseline is expected to
  return this. A red contract probe is evidence of a defect/limitation, not a
  passing test-suite claim.
- `2`: prerequisite, source-shape, execution, or control failure. Do not interpret
  this as a confirmed new product defect.

The recorded baseline has 20 checks: 9 passing controls and 11 failures grouped
into eight findings plus two explicitly documented limitations. F02 has both a
completed-addition and completed-rename scenario.

The Go fixture has a `.go.txt` extension intentionally. The runner extracts
selected original functions into an isolated temporary package and copies the
fixture as a test there. Original helpers are executed, but the binary and its
complete call graph are not built. Shell tests extract only `ZGUARD-CORE`; they do
not install the original hook or execute proof commands from real plans. Their
only proof commands are `true` and `false`.

The path selectors are asserted against the reviewed source. When refactoring
selectors or helper signatures, inspect and update the harness instead of
assuming it automatically validates a new design. The fixture constants are
frozen to the reviewed revision. These probes should ultimately be migrated into
native package tests and full hook/CI integration tests by the corrective PRs.

The standalone download includes a `source-excerpts/` workspace because the audit
execution environment could not clone GitHub. It is explicitly not a complete
checkout. To reproduce that limited execution from the bundle root:

```bash
python3 pr/docs/audit/2026-09-07-integrity-review/run.py \
  --source-root source-excerpts --source-excerpts \
  --output /tmp/mono-harness-integrity-audit.json
```

See `source-provenance.json` for the reviewed full-file blob identifiers and the
scope of the local copies, and `probe-results.json` for function hashes and exact
outputs. Do not replace full repository tests with these focused probes.

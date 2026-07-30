# Failure Ledger

Append-only. A row is added when a durable `check` gate returns `REQUEST_CHANGES`, one row per finding. Never edit or delete an existing row — a wrong row is corrected by appending a new one that says so.

**Graduation rule.** A failure class that appears here a second time must become a deterministic check under `scripts/`. A ledger note alone is not acceptable closure: any failure you do not turn into a permanent test, you will meet again. Until the check exists, the `permanent check` column reads `none yet` and every later gate must state whether the current diff is clean of that class.

Keep the class name stable across rows. The class is what recurs; the specific instance is not.

| date | phase-slug | failure class | how it was caught | permanent check |
| :--- | :--- | :--- | :--- | :--- |
| 2026-07-30 | link-integrity | broken-doc-cross-reference | wrong relative prefix: 9 links written from inside a skill's references directory but still carrying the references prefix, so each resolved one level too deep. Found by the repository-wide audit of 113 tracked markdown files | `scripts/verify-doc-links.sh` |
| 2026-07-30 | link-integrity | broken-doc-cross-reference | stale target: 2 links pointing at paths a *previous completed initiative* had moved — the write skill's SKILL.md relocated under craft, and a plan moved from the active plans directory to completed. Same class, different root cause; this is the recurrence the plan's Premise predicted | `scripts/verify-doc-links.sh` |

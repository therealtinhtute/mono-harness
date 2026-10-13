# Playbook: check (manual review axes)

Loaded by [check.md](check.md) step 5 for `full`, or `bounded` on a diff.

- **Two axes, kept apart** — run them one after the other on `git diff <base>...HEAD`. Sub-agents only for `full` and only when asked (IF unspecified → ask once: single agent (Recommended) / two sub-agents), each briefed per `work-full.md` step 4 with its own axis's sources. They do not change the `judge:` declaration.
- **Spec axis** — sources: the Goal's requirements and non-goals and the phase `goal:`. Report missing, partial, wrong, or unrequested behavior, citing the requirement ID.
- **Standards axis** — sources: the repository's documented standards, then Security, Performance, Architecture, Code Quality. Skip what tooling enforces. Flag as judgement, not violation: Fowler smells (Mysterious Name, Duplicated Code, Feature Envy, Data Clumps, Primitive Obsession, Repeated Switches, Shotgun Surgery, Divergent Change, Speculative Generality, Message Chains, Middle Man, Refused Bequest); tests that are implementation-coupled, tautological, or written in bulk ahead of the code; any deleted, skipped, or weakened test (a blocking finding). For a class-of-bug fix, search sibling instances and state whether coverage is complete.
- Report under `Spec` and `Standards` headings without merging them; step 7 weighs both.

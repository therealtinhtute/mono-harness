## Harness

Start from the requested outcome; the repository is the system of record. Read `.harness/WORKFLOW.md` and only the material the task needs.

- Answers, reviews, diagnoses, plans, and status reports are read-only.
- A bounded change needs no plan: implement and validate.
- Use one `.harness/plans/active/` file when work spans sessions, has dependencies, or needs recovery; move it to `.harness/plans/completed/` only after validation.
- Before adding an externally observable policy, find its repository authority; IF materially different choices remain open → stop before edits (a configurable default is not authority).
- Claim completion only with executable or observable evidence.

`zharness` only installs, updates, and uninstalls. There is no task database. There is no parallel control-plane state.

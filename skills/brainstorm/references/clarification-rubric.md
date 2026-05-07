# Clarification Rubric

Use this rubric before locking the spec.

## Primary dimensions
1. **Goal clarity**
   - Is the target outcome explicit?
   - Is the problem statement concrete?

2. **Actor clarity**
   - Is it clear who uses this or benefits from it?
   - Is the affected surface known?

3. **Boundary clarity**
   - Is in-scope work explicit?
   - Is out-of-scope work explicit?

4. **Constraint clarity**
   - Are technical or product limits known?
   - Are timing/dependency constraints known?

5. **Acceptance clarity**
   - Can a later planner tell when the spec is good enough?

## If clarity is weak
Ask short questions that reduce ambiguity:
- What is the smallest useful outcome?
- What will we explicitly not build here?
- Who is the primary actor?
- What hard constraints already exist?
- What result would make you say “yes, this spec is right”?

## Lock rule
Do not finalize casually.

A spec is ready when:
- goal is explicit
- scope is bounded
- constraints are visible
- acceptance criteria are concrete enough for planning

If not, still write the spec only if unresolved gaps are called out clearly in `Open Questions` and `Ambiguity Report`.

# Prompt Leverage Framework

Use this reference to combine the two source ideas into one practical execution framework.

## Source Synthesis

- Agent Flywheel contributes behavior controls: intensity, wider search, deeper analysis, fresh eyes, first-principles thinking, and future-self clarity.
- OpenAI prompt guidance contributes execution controls: clear objectives, explicit output contracts, tool persistence, dependency checks, verification loops, and completion criteria.

Treat the final framework as:

`Goal -> Context -> Work Style -> Tool Rules -> Output Contract -> Verification -> Done`

## Block Definitions

### Objective

State the task in one or two lines. Define success in observable terms.

### Context

Specify relevant files, URLs, constraints, assumptions, and information boundaries. Say when the agent must retrieve facts instead of guessing.

### Work Style

Control how the agent approaches the task.

- Go broad first when system understanding matters.
- Go deep where risk or complexity is highest.
- Use first-principles reasoning before changing things.
- Re-check with fresh eyes for non-trivial tasks.

### Tool Rules

Define when browsing, file inspection, tests, or external tools are required. Prevent skipping prerequisite checks.

### Output Contract

Define exact structure, tone, formatting, depth, and any required sections or schemas.

### Verification

Require checks for correctness, grounding, completeness, side effects, and better alternatives.

### Done Criteria

Define what must be true before the agent stops.

## Intensity Levels

Use the minimum level that matches the task.

- `Light`: simple edits, formatting, quick rewrites.
- `Standard`: typical coding, research, and drafting tasks.
- `Deep`: debugging, architecture, complex research, or high-stakes outputs.

## Task-Type Adjustments

### Coding

- Emphasize repo context, file inspection, smallest correct change, validation, and edge cases.

### Research

- Emphasize source quality, evidence gathering, synthesis, uncertainty, and citations.

### Writing

- Emphasize audience, tone, structure, constraints, and revision criteria.

### Review

- Emphasize fresh-eyes critique, failure modes, alternatives, and explicit severity.

## Prompt Upgrade Heuristics

- Add missing blocks only when they materially improve execution.
- Do not turn a one-line request into a giant spec unless the task is genuinely complex.
- Preserve user language where possible so the upgraded prompt still feels native.
- Prefer concrete completion criteria over vague quality adjectives.

## Upgrade Rubric

An upgraded prompt is good when it:

1. preserves original intent
2. reduces ambiguity
3. sets the right depth and care level
4. defines the expected output clearly
5. includes an appropriate verification step
6. tells the agent when to stop

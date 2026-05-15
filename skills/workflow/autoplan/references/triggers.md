# Autoplan Trigger Guide

Use this file when routing feels ambiguous.

## Strong positive patterns

Route to `autoplan` when the user says things like:
- "turn this into a plan"
- "break this into phases"
- "make this execution-ready"
- "start from this spec"
- "scope this out"
- "what would the roadmap look like?"
- "help me turn these notes into tasks"
- "autoplan-lite this"

## Strong auto-trigger patterns

Auto-trigger even without the word `autoplan` when the user provides:
- a rough idea + asks for a concrete plan
- a markdown spec + asks for task breakdown
- scattered notes + asks for structure, phases, roadmap, or next steps
- a partially-defined project + asks to clarify missing pieces before execution

## Negative patterns

Do not route to `autoplan` when the user is really asking for:
- immediate implementation of a specific task
- a long-running supervised worker loop
- pure brainstorming without narrowing
- deep conceptual reframing before planning

## Tie-breaks

- choose `autoplan` over `goal` when planning quality is still the bottleneck
- choose `goal` over `autoplan` when the plan already exists and follow-through is the bottleneck
- choose `problem-solving` over `autoplan` when the user keeps circling a messy abstraction or bad system shape

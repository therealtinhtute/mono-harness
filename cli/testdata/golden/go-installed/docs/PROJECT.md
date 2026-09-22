# PROJECT — identity

## What is this project?
- A consumer repository that predates zharness.

## Who is it for?
- The team that owns it.

## Non-goals
- Everything zharness does not manage.

## What are the gate commands?
- run from: repository root
- tests: `make test`
- types: n/a
- lint: `make lint`
- build: `make build`
- format: n/a

## Architecture in one breath
- runtime shape: one service
- where state lives: git
- what are the entrypoints: `cmd/serve`

## What are we working on right now?
- plan: docs/plans/active/consumer.md (active)

# CLI Framework Decision Matrix

Choose based on constraints, not preference. Use this table to recommend.

## Quick Decision

| Constraint | Recommendation |
|---|---|
| Must be single binary, no runtime | Go or Rust |
| Team knows TypeScript, internal tool | Node.js (Commander) |
| Performance-critical, systems-level | Rust (Clap) |
| Fast iteration, good enough perf | Go (Cobra) |
| Simple automation, <200 lines | Bash + gum |
| Cross-platform desktop bundling | Rust or Go |

## Framework Comparison

### Go — Cobra/urfave

**Best for**: DevOps tools, platform CLIs, moderate complexity.

| Aspect | Details |
|---|---|
| Arg parsing | Cobra (most popular), urfave/cli (simpler) |
| Binary size | ~5-10MB (static, no deps) |
| Cross-compile | `GOOS=linux GOARCH=amd64 go build` — trivial |
| Startup time | ~5ms |
| Distribution | Single binary, Homebrew, goreleaser |
| Strengths | Fast compile, easy cross-compile, great stdlib |
| Weaknesses | Verbose error handling, no sum types |

**Typical project structure**:
```
cmd/
  root.go
  serve.go
  migrate.go
internal/
  config/
  client/
main.go
```

### Rust — Clap

**Best for**: Performance-critical tools, security tools, long-lived processes.

| Aspect | Details |
|---|---|
| Arg parsing | Clap (derive or builder API) |
| Binary size | ~1-3MB (stripped, static musl) |
| Cross-compile | cross or cargo-zigbuild |
| Startup time | ~1ms |
| Distribution | Single binary, Homebrew, cargo install |
| Strengths | Zero-cost abstractions, memory safety, small binaries |
| Weaknesses | Slow compile, steeper learning curve, cross-compile less trivial |

**Typical project structure**:
```
src/
  main.rs
  cli.rs       # Clap definitions
  commands/
    mod.rs
    serve.rs
  config.rs
```

### Node.js — Commander/oclif

**Best for**: Internal tools, rapid prototyping, JS/TS ecosystem integration.

| Aspect | Details |
|---|---|
| Arg parsing | Commander (simple), oclif (enterprise), yargs (flexible) |
| Binary size | N/A (requires Node) or ~50MB (pkg/sea) |
| Cross-compile | N/A or use `pkg`/Node SEA |
| Startup time | ~50-100ms |
| Distribution | npm, npx, homebrew (via formula) |
| Strengths | Fast development, huge ecosystem, easy testing |
| Weaknesses | Requires runtime, slow startup, large bundle if compiled |

**Typical project structure**:
```
src/
  index.ts
  commands/
    serve.ts
    migrate.ts
  lib/
    config.ts
bin/
  cli.ts       # entry point
```

### Bash + gum

**Best for**: Simple automation, glue scripts, <200 lines of logic.

| Aspect | Details |
|---|---|
| Arg parsing | getopts, manual parsing, or gum |
| Binary size | N/A (script) |
| Cross-compile | N/A |
| Startup time | ~2ms |
| Distribution | Copy script, Homebrew tap, curl-pipe-bash |
| Strengths | Zero dependencies (mostly), instant iteration |
| Weaknesses | No type safety, hard to test, breaks on edge cases |

**When to graduate from Bash**: >200 lines, complex data structures, multiple subcommands, needs testing, cross-platform.

## Distribution Method by Framework

| Method | Go | Rust | Node | Bash |
|---|---|---|---|---|
| Homebrew | goreleaser | cargo-dist | formula | tap |
| npm | N/A | N/A | native | N/A |
| GitHub Releases | goreleaser | cargo-dist | pkg | tarball |
| Docker | multi-stage | multi-stage | node image | alpine |
| AUR | PKGBUILD | PKGBUILD | N/A | PKGBUILD |
| curl-pipe-sh | install script | install script | install script | native |

## Decision Factors

When recommending, weigh these in order:

1. **Team expertise** — don't pick Rust if nobody knows Rust
2. **Distribution target** — npm users expect `npx`, ops teams expect a binary
3. **Maintenance budget** — Go/Rust compile-check on CI; Bash has no safety net
4. **Performance needs** — matters for hot-path tools, irrelevant for occasional use
5. **Ecosystem deps** — if you need a specific library, that picks the language

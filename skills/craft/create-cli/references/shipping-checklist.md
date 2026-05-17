# CLI Shipping Checklist

Use this to plan distribution after the spec is approved. Check all that apply.

## Pre-Release

- [ ] Version strategy decided (semver)
- [ ] Changelog format chosen (keep-a-changelog, conventional-commits, or git-cliff)
- [ ] License file present
- [ ] README with: install, usage, examples, configuration

## Build & Package

### Go (goreleaser)

```yaml
# .goreleaser.yml minimum
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
brews:
  - repository:
      owner: <user>
      name: homebrew-tap
```

- [ ] goreleaser config
- [ ] GitHub Actions workflow for tag-triggered release
- [ ] Cross-compile targets: linux/darwin amd64+arm64, windows amd64
- [ ] Static linking (CGO_ENABLED=0)

### Rust (cargo-dist)

```toml
# Cargo.toml additions
[package.metadata.dist]
cargo-dist-version = "0.x"
installers = ["shell", "homebrew"]
targets = ["x86_64-unknown-linux-gnu", "aarch64-apple-darwin", "x86_64-apple-darwin"]
```

- [ ] cargo-dist init
- [ ] CI workflow generated
- [ ] musl target for static Linux binary
- [ ] Strip symbols in release profile

### Node.js (npm)

```json
{
  "bin": { "mytool": "./bin/cli.js" },
  "files": ["bin/", "dist/"],
  "engines": { "node": ">=18" }
}
```

- [ ] `bin` field in package.json
- [ ] `files` field (don't ship tests/src)
- [ ] `engines` field
- [ ] `prepublishOnly` script builds
- [ ] npm provenance (`--provenance` flag)

### Bash

- [ ] Shebang: `#!/usr/bin/env bash`
- [ ] Homebrew formula or tap
- [ ] Install script (curl-pipe-bash pattern)
- [ ] Version embedded in script

## Distribution Channels

### Homebrew (macOS/Linux)

**Personal tap** (fastest):
```
homebrew-tap/
  Formula/
    mytool.rb
```

**Official homebrew-core** (high bar): 30+ stars, notable, stable.

- [ ] Tap repository created
- [ ] Formula tested locally (`brew install --build-from-source`)
- [ ] Auto-update formula on release (goreleaser/cargo-dist handle this)

### npm

- [ ] `npm publish` in CI
- [ ] Scoped or unscoped package name reserved
- [ ] `npx` works without global install
- [ ] Provenance enabled for supply chain security

### GitHub Releases

- [ ] Release notes auto-generated (from commits or changelog)
- [ ] Binary checksums (SHA256) in release body
- [ ] Signed binaries (optional, but good)
- [ ] Install script in README

### Container (Docker)

- [ ] Multi-stage Dockerfile (build + runtime)
- [ ] Minimal base image (distroless, alpine, scratch)
- [ ] Published to ghcr.io or Docker Hub
- [ ] Version-tagged and `latest`

## CI/CD Pipeline

Minimum viable pipeline:

```
on push to main:     lint + test
on tag v*:           build + release
on PR:               lint + test + build (no publish)
```

- [ ] Tests run on every push
- [ ] Release triggered by git tag (v1.0.0)
- [ ] No manual steps between tag and published release
- [ ] Install script tested in CI (curl | bash in a fresh container)

## Post-Release

- [ ] Verify install works: `brew install`, `npm install -g`, `curl | bash`
- [ ] Shell completions installable and documented
- [ ] Man page generated (optional but professional)
- [ ] Update motd / version check (optional, respect user — make it opt-in)

## Versioning Rules

- Breaking CLI changes (renamed flags, changed output format): **major** bump
- New commands/flags, backwards-compatible: **minor** bump
- Bug fixes: **patch** bump
- Pre-1.0: anything goes, but document breaking changes
- After 1.0: deprecate before removing (warn for one minor version)

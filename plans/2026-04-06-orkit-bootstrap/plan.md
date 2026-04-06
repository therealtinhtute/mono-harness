---
title: Orkit - Claude Code Extensions Marketplace
description: Public marketplace/registry for Claude Code extensions with curated, production-grade skills, agents, hooks, and statusline configs
status: draft
created: 2026-04-06
priority: P1
effort: large
branch: main
tags: [marketplace, plugins, extensions, cli, validation]
---

# Orkit Implementation Plan

## Overview

Orkit is a curated marketplace for Claude Code extensions, leveraging the built-in plugin system. Focus: production-grade quality, security validation, and developer experience.

## Architecture Decision

**Approach**: Enhanced Marketplace (Hybrid)
- Use Claude Code's native `marketplace.json` format for distribution
- Build bash CLI tooling for validation, scaffolding, and quality gates
- Host on GitHub with automated CI/CD pipeline
- Installation via `/plugin marketplace add github:tinhtute/orkit`

**Rationale**: Leverage existing plugin infrastructure while adding curation layer.

## Project Structure

```
orkit/
├── .claude-plugin/
│   ├── plugin.json              # Orkit as a plugin itself
│   └── marketplace.json         # Extension catalog
├── plugins/                     # Extension source
│   ├── code-reviewer/
│   ├── tester/
│   ├── debugger/
│   ├── docs-manager/
│   └── planner/
├── cli/                         # Bash tooling
│   ├── orkit                    # Main CLI entry
│   ├── lib/
│   │   ├── validate.sh          # Schema validation
│   │   ├── security.sh          # Security scanning
│   │   ├── quality.sh           # Code quality checks
│   │   └── scaffold.sh          # Extension scaffolding
│   └── templates/               # Scaffolding templates
├── schemas/                     # JSON schemas
│   ├── skill.schema.json
│   ├── agent.schema.json
│   ├── hook.schema.json
│   └── plugin.schema.json
├── docs/                        # Developer guide
│   ├── getting-started.md
│   ├── creating-skills.md
│   ├── creating-agents.md
│   ├── creating-hooks.md
│   ├── validation-rules.md
│   └── contribution-guide.md
├── .github/
│   └── workflows/
│       ├── validate.yml         # PR validation
│       ├── security-scan.yml    # Security checks
│       └── publish.yml          # Release automation
└── tests/                       # Test suite
    ├── validation/
    ├── security/
    └── integration/
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
- Set up repository structure
- Create marketplace.json catalog format
- Define JSON schemas for all extension types
- Build basic CLI scaffolding tool
- Write initial developer guide outline

### Phase 2: Validation Pipeline (Week 1-2)
- Schema validation (YAML frontmatter, JSON configs)
- Security scanning (secrets detection, malicious code patterns)
- Code quality checks (shellcheck, markdown linting)
- Exit code standards enforcement
- Automated test harness

### Phase 3: Initial Extensions (Week 2-3)
Create 5-10 production-grade extensions:
1. **code-reviewer**: Code review agent with security/performance focus
2. **tester**: Test execution and analysis agent
3. **debugger**: Systematic debugging agent
4. **docs-manager**: Documentation maintenance agent
5. **planner**: Implementation planning agent
6. **git-hooks**: Pre-commit/pre-push automation hooks
7. **statusline-pro**: Enhanced statusline with git/cost/context info
8. **api-conventions**: API design skill
9. **error-handling**: Error handling patterns skill
10. **security-scanner**: Security audit skill

### Phase 4: CLI Tooling (Week 3)
- `orkit validate <path>`: Run validation pipeline
- `orkit scaffold <type> <name>`: Generate extension boilerplate
- `orkit test <path>`: Run test suite
- `orkit publish <path>`: Prepare for marketplace
- `orkit search <query>`: Search extensions (local)
- `orkit info <name>`: Show extension details

### Phase 5: CI/CD Pipeline (Week 3-4)
- GitHub Actions for PR validation
- Automated security scanning on push
- Release automation with date-based tags
- Marketplace.json auto-generation
- Test coverage reporting

### Phase 6: Documentation (Week 4)
- Complete developer guide
- Extension creation tutorials
- Best practices documentation
- Security guidelines
- Contribution workflow
- Example extensions walkthrough

### Phase 7: Distribution (Week 4)
- GitHub releases with date-based versions (2026-04-06)
- Installation instructions
- Update mechanism documentation
- Marketplace registration
- Public announcement

## Technical Specifications

### Marketplace.json Format
```json
{
  "name": "orkit",
  "owner": {
    "name": "Orkit Team",
    "email": "team@orkit.dev"
  },
  "metadata": {
    "description": "Curated marketplace for Claude Code extensions",
    "version": "2026-04-06",
    "pluginRoot": "./plugins"
  },
  "plugins": [
    {
      "name": "code-reviewer",
      "source": "./plugins/code-reviewer",
      "description": "Expert code review agent",
      "version": "1.0.0",
      "category": "agents",
      "tags": ["review", "security", "quality"],
      "strict": true
    }
  ]
}
```

### Validation Rules
1. **Schema Validation**: All YAML/JSON must match schemas
2. **Security Scanning**: No secrets, malicious patterns, or vulnerabilities
3. **Code Quality**: Pass shellcheck, markdown linting
4. **Naming**: kebab-case, unique names, no conflicts
5. **Size Limits**: Metadata <100 tokens, instructions <5k tokens
6. **Documentation**: README.md required for each extension
7. **Testing**: Test suite must pass
8. **License**: Must include license file

### Version Management
- **Format**: Date-based (YYYY-MM-DD)
- **Releases**: Git tags with date
- **Updates**: Users can pin or update to latest
- **Changelog**: Auto-generated from commits
- **Breaking Changes**: Major version bump + migration guide

### Distribution Strategy
- **Primary**: GitHub repository
- **Installation**: `/plugin marketplace add github:tinhtute/orkit`
- **Updates**: `/plugin marketplace update orkit`
- **Hosting**: GitHub Pages for documentation
- **CDN**: GitHub releases for assets

## Success Criteria

### MVP Launch
- [ ] CLI tool validates, scaffolds, and tests extensions
- [ ] 5-10 production-quality extensions available
- [ ] Automated validation pipeline in CI/CD
- [ ] Developer guide published
- [ ] Marketplace.json catalog working
- [ ] Installation via Claude Code plugin system

### Quality Gates
- [ ] All extensions pass validation pipeline
- [ ] Zero security vulnerabilities detected
- [ ] 100% schema compliance
- [ ] Documentation complete and accurate
- [ ] Test coverage >80%

## Dependencies

### External Tools
- `jq`: JSON processing
- `yq`: YAML processing
- `shellcheck`: Shell script linting
- `markdownlint`: Markdown linting
- `gitleaks`: Secret scanning
- `semgrep`: Security pattern matching

### Claude Code Features
- Plugin system (marketplace.json)
- Extension types (skills, agents, hooks, statusline)
- Settings hierarchy
- Installation commands

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Security vulnerabilities in extensions | High | Multi-layer scanning, manual review |
| Breaking changes in Claude Code | High | Version pinning, compatibility testing |
| Low adoption | Medium | Quality over quantity, clear docs |
| Contribution quality | Medium | Invite-only, strict validation |
| Maintenance burden | Medium | Automation, clear guidelines |

## Next Steps

1. Review and approve this plan
2. Create repository structure
3. Start Phase 1 implementation
4. Set up CI/CD pipeline
5. Begin extension development

## Open Questions

1. Should we support private extensions or public-only?
2. What's the approval process for new contributors?
3. How do we handle extension deprecation?
4. Should we provide a web UI for browsing?
5. What metrics should we track for usage?

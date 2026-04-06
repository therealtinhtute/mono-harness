---
date: 2026-04-06
project: orkit
phase: requirements-gathering
interviewer: claude
---

# Interview Notes: Orkit - Claude Code Extensions Marketplace

## Project Vision

**Orkit** is a public marketplace/registry for Claude Code extensions, enabling users to discover, install, and share skills, subagents, hooks, and statusline configurations.

## Key Requirements

### 1. Purpose & Goals
- **Primary Goal**: Public marketplace/registry (like a package registry for Claude Code extensions)
- **Project Name**: `orkit`
- **Target Audience**: Claude Code users looking to extend functionality
- **Value Proposition**: Curated, production-grade extensions with easy installation

### 2. Content Types (All Four Categories)
- ✅ **Skills**: Reusable prompt templates and workflows (`/command` syntax)
- ✅ **Subagent Definitions**: Specialized agent configurations (reviewers, testers, researchers)
- ✅ **Hooks**: Event-triggered shell commands (automation on tool calls, file changes, events)
- ✅ **Statusline Configs**: Custom statusline displays (project info, git status, context)

### 3. Distribution & Installation
- **Method**: CLI installer tool (bash-based)
- **User Flow**: 
  - Search/browse available extensions
  - Preview extension details
  - Install directly into user's `.claude` directory
- **Technology**: Pure bash/shell for maximum portability

### 4. Quality & Validation
- **Quality Bar**: Production-grade with CI/CD
- **Validation Pipeline**:
  - ✅ Schema validation (frontmatter, structure, required fields)
  - ✅ Security scanning (secrets, malicious code, vulnerabilities)
  - ✅ Code quality checks (shell linting, best practices)
- **Testing**: Automated checks before publishing

### 5. Contribution Model
- **Type**: Closed/invite-only contributors
- **Rationale**: Maintain high quality bar, curated collection
- **Process**: Invited contributors submit PRs → automated validation → maintainer review → merge

### 6. Documentation Strategy
- **Focus**: Developer guide
- **Content**: Comprehensive guide on creating skills, hooks, subagents, and statusline configs
- **Format**: Markdown documentation in repo
- **Goal**: Enable contributors to create high-quality extensions

### 7. Versioning & Updates
- **Strategy**: Date-based versions (e.g., `2026-04-06`)
- **Rationale**: Stability through snapshots, clear point-in-time references
- **Update Mechanism**: Users can update to latest snapshot or pin to specific date

### 8. Initial Scope
- **Launch Size**: Small curated set (5-10 high-quality extensions)
- **Strategy**: Demonstrate best practices with well-tested examples
- **Growth**: Expand collection over time with invited contributors

## Technical Architecture Decisions

### Technology Stack
- **CLI Tool**: Pure bash/shell scripts
- **Validation**: Shell-based linting and validation tools
- **CI/CD**: GitHub Actions for automated testing
- **Distribution**: Git-based with CLI installer

### Project Structure (To Be Designed)
- Extension registry/catalog
- CLI installer tool
- Validation tooling
- Documentation
- Example extensions
- Contribution templates

## Open Questions & Next Steps

### Questions to Resolve
1. **Extension Format**: What's the standard structure for each extension type?
2. **Installation Location**: Where in `.claude` directory should extensions be installed?
3. **Dependency Management**: How to handle extensions that depend on other extensions?
4. **Conflict Resolution**: What happens if user already has a skill with the same name?
5. **Uninstall Process**: How do users remove installed extensions?
6. **Update Notifications**: How do users know when new versions are available?

### Immediate Next Steps
1. Research existing Claude Code extension formats and conventions
2. Design project structure and directory layout
3. Define extension schemas for each type (skills, subagents, hooks, statusline)
4. Create CLI installer prototype
5. Build validation pipeline
6. Develop 2-3 example extensions as proof of concept
7. Write developer guide

## Success Criteria

### MVP Launch
- [ ] CLI tool can search, preview, and install extensions
- [ ] 5-10 production-quality extensions available
- [ ] Automated validation pipeline working
- [ ] Developer guide published
- [ ] Date-based versioning implemented

### Long-term Goals
- Grow to 20-30 curated extensions
- Establish contributor community (invite-only)
- Become go-to resource for Claude Code extensions
- Maintain high quality and security standards

## Notes
- Prioritize simplicity and portability (bash-only)
- Security is critical - scanning and validation required
- Quality over quantity - curated collection
- Clear documentation for contributors

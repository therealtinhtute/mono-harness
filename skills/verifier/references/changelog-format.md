---
category: reference
type: format
tags: [changelog, documentation]
---

# Changelog Format Guide

## Standard: Keep a Changelog

https://keepachangelog.com/

## Format

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Now removed features

### Fixed
- Bug fixes

### Security
- Security improvements

## [1.0.0] - YYYY-MM-DD

### Added
- Initial release
```

## Categories

| Category | Use For | Example |
|----------|---------|---------|
| **Added** | New features | "Add user authentication" |
| **Changed** | Changes to existing | "Update API response format" |
| **Deprecated** | Soon-to-remove | "Deprecate old API endpoint" |
| **Removed** | Removed features | "Remove legacy auth method" |
| **Fixed** | Bug fixes | "Fix login redirect issue" |
| **Security** | Security-related | "Fix XSS vulnerability" |

## Entry Format

```
- [Scope] Description (#PR)
```

Examples:
```markdown
- [Auth] Add OAuth2 login support (#123)
- [API] Fix pagination in user list (#124)
- [Docs] Update installation guide (#125)
```

## Auto-Generation from Git

### Conventional Commits Mapping

| Commit Type | Changelog Category |
|-------------|-------------------|
| `feat:` | ### Added |
| `fix:` | ### Fixed |
| `docs:` | ### Changed (if docs) |
| `style:` | (no changelog) |
| `refactor:` | ### Changed |
| `perf:` | ### Changed |
| `test:` | (no changelog) |
| `chore:` | (no changelog) |
| `security:` | ### Security |

### Git Commands

```bash
# Get commits since last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# Get conventional commits
git log --pretty=format:"%s" --no-merges -20

# Group by type
git log --pretty=format:"%s" | grep -E "^(feat|fix|docs):"
```

## Best Practices

1. **One change per line**
2. **Be specific**: "Fix bug" → "Fix login redirect for OAuth users"
3. **Link to issues/PRs**: Add `(#123)` at end
4. **Group by scope**: `[Auth]`, `[API]`, `[UI]`
5. **Keep [Unreleased] updated**: Don't wait for release
6. **Remove empty sections**: No "### Removed" if nothing removed

## Example Entries

```markdown
### Added
- [Auth] Implement JWT token refresh (#234)
- [API] Add rate limiting middleware (#235)
- [UI] Add dark mode toggle (#236)

### Changed
- [API] Improve error response format (#237)
- [Deps] Upgrade to React 18 (#238)

### Fixed
- [Auth] Fix session timeout not clearing (#239)
- [UI] Fix mobile navigation overlap (#240)

### Security
- [Auth] Add CSRF protection to forms (#241)
```

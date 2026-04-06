---
title: Phase 5 - CI/CD Pipeline
description: Automated validation, testing, and release workflows
status: draft
created: 2026-04-06
phase: 5
effort: medium
---

# Phase 5: CI/CD Pipeline

## Objectives

Establish automated CI/CD pipeline:
- PR validation workflow
- Security scanning on push
- Automated testing
- Release automation with date-based tags
- Marketplace.json auto-generation
- Test coverage reporting

## Tasks

### 5.1 PR Validation Workflow

**Create .github/workflows/validate-pr.yml:**

```yaml
name: Validate Pull Request

on:
  pull_request:
    branches: [main]
    paths:
      - 'plugins/**'
      - 'cli/**'
      - 'schemas/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq shellcheck
          
          # Install yq
          sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
          sudo chmod +x /usr/local/bin/yq
          
          # Install markdownlint
          npm install -g markdownlint-cli
          
          # Install gitleaks
          wget -qO- https://github.com/gitleaks/gitleaks/releases/latest/download/gitleaks_linux_x64.tar.gz | tar xz
          sudo mv gitleaks /usr/local/bin/
      
      - name: Detect changed extensions
        id: changes
        run: |
          # Get list of changed plugin directories
          changed_plugins=$(git diff --name-only origin/main...HEAD | grep '^plugins/' | cut -d/ -f1-2 | sort -u)
          echo "changed_plugins<<EOF" >> $GITHUB_OUTPUT
          echo "$changed_plugins" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
      
      - name: Validate changed extensions
        if: steps.changes.outputs.changed_plugins != ''
        run: |
          exit_code=0
          
          while IFS= read -r plugin; do
            if [[ -n "$plugin" ]]; then
              echo "::group::Validating ${plugin}"
              if ! ./cli/orkit validate "$plugin"; then
                echo "::error::Validation failed for ${plugin}"
                exit_code=1
              fi
              echo "::endgroup::"
            fi
          done <<< "${{ steps.changes.outputs.changed_plugins }}"
          
          exit $exit_code
      
      - name: Run tests
        if: steps.changes.outputs.changed_plugins != ''
        run: |
          exit_code=0
          
          while IFS= read -r plugin; do
            if [[ -n "$plugin" ]]; then
              echo "::group::Testing ${plugin}"
              if ! ./cli/orkit test "$plugin"; then
                echo "::error::Tests failed for ${plugin}"
                exit_code=1
              fi
              echo "::endgroup::"
            fi
          done <<< "${{ steps.changes.outputs.changed_plugins }}"
          
          exit $exit_code
      
      - name: Check marketplace.json
        run: |
          if ! jq empty .claude-plugin/marketplace.json; then
            echo "::error::Invalid marketplace.json"
            exit 1
          fi
          echo "✓ marketplace.json is valid"
      
      - name: Post validation summary
        if: always()
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const plugins = `${{ steps.changes.outputs.changed_plugins }}`.split('\n').filter(p => p);
            
            let summary = '## Validation Summary\n\n';
            summary += `**Changed Extensions:** ${plugins.length}\n\n`;
            
            for (const plugin of plugins) {
              summary += `- ${plugin}\n`;
            }
            
            await github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: summary
            });
```

### 5.2 Security Scanning Workflow

**Create .github/workflows/security-scan.yml:**

```yaml
name: Security Scan

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    # Run daily at 2 AM UTC
    - cron: '0 2 * * *'

jobs:
  gitleaks:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Run Gitleaks
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  
  semgrep:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: >-
            p/security-audit
            p/secrets
            p/bash
  
  custom-security:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq
      
      - name: Run custom security checks
        run: |
          source ./cli/lib/security.sh
          
          exit_code=0
          for plugin in plugins/*/; do
            if [[ -d "$plugin" ]]; then
              echo "::group::Scanning ${plugin}"
              if ! scan_security "$plugin"; then
                echo "::error::Security issues found in ${plugin}"
                exit_code=1
              fi
              echo "::endgroup::"
            fi
          done
          
          exit $exit_code
      
      - name: Upload security report
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: security-report
          path: security-report.txt
```

### 5.3 Test Suite Workflow

**Create .github/workflows/test.yml:**

```yaml
name: Test Suite

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq shellcheck
          
          sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
          sudo chmod +x /usr/local/bin/yq
      
      - name: Run validation tests
        run: |
          chmod +x tests/validation/*.sh
          for test in tests/validation/*.sh; do
            echo "Running: $test"
            bash "$test"
          done
      
      - name: Run security tests
        run: |
          chmod +x tests/security/*.sh
          for test in tests/security/*.sh; do
            echo "Running: $test"
            bash "$test"
          done
  
  integration-tests:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq shellcheck
      
      - name: Test CLI commands
        run: |
          # Test scaffold
          ./cli/orkit scaffold skill test-skill
          
          # Test validate
          ./cli/orkit validate plugins/test-skill
          
          # Test list
          ./cli/orkit list
          
          # Cleanup
          rm -rf plugins/test-skill
      
      - name: Test extension installation
        run: |
          # Simulate local installation
          for plugin in plugins/*/; do
            if [[ -d "$plugin" ]]; then
              plugin_name=$(basename "$plugin")
              echo "Testing installation: ${plugin_name}"
              
              # Copy to temp directory
              temp_dir="/tmp/claude-test"
              mkdir -p "${temp_dir}/skills"
              mkdir -p "${temp_dir}/agents"
              
              if [[ -f "${plugin}/SKILL.md" ]]; then
                cp -r "$plugin" "${temp_dir}/skills/"
              elif find "$plugin" -maxdepth 1 -name "*.md" -type f | grep -q .; then
                cp -r "$plugin" "${temp_dir}/agents/"
              fi
              
              echo "✓ ${plugin_name} can be installed"
            fi
          done
          
          rm -rf /tmp/claude-test
```

### 5.4 Release Automation Workflow

**Create .github/workflows/release.yml:**

```yaml
name: Release

on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Release version (YYYY-MM-DD or leave empty for today)'
        required: false
        type: string

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq
      
      - name: Determine version
        id: version
        run: |
          if [[ -n "${{ github.event.inputs.version }}" ]]; then
            version="${{ github.event.inputs.version }}"
          else
            version=$(date +%Y-%m-%d)
          fi
          echo "version=${version}" >> $GITHUB_OUTPUT
          echo "Release version: ${version}"
      
      - name: Update marketplace version
        run: |
          version="${{ steps.version.outputs.version }}"
          jq --arg v "$version" '.metadata.version = $v' \
            .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
          mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
      
      - name: Update plugin versions
        run: |
          version="${{ steps.version.outputs.version }}"
          
          # Update each plugin version in marketplace
          jq --arg v "$version" '
            .plugins = [.plugins[] | .version = $v]
          ' .claude-plugin/marketplace.json > .claude-plugin/marketplace.json.tmp
          mv .claude-plugin/marketplace.json.tmp .claude-plugin/marketplace.json
      
      - name: Generate changelog
        id: changelog
        run: |
          version="${{ steps.version.outputs.version }}"
          
          # Get previous tag
          prev_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
          
          # Generate changelog
          if [[ -n "$prev_tag" ]]; then
            changelog=$(git log ${prev_tag}..HEAD --pretty=format:"- %s (%h)" --no-merges)
          else
            changelog=$(git log --pretty=format:"- %s (%h)" --no-merges)
          fi
          
          # Save to file
          cat > CHANGELOG.tmp <<EOF
          # Release ${version}
          
          ## Changes
          
          ${changelog}
          
          ## Extensions
          
          $(jq -r '.plugins[] | "- \(.name) v\(.version) - \(.description)"' .claude-plugin/marketplace.json)
          
          ## Installation
          
          \`\`\`bash
          /plugin marketplace add github:tinhtute/orkit
          \`\`\`
          EOF
          
          echo "changelog_file=CHANGELOG.tmp" >> $GITHUB_OUTPUT
      
      - name: Commit version updates
        run: |
          version="${{ steps.version.outputs.version }}"
          
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          
          git add .claude-plugin/marketplace.json
          git commit -m "Release ${version}" || echo "No changes to commit"
          git push
      
      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          tag_name: ${{ steps.version.outputs.version }}
          name: Release ${{ steps.version.outputs.version }}
          body_path: ${{ steps.changelog.outputs.changelog_file }}
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Notify release
        run: |
          version="${{ steps.version.outputs.version }}"
          echo "✓ Released version ${version}"
          echo "  View at: https://github.com/${{ github.repository }}/releases/tag/${version}"
```

### 5.5 Marketplace Validation Workflow

**Create .github/workflows/validate-marketplace.yml:**

```yaml
name: Validate Marketplace

on:
  push:
    branches: [main]
    paths:
      - '.claude-plugin/marketplace.json'
  pull_request:
    branches: [main]
    paths:
      - '.claude-plugin/marketplace.json'

jobs:
  validate:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq
          npm install -g ajv-cli
      
      - name: Validate JSON syntax
        run: |
          if ! jq empty .claude-plugin/marketplace.json; then
            echo "::error::Invalid JSON syntax in marketplace.json"
            exit 1
          fi
          echo "✓ JSON syntax valid"
      
      - name: Validate required fields
        run: |
          # Check required top-level fields
          required_fields=("name" "owner" "metadata" "plugins")
          
          for field in "${required_fields[@]}"; do
            if ! jq -e ".$field" .claude-plugin/marketplace.json >/dev/null; then
              echo "::error::Missing required field: $field"
              exit 1
            fi
          done
          
          echo "✓ Required fields present"
      
      - name: Validate plugin entries
        run: |
          # Check each plugin has required fields
          plugins=$(jq -r '.plugins | length' .claude-plugin/marketplace.json)
          
          for ((i=0; i<plugins; i++)); do
            name=$(jq -r ".plugins[$i].name" .claude-plugin/marketplace.json)
            
            # Check required fields
            for field in name source description version; do
              if ! jq -e ".plugins[$i].$field" .claude-plugin/marketplace.json >/dev/null; then
                echo "::error::Plugin $name missing required field: $field"
                exit 1
              fi
            done
            
            # Check source path exists
            source=$(jq -r ".plugins[$i].source" .claude-plugin/marketplace.json)
            if [[ "$source" == ./* ]]; then
              if [[ ! -d "$source" ]]; then
                echo "::error::Plugin $name source path does not exist: $source"
                exit 1
              fi
            fi
          done
          
          echo "✓ All plugin entries valid"
      
      - name: Check for duplicates
        run: |
          # Check for duplicate plugin names
          duplicates=$(jq -r '.plugins[].name' .claude-plugin/marketplace.json | sort | uniq -d)
          
          if [[ -n "$duplicates" ]]; then
            echo "::error::Duplicate plugin names found:"
            echo "$duplicates"
            exit 1
          fi
          
          echo "✓ No duplicate plugin names"
```

### 5.6 Code Coverage Workflow

**Create .github/workflows/coverage.yml:**

```yaml
name: Code Coverage

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  coverage:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq shellcheck
          
          # Install kcov for bash coverage
          sudo apt-get install -y cmake libcurl4-openssl-dev libelf-dev libdw-dev binutils-dev
          wget https://github.com/SimonKagstrom/kcov/archive/master.tar.gz
          tar xzf master.tar.gz
          cd kcov-master
          mkdir build && cd build
          cmake ..
          make
          sudo make install
          cd ../..
      
      - name: Run tests with coverage
        run: |
          mkdir -p coverage
          
          # Run validation tests with coverage
          for test in tests/validation/*.sh; do
            kcov --exclude-pattern=/usr coverage "$test"
          done
          
          # Run security tests with coverage
          for test in tests/security/*.sh; do
            kcov --exclude-pattern=/usr coverage "$test"
          done
      
      - name: Generate coverage report
        run: |
          # Calculate coverage percentage
          coverage_pct=$(kcov --merge coverage coverage/* | grep "percent_covered" | cut -d: -f2 | tr -d ' %,')
          
          echo "Coverage: ${coverage_pct}%"
          
          # Fail if coverage below threshold
          threshold=80
          if (( $(echo "$coverage_pct < $threshold" | bc -l) )); then
            echo "::error::Coverage ${coverage_pct}% is below threshold ${threshold}%"
            exit 1
          fi
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          directory: ./coverage
          fail_ci_if_error: false
```

### 5.7 Dependency Update Workflow

**Create .github/workflows/update-deps.yml:**

```yaml
name: Update Dependencies

on:
  schedule:
    # Run weekly on Monday at 9 AM UTC
    - cron: '0 9 * * 1'
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Check for tool updates
        run: |
          echo "Checking for updates..."
          
          # Check jq version
          current_jq=$(jq --version 2>&1 | grep -oP '\d+\.\d+')
          echo "Current jq: ${current_jq}"
          
          # Check shellcheck version
          current_shellcheck=$(shellcheck --version | grep version: | cut -d: -f2 | xargs)
          echo "Current shellcheck: ${current_shellcheck}"
          
          # Add more checks as needed
      
      - name: Create update PR
        if: false  # Placeholder - implement when needed
        run: |
          echo "Creating PR for dependency updates..."
```

### 5.8 Documentation

**Create docs/ci-cd.md:**

```markdown
# CI/CD Pipeline

Orkit uses GitHub Actions for automated validation, testing, and releases.

## Workflows

### PR Validation
- **Trigger**: Pull requests to main
- **Actions**:
  - Validate changed extensions
  - Run tests
  - Check marketplace.json
  - Post summary comment

### Security Scanning
- **Trigger**: Push to main, PRs, daily schedule
- **Actions**:
  - Gitleaks secret scanning
  - Semgrep security analysis
  - Custom security checks

### Test Suite
- **Trigger**: Push to main, PRs
- **Actions**:
  - Unit tests
  - Integration tests
  - CLI command tests

### Release
- **Trigger**: Manual workflow dispatch
- **Actions**:
  - Update versions
  - Generate changelog
  - Create GitHub release
  - Tag with date

### Marketplace Validation
- **Trigger**: Changes to marketplace.json
- **Actions**:
  - Validate JSON syntax
  - Check required fields
  - Verify plugin entries
  - Check for duplicates

### Code Coverage
- **Trigger**: Push to main, PRs
- **Actions**:
  - Run tests with coverage
  - Generate coverage report
  - Upload to Codecov
  - Enforce 80% threshold

## Release Process

### Automated Release

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Enter version (YYYY-MM-DD) or leave empty for today
4. Workflow will:
   - Update marketplace version
   - Update plugin versions
   - Generate changelog
   - Create GitHub release
   - Tag repository

### Manual Release

```bash
# Update versions
version=$(date +%Y-%m-%d)
jq --arg v "$version" '.metadata.version = $v' \
  .claude-plugin/marketplace.json > tmp.json
mv tmp.json .claude-plugin/marketplace.json

# Commit and tag
git add .claude-plugin/marketplace.json
git commit -m "Release ${version}"
git tag "$version"
git push origin main --tags

# Create GitHub release
gh release create "$version" \
  --title "Release ${version}" \
  --generate-notes
```

## Status Badges

Add to README.md:

```markdown
![Validate PR](https://github.com/tinhtute/orkit/workflows/Validate%20Pull%20Request/badge.svg)
![Security Scan](https://github.com/tinhtute/orkit/workflows/Security%20Scan/badge.svg)
![Test Suite](https://github.com/tinhtute/orkit/workflows/Test%20Suite/badge.svg)
![Coverage](https://codecov.io/gh/tinhtute/orkit/branch/main/graph/badge.svg)
```

## Troubleshooting

### Validation Failures
- Check validation logs in Actions tab
- Run locally: `./cli/orkit validate plugins/extension-name`
- Fix issues and push again

### Security Scan Failures
- Review security report artifact
- Check for secrets or malicious patterns
- Update code and re-run

### Test Failures
- Check test logs
- Run locally: `./cli/orkit test plugins/extension-name`
- Fix failing tests

### Release Failures
- Ensure marketplace.json is valid
- Check for merge conflicts
- Verify GitHub token permissions
```

## Acceptance Criteria

- [ ] PR validation workflow validates changed extensions
- [ ] Security scanning detects secrets and vulnerabilities
- [ ] Test suite runs all tests successfully
- [ ] Release workflow creates dated releases
- [ ] Marketplace validation ensures catalog integrity
- [ ] Code coverage enforces 80% threshold
- [ ] All workflows have proper error handling
- [ ] Documentation explains CI/CD process
- [ ] Status badges added to README

## Dependencies

- GitHub Actions
- jq, yq, shellcheck, markdownlint
- gitleaks, semgrep
- kcov (for coverage)

## Estimated Effort

3-4 days

## Next Phase

Phase 6: Documentation

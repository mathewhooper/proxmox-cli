# Code Linting and Quality Standards

This document describes the linting and code quality standards enforced in this project.

## Overview

The project uses comprehensive linting to ensure code quality, catch bugs early, and maintain consistent code style. Linting runs automatically in CI/CD and can be run locally before committing.

## Quick Start

```bash
# Install all linting tools
make install-linters

# Run all linting checks
make lint

# Format code
make fmt

# Run all CI checks (format, lint, vet, test)
make ci
```

## Linting Tools

The project uses multiple linting tools, each serving a specific purpose:

### 1. **gofmt** (Code Formatting)
- **Purpose**: Ensures consistent code formatting
- **Status**: **REQUIRED** - Build will fail if not formatted
- **Fix**: Run `gofmt -w .` or `make fmt`

### 2. **go vet** (Static Analysis)
- **Purpose**: Detects suspicious constructs and common mistakes
- **Status**: **REQUIRED** - Build will fail if issues found
- **Examples**:
  - Unreachable code
  - Misuse of unsafe pointers
  - Printf format mismatches

### 3. **staticcheck** (Advanced Linter)
- **Purpose**: Advanced static analysis for bugs and performance issues
- **Status**: RECOMMENDED (runs in CI)
- **Install**: `go install honnef.co/go/tools/cmd/staticcheck@latest`

### 4. **golangci-lint** (Meta Linter)
- **Purpose**: Runs multiple linters in parallel with caching, including style checking
- **Status**: RECOMMENDED (comprehensive checks)
- **Install**: See installation section below
- **Config**: `.golangci.yml`
- **Includes**: revive (modern replacement for deprecated golint), gofmt, goimports, and many more

### 5. **errcheck** (Error Handling)
- **Purpose**: Ensures all errors are checked
- **Status**: RECOMMENDED (runs in CI)
- **Install**: `go install github.com/kisielk/errcheck@latest`
- **Why**: Unchecked errors can cause runtime issues
- **Config**: `.errcheck_excludes.txt` - Excludes common test utilities
- **Note**: Test files are excluded from errcheck scanning (`-ignoretests`)

### 6. **gosec** (Security Analysis)
- **Purpose**: Identifies security vulnerabilities
- **Status**: RECOMMENDED (runs in CI)
- **Install**: `go install github.com/securego/gosec/v2/cmd/gosec@latest`
- **Examples**:
  - Weak cryptographic practices
  - SQL injection vulnerabilities
  - Unsafe file operations

## Configuration Files

### `.golangci.yml`
Configures golangci-lint with enabled linters and settings:

### `.errcheck_excludes.txt`
Excludes common functions from errcheck that are safe to ignore:
- `(net/http.ResponseWriter).Write` - Test response writing
- `(*github.com/spf13/cobra.Command).Help` - Help display in tests
- `(*github.com/spf13/pflag.FlagSet).Set` - Flag setting in tests

Test files are automatically excluded using the `-ignoretests` flag.

```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - revive
    - unconvert
    - unparam
    - gosec
```

## Linting Script

**Location**: `scripts/lint.sh`

This shell script runs all linting checks in order:

1. **gofmt** - Code formatting
2. **go vet** - Static analysis
3. **staticcheck** - Advanced analysis
4. **golangci-lint** - Comprehensive checking (includes style checking via revive)
5. **errcheck** - Error handling
6. **gosec** - Security analysis

### Output

The script provides colored output showing:
- ✅ Green checkmarks for passing checks
- ❌ Red X marks for failing checks
- ⚠️  Yellow warnings for skipped checks (missing tools)
- Summary with pass/fail counts

### Exit Codes

- `0` - All checks passed
- `1` - One or more checks failed

## Installation

### Install All Linters (Recommended)

```bash
make install-linters
```

### Install Individually

```bash
# golangci-lint (recommended - includes many linters)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0

# staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest

# errcheck
go install github.com/kisielk/errcheck@latest

# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

## Running Linting

### Via Makefile (Recommended)

```bash
# Run all linting checks
make lint

# Run individual checks
make fmt      # Format code
make vet      # Run go vet
make ci       # Run all CI checks (fmt + vet + lint + test)

# Pre-commit checks
make pre-commit
```

### Directly

```bash
# Run linting script
./scripts/lint.sh

# Run individual linters
gofmt -l .                  # Check formatting
gofmt -w .                  # Fix formatting
go vet ./...                # Static analysis
staticcheck ./...           # Advanced analysis
errcheck ./...              # Error check
gosec ./...                 # Security check
golangci-lint run ./...     # Comprehensive check (includes style checking)
```

## CI/CD Integration

### GitHub Actions

Linting runs automatically on every pull request before tests:

```yaml
- name: Run linting checks
  run: |
    # Install linters
    curl -sSfL ... | sh -s -- -b $(go env GOPATH)/bin v1.61.0
    go install honnef.co/go/tools/cmd/staticcheck@latest
    go install github.com/kisielk/errcheck@latest
    go install github.com/securego/gosec/v2/cmd/gosec@latest

    # Run linting
    chmod +x scripts/lint.sh
    ./scripts/lint.sh
```

**Workflow**: `.github/workflows/test.yml`

### Build Failure

If linting fails:
1. ❌ The build step is skipped
2. ❌ Tests are not run
3. ❌ The PR cannot be merged (if branch protection is enabled)

The PR comment will show:
```
## ❌ Build and Test Results

Linting failed. Please fix the issues and push again.
```

## Common Issues and Fixes

### Issue: gofmt failed

**Problem**: Code is not formatted correctly

**Fix**:
```bash
gofmt -w .
# or
make fmt
```

### Issue: go vet failed

**Problem**: Static analysis detected issues

**Fix**: Review the output and fix the specific issues mentioned

**Example**:
```
# Problem: Unused variable
var x int  // declared but not used

# Fix: Remove or use the variable
```

### Issue: errcheck failed

**Problem**: Unchecked errors

**Fix**:
```go
// Before (error not checked)
file.Close()

// After (error checked)
if err := file.Close(); err != nil {
    log.Printf("Failed to close file: %v", err)
}
```

### Issue: gosec failed

**Problem**: Security vulnerability detected

**Fix**: Review the specific security issue and apply the recommended fix

**Example**:
```go
// Problem: Weak random number generator
rand.Intn(100)  // G404: Use of weak random number generator

// Fix: Use crypto/rand for security-sensitive operations
```

## Best Practices

### Before Committing

Always run linting before committing:

```bash
make pre-commit
```

Or add a git pre-commit hook:

```bash
#!/bin/bash
# .git/hooks/pre-commit

make fmt
make vet
make lint

if [ $? -ne 0 ]; then
    echo "Linting failed. Please fix issues before committing."
    exit 1
fi
```

### IDE Integration

Configure your IDE to run linters on save:

**VS Code** (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "[go]": {
    "editor.formatOnSave": true
  }
}
```

**GoLand/IntelliJ**:
- Settings → Tools → File Watchers → Add gofmt
- Settings → Tools → External Tools → Add golangci-lint

### Continuous Improvement

The linting configuration may evolve:
- New linters may be added
- Rules may be adjusted based on team feedback
- See `.golangci.yml` for current configuration

## Disabling Specific Checks

### Per Line

```go
//nolint:errcheck
file.Close()

//nolint:gosec
password := os.Getenv("PASSWORD")
```

### Per File

```go
//nolint:all
package unsafe_code
```

### In Configuration

Edit `.golangci.yml`:

```yaml
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck  # Don't check errors in tests
```

**Warning**: Use sparingly and only when necessary!

## Troubleshooting

### Linters not found

**Solution**: Run `make install-linters`

### Script permission denied

**Solution**: `chmod +x scripts/lint.sh`

### Different results locally vs CI

**Possible causes**:
1. Different linter versions
2. Different Go versions
3. Uncommitted changes

**Solution**: Ensure same versions as CI (see workflow file)

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [staticcheck Documentation](https://staticcheck.io/)

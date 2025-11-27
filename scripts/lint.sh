#!/bin/bash

# Go Linting Script
# This script runs comprehensive linting checks on the Go codebase
# Exit code 0 = all checks pass, non-zero = linting failures
#
# Usage: ./lint.sh [--from-ci]
#   --from-ci    Skip golangci-lint check (handled by separate CI pipeline)

set -e

# Parse arguments
isCi=0
for arg in "$@"; do
    if [ "$arg" == "--from-ci" ]; then
        isCi=1
    fi
done

# Add GOPATH/bin to PATH
export PATH="$(go env GOPATH)/bin:$PATH"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}       Go Code Linting and Quality Checks${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Function to run a check
run_check() {
    local name="$1"
    local command="$2"

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    echo -e "${BLUE}Running: ${name}${NC}"

    if eval "$command"; then
        echo -e "${GREEN}✓ ${name} passed${NC}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        echo ""
        return 0
    else
        echo -e "${RED}✗ ${name} failed${NC}"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        echo ""
        return 1
    fi
}

# Track overall failure status
LINT_FAILED=0

# 1. Check if gofmt needs to be run
echo -e "${YELLOW}[1/6] Checking code formatting with gofmt...${NC}"
if ! run_check "gofmt" "test -z \"\$(gofmt -l .)\""; then
    echo -e "${RED}Files need formatting. Run: gofmt -w .${NC}"
    echo "Files that need formatting:"
    gofmt -l .
    LINT_FAILED=1
fi

# 2. Run go vet
echo -e "${YELLOW}[2/6] Running go vet...${NC}"
if ! run_check "go vet" "go vet ./..."; then
    LINT_FAILED=1
fi

# 3. Run staticcheck (if available)
echo -e "${YELLOW}[3/6] Running staticcheck...${NC}"
if command -v staticcheck &> /dev/null; then
    if ! run_check "staticcheck" "staticcheck ./..."; then
        LINT_FAILED=1
    fi
else
    echo -e "${YELLOW}⚠ staticcheck not installed, skipping (install: go install honnef.co/go/tools/cmd/staticcheck@latest)${NC}"
    echo ""
fi

# 4. Run golangci-lint (if available) - comprehensive linter
echo -e "${YELLOW}[4/6] Running golangci-lint...${NC}"
if [ "$isCi" -eq 1 ]; then
    echo -e "${YELLOW}⚠ Skipping golangci-lint (handled by separate CI pipeline)${NC}"
    echo ""
elif command -v golangci-lint &> /dev/null; then
    if ! run_check "golangci-lint" "golangci-lint run ./..."; then
        LINT_FAILED=1
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not installed, skipping${NC}"
    echo -e "${YELLOW}  Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
    echo ""
fi

# 5. Check for common mistakes with errcheck (if available)
echo -e "${YELLOW}[5/6] Running errcheck...${NC}"
if command -v errcheck &> /dev/null; then
    if ! run_check "errcheck" "errcheck -exclude .errcheck_excludes.txt -ignoregenerated -ignoretests ./..."; then
        LINT_FAILED=1
    fi
else
    echo -e "${YELLOW}⚠ errcheck not installed, skipping (install: go install github.com/kisielk/errcheck@latest)${NC}"
    echo ""
fi

# 6. Check for security issues with gosec (if available)
echo -e "${YELLOW}[6/6] Running gosec (security check)...${NC}"
if command -v gosec &> /dev/null; then
    if ! run_check "gosec" "gosec -quiet ./..."; then
        LINT_FAILED=1
    fi
else
    echo -e "${YELLOW}⚠ gosec not installed, skipping (install: go install github.com/securego/gosec/v2/cmd/gosec@latest)${NC}"
    echo ""
fi

# Print summary
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                    Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "Total checks run: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
if [ $FAILED_CHECKS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_CHECKS${NC}"
else
    echo "Failed: $FAILED_CHECKS"
fi
echo ""

if [ $LINT_FAILED -ne 0 ]; then
    echo -e "${RED}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${RED}        ✗ LINTING FAILED - Please fix the issues above${NC}"
    echo -e "${RED}═══════════════════════════════════════════════════════════${NC}"
    exit 1
else
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}        ✓ ALL LINTING CHECKS PASSED!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    exit 0
fi

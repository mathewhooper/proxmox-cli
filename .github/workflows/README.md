# GitHub Actions Workflows

This directory contains GitHub Actions workflows for continuous integration and testing.

## Build and Test Workflow

**File**: `test.yml`

### Triggers

This workflow runs on:
- Pull requests to the `main` branch
- Direct pushes to the `main` branch

### What it Does

The workflow performs the following steps:

1. **Module Verification**
   - Downloads and verifies Go modules
   - Runs `go mod tidy`
   - Checks for uncommitted changes to `go.mod` or `go.sum`

2. **Build Verification**
   - Compiles the project with `go build`
   - Ensures the code compiles successfully before running tests
   - Outputs: `proxmox-cli` binary

3. **Code Quality Checks**
   - Runs `golint` (non-blocking, informational only)
   - Runs `go vet` to detect suspicious code

4. **Testing**
   - Runs all tests with `go test -v ./...`
   - Generates test coverage report
   - Creates JUnit XML report for test results
   - Captures test output for reporting

5. **Reporting**
   - Posts test results as a check on the PR
   - Creates a detailed comment on the PR with:
     - Build status (✅ or ❌)
     - Total number of tests
     - Number of failed tests
     - Test coverage percentage
     - List of failed tests (if any)
     - Full test output (in collapsible section)
     - Coverage report (in collapsible section)
   - Uploads coverage artifacts

### PR Comment Format

When tests pass:
```
## ✅ Build and Test Results

### Summary
- **Build Status**: Success
- **Total Tests**: 76
- **Failed Tests**: 0
- **Test Coverage**: 75.2%

### ✅ All tests passed!

<details>
<summary>Coverage Report</summary>
...
</details>
```

When tests fail:
```
## ❌ Build and Test Results

### Summary
- **Build Status**: Failed
- **Total Tests**: 76
- **Failed Tests**: 3
- **Test Coverage**: 70.1%

### ❌ Failed Tests

TestVMService_StartVM_HttpError
TestNodesService_GetNodeStatus_SessionError
TestClusterService_ListResources_InvalidJSON

<details>
<summary>View full test output</summary>
...
</details>

<details>
<summary>Coverage Report</summary>
...
</details>
```

### Artifacts

The workflow uploads the following artifacts:
- `coverage.out` - Raw coverage data (can be used with `go tool cover`)
- `coverage.txt` - Human-readable coverage report

### Required Permissions

The workflow requires:
- `checks: write` - To post check results
- `pull-requests: write` - To comment on PRs
- `contents: read` - To read repository contents

### Viewing Results

1. **In the PR**: Check the automated comment for a summary
2. **In the Checks tab**: View detailed test results
3. **In Actions tab**: Download coverage artifacts for local analysis

### Local Testing

To run the same checks locally:

```bash
# Verify modules
go mod download
go mod verify
go mod tidy

# Check for changes
git diff go.mod go.sum

# Build
go build -v -o proxmox-cli .

# Run linters
go install golang.org/x/lint/golint@latest
golint ./...
go vet ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...

# View coverage report
go tool cover -func=coverage.out
go tool cover -html=coverage.out  # Opens in browser
```

### Troubleshooting

**Build fails but works locally**:
- Ensure `go.mod` and `go.sum` are committed
- Run `go mod tidy` and commit any changes

**Tests fail in CI but pass locally**:
- Check if tests depend on local environment
- Ensure all test fixtures are committed
- Verify Go version matches (1.24.3)

**Coverage artifacts not available**:
- Check that tests actually ran (even if they failed)
- Verify the workflow completed (check Actions tab)

# Proxmox CLI Tool

## Overview

This project is a Command-Line Interface (CLI) tool designed to interact with a Proxmox server remotely via the web API. Unlike Proxmox's built-in CLI tools (such as `pvesh`, `qm`, `pct`) which are designed to run directly on the Proxmox host, this tool is designed to be executed from an end user's computer, communicating with Proxmox servers over HTTPS.

It provides functionality to log in, validate sessions, and securely communicate with the Proxmox REST API. The tool is written in Go and leverages modular design principles to ensure maintainability and scalability.

## Features

- **Login Command**: Authenticate with the Proxmox server using credentials.
- **Session Validation**: Validate the current session by checking cookies, headers, and payloads.
- **Secure Communication**: Support for SSL certificate trust options.
- **Session Management**: Read and write session data to a file in the user's home directory.
- **SDN Management**: Manage Software Defined Networking (SDN) zones in Proxmox.
  - **Create Zone**: Add a new SDN zone with a specified name and type.
  - **Delete Zone**: Remove an existing SDN zone by name.
  - **Update Zone**: Modify the type of an existing SDN zone.
  - **Apply Configuration**: Apply the configuration of a specified SDN zone.

## Getting Started

### Building and Running

**Using Make (Recommended)**:

```sh
# Install dependencies
go mod tidy

# Build the binary
make build

# Run the CLI with arguments
make run ARGS="login -s proxmox.example.com -u root@pam"

# Or run directly after building
./proxmox-cli login -s <server> -u <username>
./proxmox-cli validate
```

**Using Go commands directly**:

```sh
# Install dependencies
go mod tidy

# Build the binary
go build -o proxmox-cli

# Run the CLI
./proxmox-cli [command]
```

### Available Make Targets

```sh
# Show all available commands
make help

# Development commands
make build           # Build the binary
make test            # Run all tests
make test-coverage   # Run tests with coverage
make fmt             # Format code with gofmt
make vet             # Run go vet
make ci              # Run all CI checks
make pre-commit      # Run pre-commit checks
make clean           # Clean build artifacts
```

**Linting**: This project uses `golangci-lint` for comprehensive code linting.
```sh
# Run linting
golangci-lint run

# Install golangci-lint
# See: https://golangci-lint.run/docs/welcome/install/#binaries
```

## Architecture

The project follows a modular and object-oriented design using Go idioms:

- **Command Layer**: Handles CLI input and output using Cobra, located in the `commands/` directory.
- **Service Layer**: Encapsulates business logic and API communication. For example, authentication logic is implemented in the `AuthService` struct in `services/auth.go`.
- **Config Layer**: Manages configuration, logging, and global flags.

### AuthService Example

Authentication and session validation are handled by the `AuthService` struct. This approach enables encapsulation, testability, and clear dependency management.

```go
import (
    "proxmox-cli/services"
    "proxmox-cli/config"
)

// Initialize AuthService with logger and trust flag
authService := services.NewAuthService(config.Logger, config.Trust)

// Log in to Proxmox
err := authService.LoginToProxmox(server, port, httpScheme, username, password)
if err != nil {
    config.Logger.Error("Login failed: ", err)
}

// Validate session
if authService.ValidateSession() {
    fmt.Println("Session is valid.")
} else {
    fmt.Println("Session is invalid.")
}
```

## Testing

Unit tests are provided for core services, following the same directory structure as the main project. For example, tests for `AuthService` are located in `tests/services/auth_test.go`.

- **Mocking**: The tests use mock implementations for HTTP and session services to isolate logic and ensure test reliability.
- **Test Examples**: See `TestAuthService_LoginToProxmox_Success` and `TestAuthService_ValidateSession_Failure` for sample test cases.
- **Test Coverage**: The project maintains comprehensive test coverage across all services.

### Running Tests Locally

**Using Make (Recommended)**:

```sh
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# View HTML coverage report (after running test-coverage)
go tool cover -html=coverage.out
```

**Using Go commands directly**:

```sh
# Run all tests
go test ./tests/...

# Run tests with verbose output
go test -v ./tests/...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Code Quality and Linting

This project enforces strict code quality standards through linting:

```bash
# Run comprehensive linting
golangci-lint run

# Format code
make fmt

# Run go vet
make vet

# Run all CI checks (format, vet, test)
make ci
```

**Linting Tool**:
- ✅ **golangci-lint** - Comprehensive linting suite that includes:
  - gofmt, goimports - Code formatting
  - go vet - Static analysis
  - staticcheck - Advanced analysis
  - errcheck - Error handling verification
  - gosec - Security vulnerability scanning
  - And many more linters

**Installation**: https://golangci-lint.run/docs/welcome/install/#binaries

Configuration: See [.golangci.yml](.golangci.yml) for the complete linter configuration.

### Continuous Integration

This project uses GitHub Actions for automated testing. On every pull request:

- ✅ **Linting**: Runs golangci-lint for comprehensive code quality checks (separate workflow)
- ✅ **Basic Checks**: Runs go vet for static analysis
- ✅ **Build Verification**: Ensures the project compiles successfully
- ✅ **Test Execution**: Runs all unit tests with coverage reporting
- ✅ **PR Reporting**: Posts detailed results as a comment on the PR

**Pull Request Comments Include**:
- Build status (success/failure)
- Total tests run and number of failures
- List of failed tests (if any)
- Test coverage percentage
- Full test output (in collapsible section)

See [.github/workflows/README.md](.github/workflows/README.md) for more details on the CI/CD pipeline.

## Development Process

This solution was generated using "VIBE" coding, a collaborative and iterative approach that combines automation and human creativity. While this method accelerates development, it may introduce inconsistencies or errors. **Extreme caution is advised when using this tool in production environments.**

## Next Steps

The code will undergo a thorough review to ensure correctness, security, and adherence to best practices. The current implementation is functional but may require refinements based on real-world testing and feedback.

## Acknowledgment

This project was a lot of fun to generate, and the process highlighted the potential of combining automated tools with human oversight to create robust solutions.

## License

This project is licensed under the [MIT License](./LICENSE).

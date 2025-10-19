# Proxmox CLI Tool

## Overview

This project is a Command-Line Interface (CLI) tool designed to interact with a Proxmox server. It provides functionality to log in, validate sessions, and securely communicate with the Proxmox API. The tool is written in Go and leverages modular design principles to ensure maintainability and scalability.

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

To run all tests:

```sh
go test ./tests/...
```

## Development Process

This solution was generated using "VIBE" coding, a collaborative and iterative approach that combines automation and human creativity. While this method accelerates development, it may introduce inconsistencies or errors. **Extreme caution is advised when using this tool in production environments.**

## Next Steps

The code will undergo a thorough review to ensure correctness, security, and adherence to best practices. The current implementation is functional but may require refinements based on real-world testing and feedback.

## Acknowledgment

This project was a lot of fun to generate, and the process highlighted the potential of combining automated tools with human oversight to create robust solutions.

## License

This project is licensed under the [MIT License](./LICENSE).

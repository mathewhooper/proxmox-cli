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

## Development Process

This solution was generated using "VIBE" coding, a collaborative and iterative approach that combines automation and human creativity. While this method accelerates development, it may introduce inconsistencies or errors. **Extreme caution is advised when using this tool in production environments.**

## Next Steps

The code will undergo a thorough review to ensure correctness, security, and adherence to best practices. The current implementation is functional but may require refinements based on real-world testing and feedback.

## Acknowledgment

This project was a lot of fun to generate, and the process highlighted the potential of combining automated tools with human oversight to create robust solutions.

## License

This project is licensed under the [MIT License](./LICENSE).

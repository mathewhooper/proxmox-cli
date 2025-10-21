# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool designed to replicate all API calls available in the Proxmox VE API. The goal is to provide a comprehensive command-line interface for interacting with Proxmox servers, covering authentication, session management, cluster operations, node management, VM/Container operations, storage, networking (SDN), and more.

**Official Proxmox API Documentation**: https://pve.proxmox.com/pve-docs/api-viewer/index.html

The Proxmox VE API is a REST-like API that uses JSON as the primary data format, operates over HTTPS on port 8006, and is formally defined using JSON Schema.

## Development Commands

### Using Make (Recommended)

This project includes a Makefile for common development tasks:

```sh
# Show all available make targets
make help

# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Build the binary
make build

# Build and run the CLI (pass arguments with ARGS)
make run ARGS="login -s proxmox.example.com -u root@pam"

# Format code
make fmt

# Run linting checks
make lint

# Run go vet
make vet

# Run all CI checks (format, lint, vet, test)
make ci

# Pre-commit checks (format, vet, lint)
make pre-commit

# Clean build artifacts
make clean

# Install linting tools
make install-linters
```

### Direct Go Commands

You can also use Go commands directly:

```sh
# Run all tests
go test ./tests/...

# Run tests with verbose output
go test -v ./tests/...

# Run tests and generate JUnit report (used in CI)
go test -v 2>&1 ./... | go-junit-report -set-exit-code > report.xml

# Run a specific test file
go test -v ./tests/services/auth_test.go

# Run a specific test function
go test -v ./tests/services/auth_test.go -run TestAuthService_LoginToProxmox_Success

# Install dependencies
go mod tidy

# Build the binary
go build -o proxmox-cli

# Run the CLI
./proxmox-cli [command]

# Example: Login to Proxmox
./proxmox-cli login -s <server> -u <username>

# Example: Validate session
./proxmox-cli validate
```

## Architecture

The codebase follows a three-layer architecture:

### 1. Command Layer (`commands/`)

- Built using [Cobra](https://github.com/spf13/cobra) for CLI structure
- Commands are organized hierarchically:
  - `login` - Authenticate with Proxmox server
  - `validate` - Validate current session
  - `cluster` - Parent command for cluster operations
    - `sdn` - SDN management subcommands (zone create/delete/update)
- Commands handle CLI flags, user input, and delegate business logic to services

### 2. Service Layer (`services/`)

- **AuthService** (`services/auth.go`): Handles authentication and session validation

  - Manages login flow and session renewal
  - Uses dependency injection for testability
  - Constructor pattern: `NewAuthService()` for production, `NewAuthServiceWithDeps()` for testing

- **SessionService** (`services/session.go`): Manages session persistence

  - Stores session data in `~/.proxmox/session` as JSON
  - Validates all required fields on read
  - Provides typed access to session data (server, port, httpScheme, auth tokens)

- **HttpService** (`services/http.go`): HTTP client abstraction
  - Configurable SSL certificate trust via `--trust` flag
  - Supports custom headers and cookies for Proxmox API authentication
  - All services implement interfaces (`HTTPServiceInterface`, `SessionServiceInterface`) for mocking

### 3. Config Layer (`config/`)

- **Logger** (`config/logger.go`): Global logrus logger instance
  - Default log level: `ErrorLevel` (can be changed with `--show-log` flag)
  - Set dynamically via `config.SetLogLevel()`
- **Flags** (`config/flags.go`): Global flags accessible across commands
  - `Trust` flag for SSL certificate handling

## Key Design Patterns

### Dependency Injection

Services use constructor injection to allow mocking in tests:

```go
// Production usage
authService := services.NewAuthService(config.Logger, config.Trust)

// Testing usage with mocks
authService := services.NewAuthServiceWithDeps(logger, trust, mockHTTP, mockSession)
```

### Session Management

Session data includes:

- Server connection details (server, port, httpScheme)
- Authentication tokens (PVEAuthCookie ticket, CSRFPreventionToken, username)
- Stored in `~/.proxmox/session` file
- Validated on every read to ensure all required fields are present

### Proxmox API Integration

The tool integrates with the Proxmox VE REST API. Understanding the API structure is critical for implementing new commands.

#### API Base Paths

- **JSON API**: `/api2/json/` - Primary REST API endpoint returning JSON responses
- **ExtJS API**: `/api2/extjs/` - Alternative endpoint for ExtJS frontend (some operations like SDN use this)
- **Default Port**: 8006 (HTTPS)
- **Protocol**: HTTPS (HTTP also supported but not recommended)

#### API Endpoint Hierarchy

The API is organized hierarchically around these main resource categories:

1. **`/access`** - Authentication and authorization

   - `/access/ticket` - Create authentication tickets (login)
   - `/access/users` - User management
   - `/access/groups` - Group management
   - `/access/roles` - Role definitions
   - `/access/acl` - Access control lists
   - `/access/domains` - Authentication realms (PAM, LDAP, AD, etc.)

2. **`/cluster`** - Cluster-wide operations

   - `/cluster/resources` - List all cluster resources
   - `/cluster/tasks` - Cluster task history
   - `/cluster/backup` - Backup scheduling
   - `/cluster/ha` - High availability configuration
   - `/cluster/firewall` - Cluster-wide firewall rules
   - `/cluster/sdn` - Software Defined Networking
     - `/cluster/sdn/zones` - SDN zones (VLAN, VXLAN, etc.)
     - `/cluster/sdn/vnets` - Virtual networks
     - `/cluster/sdn/subnets` - Subnet definitions
     - `/cluster/sdn/controllers` - SDN controllers

3. **`/nodes/{node}`** - Node-specific operations

   - `/nodes/{node}/qemu` - Virtual machines (KVM)
     - `/nodes/{node}/qemu/{vmid}/status` - VM status operations (start, stop, reset)
     - `/nodes/{node}/qemu/{vmid}/config` - VM configuration
     - `/nodes/{node}/qemu/{vmid}/snapshot` - VM snapshots
     - `/nodes/{node}/qemu/{vmid}/clone` - VM cloning
     - `/nodes/{node}/qemu/{vmid}/migrate` - VM migration
   - `/nodes/{node}/lxc` - Containers (LXC)
     - Similar structure to qemu for container operations
   - `/nodes/{node}/storage` - Node storage access
   - `/nodes/{node}/network` - Network configuration
   - `/nodes/{node}/tasks` - Node task history
   - `/nodes/{node}/services` - System services
   - `/nodes/{node}/apt` - Package management
   - `/nodes/{node}/certificates` - SSL certificates
   - `/nodes/{node}/firewall` - Node-level firewall

4. **`/storage`** - Storage definitions

   - `/storage/{storage}` - Storage configuration
   - `/storage/{storage}/content` - Storage contents

5. **`/pools`** - Resource pools

   - `/pools/{poolid}` - Pool management

6. **`/version`** - API version information

#### HTTP Methods and Semantics

- **GET** - Retrieve resource information (list or single resource)

  - No CSRF token required
  - Cookie authentication sufficient

- **POST** - Create new resources or trigger actions

  - Requires CSRFPreventionToken header
  - Content-Type: `application/x-www-form-urlencoded; charset=UTF-8`
  - Used for: creating VMs, starting/stopping VMs, creating zones

- **PUT** - Update existing resources

  - Requires CSRFPreventionToken header
  - Content-Type: `application/x-www-form-urlencoded; charset=UTF-8`
  - Used for: modifying VM configs, updating zone settings

- **DELETE** - Remove resources
  - Requires CSRFPreventionToken header
  - Used for: deleting VMs, removing zones, deleting backups

#### Authentication Flow

1. **Initial Login** (POST `/api2/json/access/ticket`):

   ```
   Payload: username=user@realm&password=pass&realm=pam&new-format=1
   Response: {
     "data": {
       "username": "user@pam",
       "ticket": "PVE:user@pam:...",
       "CSRFPreventionToken": "..."
     }
   }
   ```

2. **Authenticated Requests**:

   - **Cookie**: `PVEAuthCookie=<ticket>` (URL-encoded)
   - **Header** (for POST/PUT/DELETE): `CSRFPreventionToken: <token>`

3. **API Token Authentication** (Alternative):
   - Header: `Authorization: PVEAPIToken=USER@REALM!TOKENID=UUID`
   - No CSRF token needed for token-based auth
   - Tokens can have limited permissions and expiration

#### Authentication Realms

Proxmox supports multiple authentication sources:

- **pam** - Linux PAM (default)
- **pve** - Proxmox VE built-in authentication
- **ldap** - LDAP directory
- **ad** - Microsoft Active Directory

#### Request/Response Format

- **Request Body**: `application/x-www-form-urlencoded` (parameter1=value1&parameter2=value2)
- **Response**: JSON with structure:
  ```json
  {
    "data": { ... },  // Response payload
    "success": 1      // Optional success indicator
  }
  ```
- **Error Response**:
  ```json
  {
    "errors": { "field": "error message" },
    "message": "overall error description"
  }
  ```

#### Common API Patterns

1. **List Resources**: `GET /api2/json/{resource}`
2. **Get Single Resource**: `GET /api2/json/{resource}/{id}`
3. **Create Resource**: `POST /api2/json/{resource}` with parameters
4. **Update Resource**: `PUT /api2/json/{resource}/{id}` with parameters
5. **Delete Resource**: `DELETE /api2/json/{resource}/{id}`
6. **Resource Actions**: `POST /api2/json/{resource}/{id}/{action}`

#### Important API Implementation Notes

- **Ticket Rotation**: Authentication tickets are signed by a cluster-wide key that rotates daily
- **Task System**: Long-running operations return a task ID (UPID) that can be monitored via `/nodes/{node}/tasks/{upid}/status`
- **Permissions**: All operations require appropriate permissions based on role-based access control (RBAC)
- **SSL Certificates**: Production should use valid certificates; `--trust` flag bypasses verification for testing
- **pvesh Tool**: Command-line utility for exploring API on Proxmox nodes (`pvesh ls`, `pvesh get`, etc.)

#### Current Implementation Status

Currently implemented:

- Authentication (login, validate session)
- SDN Zone management (create, update, delete zones)

To be implemented (goal is comprehensive coverage):

- Node operations (list, status, configuration)
- VM operations (create, start, stop, snapshot, clone, migrate)
- Container operations (create, start, stop)
- Storage management
- Network configuration
- Backup operations
- User/permission management
- Firewall rules
- High availability configuration
- Task monitoring
- And all other API endpoints

## Implementing New API Endpoints

When adding new Proxmox API endpoints to this CLI tool, follow these patterns:

### Step 1: Research the API Endpoint

1. Consult the official API documentation: https://pve.proxmox.com/pve-docs/api-viewer/index.html
2. Identify the endpoint path, HTTP method, required parameters, and response format
3. Note authentication requirements (cookie + CSRF token for POST/PUT/DELETE)
4. Test the endpoint manually with `curl` or via the Proxmox web UI's API viewer

### Step 2: Add Command Structure

1. Determine command hierarchy (e.g., `proxmox-cli nodes list` or `proxmox-cli vm create`)
2. Create or extend command files in `commands/` directory
3. Follow existing patterns:
   - Use Cobra command builders
   - Define flags for parameters
   - Validate required flags
   - Call service layer functions

Example command structure:

```go
func ListNodesCommand() *cobra.Command {
    var cmd = &cobra.Command{
        Use:   "list",
        Short: "List all cluster nodes",
        Run: func(cmd *cobra.Command, args []string) {
            // Call service layer
            nodes, err := services.ListNodes(config.Logger, config.Trust)
            if err != nil {
                config.Logger.Error("Failed to list nodes: ", err)
                return
            }
            // Output results
            fmt.Printf("Nodes: %+v\n", nodes)
        },
    }
    return cmd
}
```

### Step 3: Implement Service Layer Logic

1. Add functions to appropriate service file or create new service
2. Use existing `HttpService` and `SessionService` for API communication
3. Read session data to get server, auth tokens
4. Build the API URL and payload
5. Call appropriate HTTP method (GET, POST, PUT, DELETE)
6. Parse JSON response
7. Handle errors appropriately

Example service function:

```go
func ListNodes(logger *logrus.Logger, trust bool) ([]Node, error) {
    sessionService, err := NewSessionService(logger)
    if err != nil {
        return nil, err
    }

    sessionData, err := sessionService.ReadSessionFile()
    if err != nil {
        return nil, err
    }

    uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes",
        sessionData.HttpScheme, sessionData.Server, sessionData.Port)

    httpService := NewHttpService(logger, trust)

    // GET requests use cookies only
    cookies := []*http.Cookie{
        {
            Name:  "PVEAuthCookie",
            Value: url.QueryEscape(sessionData.Response.Data.Ticket),
        },
    }

    resp, err := httpService.Get(uri, nil, cookies)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse response
    var result NodeListResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Data, nil
}
```

### Step 4: Add Tests

1. Create test file in `tests/` directory mirroring source structure
2. Use mock services to isolate logic
3. Test success cases, error cases, and edge cases
4. Follow existing test patterns with `mockHTTPService` and `mockSessionService`

Example test:

```go
func TestListNodes_Success(t *testing.T) {
    logger := logrus.New()
    mockHTTP := &mockHTTPService{
        getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
            body := `{"data": [{"node": "pve1", "status": "online"}]}`
            return &http.Response{
                Body: io.NopCloser(strings.NewReader(body)),
            }, nil
        },
    }

    // Test logic here
}
```

### Step 5: Update Documentation

1. Add the new command to this CLAUDE.md file under "Current Implementation Status"
2. Update README.md if the feature is significant
3. Add usage examples

## Testing Strategy

### Test Structure

- Tests mirror source structure: `tests/services/` for `services/`, `tests/commands/` for `commands/`
- Use table-driven tests where appropriate
- Mock all external dependencies (HTTP, file system) using interfaces

### Mock Pattern

Each service interface has a corresponding mock in test files:

- `mockHTTPService` - mocks HTTP operations
- `mockSessionService` - mocks session file operations
- Mocks use function fields to allow per-test behavior customization

### Running Tests in CI

Tests run automatically on PRs to `main` via GitHub Actions (`.github/workflows/test.yml`):

- Uses Go 1.24.3
- Generates JUnit XML report
- Publishes test results to PR

## API Exploration and Development Tips

### Using the API Viewer

The interactive API viewer at https://pve.proxmox.com/pve-docs/api-viewer/index.html is the primary reference:

- Browse the hierarchical endpoint structure
- View required/optional parameters for each endpoint
- See parameter types, descriptions, and valid values
- Check response formats and return values
- Identify which HTTP methods are supported

### Using pvesh on Proxmox Nodes

If you have access to a Proxmox node, use `pvesh` to explore and test API calls:

```bash
# List API paths
pvesh ls /

# Navigate to specific paths
pvesh ls /nodes
pvesh ls /cluster/sdn

# Get specific resource information
pvesh get /nodes/{node}/qemu
pvesh get /version

# Create/modify resources
pvesh create /nodes/{node}/qemu -vmid 100 -name test

# Use help to see parameters
pvesh help create /nodes/{node}/qemu
```

### Testing API Calls with curl

Before implementing in Go, test endpoints with curl:

```bash
# 1. Get authentication ticket
curl -k -d 'username=root@pam&password=yourpass' \
  https://proxmox.example.com:8006/api2/json/access/ticket

# 2. Use ticket for authenticated requests
# GET request (no CSRF needed)
curl -k -b "PVEAuthCookie=PVE%3Aroot..." \
  https://proxmox.example.com:8006/api2/json/nodes

# POST request (CSRF required)
curl -k -X POST \
  -H "CSRFPreventionToken: 12345..." \
  -b "PVEAuthCookie=PVE%3Aroot..." \
  -d "vmid=100&name=test" \
  https://proxmox.example.com:8006/api2/json/nodes/pve/qemu
```

### Common Implementation Gotchas

- **URL Encoding**: The ticket value in cookies must be URL-encoded (`url.QueryEscape()`)
- **ExtJS vs JSON**: Some endpoints use `/api2/extjs/` instead of `/api2/json/` (e.g., SDN zones)
- **CSRF Token**: Required for all POST/PUT/DELETE, but not for GET
- **Task UPIDs**: Long operations return task IDs that need separate polling
- **Node-specific Paths**: Many operations require specifying the node name in the path
- **Parameter Format**: Use `key=value&key2=value2` format, not JSON for request bodies
- **Response Parsing**: Check for both `data` field and potential `errors` field in responses

## Important Notes

- **Project Goal**: Replicate ALL Proxmox API calls - this is a comprehensive API client tool
- **SSL Trust Flag**: The `--trust` / `-t` flag disables SSL certificate verification. Use cautiously.
- **Session File Location**: `~/.proxmox/session` - contains sensitive authentication tokens
- **PAM Realm**: Login currently hardcoded to use PAM realm (`realm=pam` in login payload)
- **SDN Zone Types**: Valid types are: `simple`, `vlan`, `vxlan`, `gre`, `ipsec`, `l2tpv3`, `vxlan-ipsec`, `l2tpv3-ipsec`
- **API Documentation**: Always reference https://pve.proxmox.com/pve-docs/api-viewer/index.html as the source of truth

## Module Dependencies

Key external dependencies (see `go.mod`):

- `github.com/spf13/cobra` - CLI framework
- `github.com/sirupsen/logrus` - Structured logging
- `github.com/stretchr/testify` - Testing assertions
- `golang.org/x/term` - Secure password input

## Quick Reference: Common API Endpoints

This quick reference shows the most commonly used Proxmox API endpoints for implementation:

### Authentication & Authorization

- `POST /api2/json/access/ticket` - Login and get auth ticket
- `GET /api2/json/access/permissions` - Get current user permissions
- `GET /api2/json/access/users` - List users
- `POST /api2/json/access/users` - Create user
- `GET /api2/json/access/groups` - List groups
- `GET /api2/json/access/roles` - List roles
- `GET /api2/json/access/acl` - List ACL entries

### Cluster Operations

- `GET /api2/json/cluster/resources` - List all cluster resources (VMs, containers, storage, nodes)
- `GET /api2/json/cluster/status` - Get cluster status
- `GET /api2/json/cluster/tasks` - List cluster tasks
- `GET /api2/json/cluster/backup` - List backup jobs
- `POST /api2/json/cluster/backup` - Create backup job
- `GET /api2/json/cluster/ha/resources` - List HA resources
- `GET /api2/json/cluster/firewall/rules` - List firewall rules

### Nodes

- `GET /api2/json/nodes` - List all nodes
- `GET /api2/json/nodes/{node}/status` - Get node status
- `GET /api2/json/nodes/{node}/version` - Get node version info
- `GET /api2/json/nodes/{node}/time` - Get node time
- `GET /api2/json/nodes/{node}/services` - List node services
- `POST /api2/json/nodes/{node}/services/{service}/start` - Start service
- `GET /api2/json/nodes/{node}/tasks` - List node tasks
- `GET /api2/json/nodes/{node}/storage` - List storage on node

### Virtual Machines (QEMU)

- `GET /api2/json/nodes/{node}/qemu` - List VMs on node
- `GET /api2/json/nodes/{node}/qemu/{vmid}/config` - Get VM config
- `POST /api2/json/nodes/{node}/qemu` - Create VM
- `PUT /api2/json/nodes/{node}/qemu/{vmid}/config` - Update VM config
- `DELETE /api2/json/nodes/{node}/qemu/{vmid}` - Delete VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/start` - Start VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/stop` - Stop VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/shutdown` - Shutdown VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/reboot` - Reboot VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/reset` - Reset VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/suspend` - Suspend VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/status/resume` - Resume VM
- `GET /api2/json/nodes/{node}/qemu/{vmid}/status/current` - Get current VM status
- `POST /api2/json/nodes/{node}/qemu/{vmid}/clone` - Clone VM
- `POST /api2/json/nodes/{node}/qemu/{vmid}/migrate` - Migrate VM
- `GET /api2/json/nodes/{node}/qemu/{vmid}/snapshot` - List snapshots
- `POST /api2/json/nodes/{node}/qemu/{vmid}/snapshot` - Create snapshot
- `DELETE /api2/json/nodes/{node}/qemu/{vmid}/snapshot/{snapname}` - Delete snapshot

### Containers (LXC)

- `GET /api2/json/nodes/{node}/lxc` - List containers on node
- `GET /api2/json/nodes/{node}/lxc/{vmid}/config` - Get container config
- `POST /api2/json/nodes/{node}/lxc` - Create container
- `PUT /api2/json/nodes/{node}/lxc/{vmid}/config` - Update container config
- `DELETE /api2/json/nodes/{node}/lxc/{vmid}` - Delete container
- `POST /api2/json/nodes/{node}/lxc/{vmid}/status/start` - Start container
- `POST /api2/json/nodes/{node}/lxc/{vmid}/status/stop` - Stop container
- `POST /api2/json/nodes/{node}/lxc/{vmid}/status/shutdown` - Shutdown container
- `GET /api2/json/nodes/{node}/lxc/{vmid}/status/current` - Get current container status
- `POST /api2/json/nodes/{node}/lxc/{vmid}/clone` - Clone container
- `POST /api2/json/nodes/{node}/lxc/{vmid}/migrate` - Migrate container
- `GET /api2/json/nodes/{node}/lxc/{vmid}/snapshot` - List snapshots
- `POST /api2/json/nodes/{node}/lxc/{vmid}/snapshot` - Create snapshot

### Storage

- `GET /api2/json/storage` - List storage
- `GET /api2/json/storage/{storage}` - Get storage config
- `POST /api2/json/storage` - Create storage
- `PUT /api2/json/storage/{storage}` - Update storage config
- `DELETE /api2/json/storage/{storage}` - Delete storage
- `GET /api2/json/nodes/{node}/storage/{storage}/content` - List storage content
- `POST /api2/json/nodes/{node}/storage/{storage}/content` - Allocate storage
- `DELETE /api2/json/nodes/{node}/storage/{storage}/content/{volume}` - Delete volume

### Networking (SDN)

- `GET /api2/json/cluster/sdn/zones` - List SDN zones
- `GET /api2/json/cluster/sdn/zones/{zone}` - Get zone config
- `POST /api2/extjs/cluster/zones` - Create SDN zone (note: uses extjs)
- `PUT /api2/extjs/cluster/zones` - Update SDN zone (note: uses extjs)
- `DELETE /api2/extjs/cluster/zones` - Delete SDN zone (note: uses extjs)
- `GET /api2/json/cluster/sdn/vnets` - List virtual networks
- `POST /api2/json/cluster/sdn/vnets` - Create vnet
- `GET /api2/json/cluster/sdn/controllers` - List SDN controllers

### Pools

- `GET /api2/json/pools` - List resource pools
- `POST /api2/json/pools` - Create pool
- `PUT /api2/json/pools/{poolid}` - Update pool
- `DELETE /api2/json/pools/{poolid}` - Delete pool

### Tasks

- `GET /api2/json/nodes/{node}/tasks` - List node tasks
- `GET /api2/json/nodes/{node}/tasks/{upid}/status` - Get task status
- `GET /api2/json/nodes/{node}/tasks/{upid}/log` - Get task log
- `DELETE /api2/json/nodes/{node}/tasks/{upid}` - Delete task from log

### System Info

- `GET /api2/json/version` - Get API version
- `GET /api2/json/nodes/{node}/version` - Get node version
- `GET /api2/json/nodes/{node}/status` - Get node status
- `GET /api2/json/nodes/{node}/subscription` - Get subscription status

## Resources

- **Official API Docs**: https://pve.proxmox.com/pve-docs/api-viewer/index.html
- **Proxmox Wiki**: https://pve.proxmox.com/wiki/Proxmox_VE_API
- **Admin Guide**: https://pve.proxmox.com/pve-docs/pve-admin-guide.html
- **GitHub Issues**: Report issues or request features for this CLI tool

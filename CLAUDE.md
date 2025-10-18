# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Agent Routing

**MANDATORY**: All prompts and tasks in this project must be routed through specialized agents:

1. **skipgraph-protocol-architect** (inherit)
   - Use for Skip Graph protocol implementation, algorithm correctness, distributed systems properties
   - Required for core Skip Graph operations (search, insert, delete)
   - Validates complexity bounds and theoretical properties
   - Ensures adherence to academic Skip Graph specifications

2. **skipgraph-go-engineer** (inherit)
   - Use for all Go implementation work in this project
   - Required for feature implementation, code review, testing, and refactoring
   - Enforces project-specific conventions (logger injection, testing standards, etc.)
   - Ensures adherence to project architecture and coding standards

**Routing Pattern:**
- Protocol/algorithm questions → `skipgraph-protocol-architect`
- Implementation/coding tasks → `skipgraph-go-engineer`
- Complex tasks requiring both protocol and implementation expertise → Both agents in sequence or parallel as appropriate

## Development Commands

- **Build**: `make build` - Builds the entire project
- **Test**: `make test` - Runs all tests with verbose output
- **Lint**: `make lint` - Runs golangci-lint with auto-fix enabled using configuration at `integration/golangci-lint.yml`
- **Tidy**: `make tidy` - Runs `go mod tidy` to clean up module dependencies
- **Install Tools**: `make install-tools` - Installs required development tools (golangci-lint v1.64.5)
- **Test Coverage**: `go test -cover` - Runs tests with coverage information
- **Single Test**: `go test -v ./path/to/package` - Runs tests for a specific package

All make targets automatically check for Go 1.24.0+ requirement and run `go mod tidy` as needed.

## Architecture

This is a Skip Graph middleware implementation in Go. The system follows a layered architecture:

### Core Components

1. **Skip Graph Node**: Each node has a unique 32-byte identifier and consists of two main components:
   - **Node**: Contains skip graph routing logic
   - **Network**: Provides network communication services between nodes

2. **Model Layer** (`model/`):
   - `skipgraph/identifier.go`: Core 32-byte identifier type with comparison operations
   - `skipgraph/identity.go`: Node identity management
   - `skipgraph/lookupTable.go`: Skip graph lookup table implementation
   - `skipgraph/membershipVector.go`: Membership vector for skip graph levels
   - `messages/message.go`: Message types for inter-node communication
   - `address.go`: Network addressing

3. **Network Layer** (`net/`):
   - `network.go`: Interface definition for network abstraction
   - `network/network.go`: Concrete network implementation
   - `internal/connection.go`: Connection management
   - `internal/connection/grpc_connection.go`: gRPC-based connections
   - `internal/connection/message.proto`: Protocol buffer definitions

4. **Testing Infrastructure** (`unittest/`):
   - `mocknet/`: Mock network implementations for testing
   - `fixtures.go`: Test fixtures and utilities
   - `bytes.go`: Byte manipulation utilities for tests

### Key Design Patterns

- **Interface-based Architecture**: The network layer is defined as an interface, allowing for different network implementations
- **Message Handler Pattern**: The network layer uses configurable message handlers based on message type
- **Identifier-based Routing**: Network communication uses 32-byte identifiers, abstracting away IP addresses
- **Validation-driven Development**: Uses `github.com/go-playground/validator/v10` for struct validation

## Code Conventions

- Follow semantic commit format: `[type][scope] Summary` where type is one of: `feat`, `improve`, `fix`, `cleanup`, `refactor`, `revert`
- Use godoc comments for all public functions, types, and packages
- Add tests for all new functionality - test coverage is required
- Use `gofmt` for code formatting
- Follow Go idiomatic style and conventions
- Keep functions small and focused on single tasks
- Use meaningful variable and function names
- Update godoc comments when modifying existing code
- Always follow the guidelines outlined in `AGENTS.md` for code style, testing, and contribution standards

### Logger Dependency Injection

**MANDATORY**: Components with state or logic must use dependency injection for logging. Loggers must never be initialized internally.

**Scope:**
- **REQUIRED for**: Components with state, business logic, network operations, or side effects
- **NOT REQUIRED for**: Pure model data structures that only hold data without behavior (e.g., `NewSortedEntryList`, simple value objects, DTOs)

**Function Signatures:**
- Logger must always be the **first parameter** in constructors and functions (for components that require logging)
- Exception: Test helpers with `*testing.T` - logger must be the **second parameter** (after `*testing.T`)

**Struct Definitions:**
- Logger must always be the **first field** in struct definitions (for components that require logging)
- Use `zerolog.Logger` for structured logging

**Examples:**

```go
// ✅ CORRECT: Constructor with logger first
func NewWorkerPool(logger zerolog.Logger, queueSize int, workerCount int) *Pool {
    logger = logger.With().Str("component", "worker_pool").Logger()
    // ...
}

// ✅ CORRECT: Struct with logger first
type Pool struct {
    logger      zerolog.Logger  // First field
    workerCount int
    queue       chan modules.Job
    // ...
}

// ✅ CORRECT: Test helper with logger second (after *testing.T)
func NewMockComponent(t *testing.T, logger zerolog.Logger) *MockComponent {
    // ...
}

// ✅ CORRECT: Regular function with logger first
func ProcessMessage(logger zerolog.Logger, msg Message) error {
    logger.Debug().Msg("Processing message")
    // ...
}

// ❌ INCORRECT: Logger not first
func NewPool(queueSize int, logger zerolog.Logger) *Pool { /* ... */ }

// ❌ INCORRECT: Logger not first field
type Pool struct {
    queueSize int
    logger    zerolog.Logger  // Should be first
}

// ❌ INCORRECT: Internal logger initialization
func NewPool(queueSize int) *Pool {
    logger := log.With().Str("component", "pool").Logger()  // Don't do this
    // ...
}
```

**Testing:**
- Use `unittest.Logger(zerolog.TraceLevel)` to create loggers for tests
- Inject logger into constructors during testing, never use global loggers

## Testing Best Practices

- **Use unittest package helpers**: The `unittest` package provides test helpers to avoid boilerplate code
  - `ChannelMustCloseWithinTimeout`: Assert a single channel closes within timeout
  - `ChannelsMustCloseWithinTimeout`: Assert multiple channels close within timeout
  - `CallMustReturnWithinTimeout`: Assert a function returns within timeout
  - `RequireAllReady`: Assert components become ready within default timeout
  - `RequireAllDone`: Assert components become done within default timeout
  - `ChannelMustNotCloseWithinTimeout`: Assert a channel does not close within timeout
- **Avoid redundant patterns**: Never use `select` with `time.After` for channel timeouts - use unittest helpers instead
- **Channel types**: When testing channels, use `chan interface{}` for compatibility with unittest helpers

## Dependencies

- Go 1.23+ (Makefile enforces 1.24.0+)
- `github.com/go-playground/validator/v10` for validation
- `github.com/stretchr/testify` for testing
- golangci-lint v1.64.5 for linting

## Reference Documentation

The `/docs` folder contains reference documentation and blueprints for this project:

- `skip-graphs-journal.pdf`: Core academic paper and technical specifications for Skip Graph implementation
- Additional design documents and implementation blueprints (as they are added)

**When asked to "consult the docs":**
1. Read the relevant documentation in the `/docs` folder
2. Understand the theoretical foundation and design specifications
3. Implement the feature according to the documented specifications
4. Ensure implementation follows both the academic model and project conventions

## Cross-Language Development

This Go implementation is developed in tandem with a reference Rust implementation at `github.com/thep2p/skipgraph-rust`. Both repositories:

- Share the same core architecture (Node/Network pattern)
- Use identical 32-byte identifier systems
- Maintain feature parity across languages
- Are developed under the same `thep2p` organization

**Common Development Patterns:**
- When implementing new features, reference the Rust implementation for design patterns and behavior
- Port Rust features to Go while adapting to Go idioms and conventions
- Maintain consistent API designs across both implementations
- Ensure comprehensive testing for all ported features

**For Feature Implementation:**
1. Study the feature implementation in the Rust version
2. Adapt the design to Go conventions and patterns
3. Implement with proper error handling and validation
4. Write comprehensive tests including unit and integration tests
5. Update documentation and godoc comments
6. Run full test suite and linting to ensure code quality

## Module Path

The module uses `github/thep2p/skipgraph-go` as its path (note: this differs from the GitHub URL which is `github.com/thep2p/skipgraph-go`).
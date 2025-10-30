# Skip Graph Middleware in Go

Skip Graph Middleware is the implementation of a SkipGraph node.
Each node is identified by a unique 32 bytes identifier.
Each node comprises two components, namely, 1) Node and 2) Network.
The node holds the logic for skip graph routing whereas the network provides network communication services between nodes.
The network exposes the necessary interface through which a node can communicate with other nodes in the network.
The node instructs the network to communicate with another node only by specifying the receiver's identifier.
Other network information such as IP address is handled by the network unit and is transparent to the node.

## Usage

To use the Skip Graph Middleware, follow these instructions:

1. Clone the repository:
```bash
git clone https://github.com/thep2p/skipgraph-go.git
cd skipgraph-go
```
2. Build the project:
```bash
make build
```

3. Run the tests:
```bash
make test
```

## Requirements

- Go 1.23 or later

## Testing and Mocking Infrastructure

This project uses a comprehensive mocking infrastructure to enable thorough testing of Skip Graph functionality without requiring real network I/O or complex setup.

### Overview

The testing infrastructure consists of three main components:

1. **Generated Mocks**: Interface mocks generated using [mockery](https://github.com/vektra/mockery) and [testify/mock](https://github.com/stretchr/testify)
2. **MockNet Package**: Complete in-memory network simulation for testing multi-node scenarios
3. **Test Utilities**: Helper functions to eliminate boilerplate and enforce best practices

### Mock Generation

#### Installation

Install the required mocking tools:

```bash
make install-tools
```

This installs:
- mockery v2.43.0 (mock generation)
- golangci-lint v1.64.5 (code linting)

#### Current Mocked Interfaces

- **`core.ImmutableLookupTable`**: Mock for the lookup table interface used in Skip Graph routing

#### Generating Mocks (Manual)

Currently, mocks are manually maintained in `unittest/mock/` due to package loading issues with automatic generation. To generate a new mock:

```bash
mockery --name=InterfaceName --dir=./package/path --output=./unittest/mock --outpkg=mock
```

#### Validating Mocks

Before running tests, validate that required mocks exist:

```bash
make generate-mocks
```

### Using Testify Mocks in Tests

#### Creating a Mock

```go
import (
    "github.com/thep2p/skipgraph-go/unittest/mock"
    "testing"
)

func TestExample(t *testing.T) {
    // Create mock with automatic cleanup and assertion verification
    mockLT := mock.NewImmutableLookupTable(t)

    // ... your test code
}
```

#### Setting Expectations

Use the `.On()` pattern to configure expected method calls:

```go
// Single expectation
mockLT.On("GetEntry", types.DirectionLeft, level).Return(nil, mockError)

// Multiple expectations with different arguments
for level := types.Level(0); level < maxLevel; level++ {
    mockLT.On("GetEntry", types.DirectionLeft, level).Return(&identity, nil)
}
```

**Note**: Mocks automatically verify all expectations were met during test cleanup. If a method is called without being configured via `.On()`, the mock panics with "no return value specified".

#### Complete Example

```go
func TestSearchByIDErrorPropagation(t *testing.T) {
    // Create mock lookup table
    mockLT := mock.NewImmutableLookupTable(t)

    // Configure mock to return error at level 2
    errorAtLevel := types.Level(2)
    mockError := fmt.Errorf("simulated lookup table error")
    mockLT.On("GetEntry", types.DirectionLeft, errorAtLevel).Return(nil, mockError)

    // For other levels, return nil (no neighbor)
    for level := types.Level(0); level < errorAtLevel; level++ {
        mockLT.On("GetEntry", types.DirectionLeft, level).Return(nil, nil)
    }

    // Create node with mock lookup table
    nodeID := unittest.IdentifierFixture(t)
    memVec := unittest.MembershipVectorFixture(t)
    identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))
    node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, mockLT)

    // Test the node's behavior with the mocked lookup table
    result, err := node.SearchByID(targetID)
    // ... assertions
}
```

### MockNet Package

The `unittest/mocknet` package provides a complete in-memory network implementation for testing multi-node Skip Graph scenarios without real network I/O.

#### Components

1. **NetworkStub**: Central router that connects multiple mock networks and routes messages between nodes
2. **MockNetwork**: Implements `net.Network` interface for a single node
3. **MockConduit**: Implements `net.Conduit` interface for sending messages
4. **MockMessageProcessor**: Implements `net.MessageProcessor` interface with custom processing logic

#### Usage Example

```go
func TestMultiNodeCommunication(t *testing.T) {
    // Create network stub (central router)
    stub := mocknet.NewNetworkStub()

    // Create two mock networks with different identifiers
    id1 := unittest.IdentifierFixture(t)
    network1 := stub.NewMockNetwork(t, id1)

    id2 := unittest.IdentifierFixture(t)
    network2 := stub.NewMockNetwork(t, id2)

    // Start networks
    tCtx := unittest.NewMockThrowableContext(t)
    network1.Start(tCtx)
    network2.Start(tCtx)

    // Wait for networks to be ready
    unittest.ChannelsMustCloseWithinTimeout(
        t, 100*time.Millisecond,
        "could not start networks on time",
        network1.Ready(), network2.Ready(),
    )

    // Register message handler at network1
    received := false
    var receivedPayload interface{}
    processor := mocknet.NewMockMessageProcessor(func(channel net.Channel, originID model.Identifier, msg net.Message) {
        received = true
        receivedPayload = msg.Payload
        require.Equal(t, id2, originID)
    })
    _, err := network1.Register(net.TestChannel, processor)
    require.NoError(t, err)

    // Send message from network2 to network1
    conduit, err := network2.Register(net.TestChannel, mocknet.NewMockMessageProcessor(func(channel net.Channel, originID model.Identifier, msg net.Message) {
        // No-op handler for network2
    }))
    require.NoError(t, err)

    msg := unittest.TestMessageFixture(t)
    require.NoError(t, conduit.Send(id1, *msg))

    // Verify message was received
    require.True(t, received)
    require.Equal(t, msg.Payload, receivedPayload)
}
```

### Test Utilities

The `unittest` package provides helpers to eliminate boilerplate and enforce best practices.

#### Channel and Timeout Utilities

**NEVER** use `select` with `time.After` manually. Always use these helpers:

```go
// Assert a single channel closes within timeout
unittest.ChannelMustCloseWithinTimeout(t, ch, 1*time.Second, "channel did not close")

// Assert multiple channels close within timeout
unittest.ChannelsMustCloseWithinTimeout(t, 1*time.Second, "channels did not close", ch1, ch2, ch3)

// Assert a channel does NOT close within timeout
unittest.ChannelMustNotCloseWithinTimeout(t, ch, 1*time.Second, "channel should not close")

// Assert a function returns within timeout
unittest.CallMustReturnWithinTimeout(t, func() { doSomething() }, 1*time.Second, "operation timed out")

// Assert components become ready/done
unittest.RequireAllReady(t, component1, component2)
unittest.RequireAllDone(t, component1, component2)
```

#### Fixture Generators

Use fixtures to generate test data:

```go
// Generate random identifiers and data
id := unittest.IdentifierFixture(t)
memVec := unittest.MembershipVectorFixture(t)
address := unittest.AddressFixture(t)
msg := unittest.TestMessageFixture(t)

// Generate constrained identifiers
greaterID := unittest.IdentifierFixture(t, unittest.WithIdsGreaterThan(baseID))
lesserID := unittest.IdentifierFixture(t, unittest.WithIdsLessThan(baseID))

// Generate full lookup tables
table := unittest.RandomLookupTable(t)
table := unittest.RandomLookupTable(t, unittest.WithIdsGreaterThan(someID))
```

#### Mock Components

```go
// Mock component with basic lifecycle
comp := unittest.NewMockComponent(t)

// Mock component with custom ready/done logic
comp := unittest.NewMockComponentWithLogic(t,
    func() { /* ready logic */ },
    func() { /* done logic */ },
)

// Mock throwable context for error handling tests
ctx := unittest.NewMockThrowableContext(t)
ctx.Cancel() // Cancel the context
```

#### Test Loggers

Always use the unittest logger for testing:

```go
logger := unittest.Logger(zerolog.TraceLevel)
component := NewComponent(logger, otherParams...)
```

### Best Practices

1. **Use Testify Mocks for Interfaces**
   - Prefer generated mocks over manual mocks
   - Use `.On()` to set expectations clearly
   - Let `NewXxx(t)` handle automatic assertion cleanup

2. **Use MockNet for Network Testing**
   - Always use `mocknet.NetworkStub` for multi-node tests
   - Create separate `MockNetwork` instances per node
   - Use `MockMessageProcessor` for custom message handling

3. **Use Unittest Helpers**
   - NEVER use `select` with `time.After` manually
   - Always use `unittest.ChannelMustCloseWithinTimeout` and related helpers
   - Use `unittest.Logger(zerolog.TraceLevel)` for test loggers

4. **Logger Injection in Tests**
   - Always inject logger as first parameter (following project conventions)
   - Use `unittest.Logger(zerolog.TraceLevel)` to create test loggers

5. **Fixture Generation**
   - Use unittest fixtures for all test data
   - Apply constraints with options like `WithIdsGreaterThan`

### Files Organization

```
unittest/
├── mock/                          # Generated mocks (testify)
│   └── immutable_lookup_table.go  # Mock for core.ImmutableLookupTable
├── mocknet/                       # Network mocking infrastructure
│   ├── stub.go                    # NetworkStub (router)
│   ├── underlay.go                # MockNetwork
│   ├── conduit.go                 # MockConduit
│   ├── processor.go               # MockMessageProcessor
│   └── stub_test.go               # Usage examples
├── bytes.go                       # Byte manipulation utilities
├── component.go                   # MockComponent
├── fixtures.go                    # Test data generators
├── logger.go                      # Test logger creation
├── lookup.go                      # Lookup table test helpers
├── throwable.go                   # MockThrowableContext
└── utils.go                       # Channel/timeout utilities
```



# Component Package

The `component` package provides a framework for managing application components with lifecycle awareness and hierarchical organization.

## Overview

The component system enables building modular applications where each `Component` has a well-defined lifecycle with ready and done states. `Component`s can be organized hierarchically using the `Manager`, creating tree-like structures where parent `Component`s manage their children.

## Core Concepts

### Component
A `Component` is a module that:
- Can be **started** via the `Start()` method
- Signals when it's **ready** to process requests
- Signals when it's **done** processing and has shut down

### Manager
A `Manager` is itself a component that can manage other components:
- Starts all its child components when started
- Becomes ready only after ALL child components are ready
- Becomes done only after ALL child components are done
- Can contain other `Manager`s, enabling recursive tree structures

## Architecture

```
                    Root Manager
                           │
                           │ Start()
                           ▼
            ┌──────────────┴──────────────┐
            │                             │
            ▼                             ▼
        Manager A                    Manager B
            │                             │
      ┌─────┴─────┐                 ┌────┴────┐
      │           │                 │         │
      ▼           ▼                 ▼         ▼
  Component1  Component2        Component3   Manager C
                                                   │
                                             ┌─────┴─────┐
                                             │           │
                                             ▼           ▼
                                         Component4  Component5

Ready Flow: Component1,2,3,4,5 → Ready → Managers C,A,B → Ready → Root Ready
Done Flow:  Component1,2,3,4,5 → Done  → Managers C,A,B → Done  → Root Done
```

## Usage Example

```go
// Create sub-managers with their components
networkManager := component.NewManager(
    component.WithComponent(tcpServer),
    component.WithComponent(grpcServer),
)

storageManager := component.NewManager(
    component.WithComponent(database),
    component.WithComponent(cache),
)

// Create root manager with sub-managers
root := component.NewManager(
    component.WithComponent(networkManager),
    component.WithComponent(storageManager),
)

// Start entire tree
ctx := modules.NewThrowableContext(context.Background())
root.Start(ctx)

// Wait for all components to be ready
<-root.Ready()
// Application is now fully initialized

// Shutdown
ctx.Cancel()
<-root.Done()
// Application has cleanly shut down
```

## Key Features

- **Lifecycle Management**: Automatic propagation of start, ready, and done signals through the component tree
- **Recursive Structure**: Managers can contain other Managers, enabling complex hierarchies
- **Synchronization**: Parent components wait for all children before signaling ready/done
- **Thread Safety**: Safe concurrent access to component state
- **Error Propagation**: Irrecoverable errors bubble up through the ThrowableContext

## Embedding Pattern

One of the most powerful patterns with the Manager is **embedding** it within your own structs. This allows any struct to become a Component with full lifecycle management capabilities:

```go
// Your struct embeds Manager to gain Component capabilities
type ConnectionPool struct {
    *component.Manager  // Embedded manager provides Component interface

    // Your struct's specific fields
    maxConnections int
    connections    []net.Conn
}

// Create your component with its own logic and sub-components
func NewConnectionPool(maxConn int) *ConnectionPool {
    pool := &ConnectionPool{
        maxConnections: maxConn,
        connections:    make([]net.Conn, 0),
    }

    // Initialize the embedded Manager with startup/shutdown logic
    pool.Manager = component.NewManager(
        component.WithStartupLogic(func(ctx modules.ThrowableContext) {
            // Your component's initialization logic
            pool.initializeConnections(ctx)
        }),
        component.WithShutdownLogic(func() {
            // Your component's cleanup logic
            pool.closeAllConnections()
        }),
        // Can also manage sub-components
        component.WithComponent(metricsCollector),
        component.WithComponent(healthChecker),
    )

    return pool
}

// Now ConnectionPool IS a Component and can be used in other Managers
appManager := component.NewManager(
    component.WithComponent(NewConnectionPool(100)),
    component.WithComponent(NewHTTPServer()),
)
```

### Benefits of Embedding

1. **Automatic Interface Implementation**: Your struct immediately satisfies the Component interface
2. **Lifecycle Hooks**: Use WithStartupLogic and WithShutdownLogic for initialization and cleanup
3. **Sub-component Management**: Your component can manage its own sub-components
4. **Tree-like Composition**: Components with embedded Managers can be nested infinitely

## Recursive Tree Structure

The embedding pattern enables building complex application architectures where components form a tree:

```go
// NetworkService embeds Manager and manages network components
type NetworkService struct {
    *component.Manager
    config NetworkConfig
}

func NewNetworkService(cfg NetworkConfig) *NetworkService {
    svc := &NetworkService{config: cfg}

    // Create sub-components for network service
    tcpListener := NewTCPListener(cfg.TCPPort)
    grpcServer := NewGRPCServer(cfg.GRPCPort) // Also embeds Manager

    svc.Manager = component.NewManager(
        component.WithStartupLogic(func(ctx modules.ThrowableContext) {
            svc.setupRouting(ctx)
        }),
        component.WithComponent(tcpListener),
        component.WithComponent(grpcServer),
    )

    return svc
}

// StorageService also embeds Manager
type StorageService struct {
    *component.Manager
    dbPath string
}

func NewStorageService(dbPath string) *StorageService {
    svc := &StorageService{dbPath: dbPath}

    // Create sub-components for storage
    database := NewDatabase(dbPath)
    cache := NewCache()
    replicator := NewReplicator() // Could also embed Manager

    svc.Manager = component.NewManager(
        component.WithStartupLogic(func(ctx modules.ThrowableContext) {
            svc.initializeSchema(ctx)
        }),
        component.WithComponent(database),
        component.WithComponent(cache),
        component.WithComponent(replicator),
    )

    return svc
}

// Application root composes all services
type Application struct {
    *component.Manager
}

func NewApplication() *Application {
    app := &Application{}

    app.Manager = component.NewManager(
        component.WithStartupLogic(func(ctx modules.ThrowableContext) {
            log.Println("Application starting...")
        }),
        component.WithShutdownLogic(func() {
            log.Println("Application shutting down...")
        }),
        component.WithComponent(NewNetworkService(networkConfig)),
        component.WithComponent(NewStorageService("/data/db")),
        component.WithComponent(NewMonitoringService()),
    )

    return app
}

// Usage
app := NewApplication()
ctx := modules.NewThrowableContext(context.Background())
app.Start(ctx)  // Starts entire tree recursively

<-app.Ready()   // Waits for entire tree to be ready
// Application fully initialized with all services and sub-components

ctx.Cancel()    // Trigger shutdown
<-app.Done()    // Waits for entire tree to shutdown
// Clean shutdown of all components in reverse order
```

## Dependency Graph Visualization

The embedding pattern creates a dependency graph where parent components depend on their children. Here's how the example above forms a dependency structure:

```
                            Application
                                 │
                                 ├──────────────────┬───────────────────┐
                                 ▼                  ▼                   ▼
                          NetworkService      StorageService    MonitoringService
                                 │                  │                   │
                         ┌───────┴────────┐   ┌────┼────┬──────┐      │
                         ▼                ▼   ▼    ▼    ▼      ▼      ▼
                   TCPListener      GRPCServer  DB Cache Replicator  Metrics
                         │                │
                   ┌─────┴────┐     ┌─────┴─────┐
                   ▼          ▼     ▼           ▼
              ConnPool   RateLimiter  Auth   RequestHandler
```

### Dependency Resolution Order

```
START FLOW (Top-Down):
1. Application.Start()
   ├─ Application.WithStartupLogic()
   ├─ NetworkService.Start()
   │  ├─ NetworkService.WithStartupLogic()
   │  ├─ TCPListener.Start()
   │  │  ├─ ConnPool.Start()
   │  │  └─ RateLimiter.Start()
   │  └─ GRPCServer.Start()
   │     ├─ Auth.Start()
   │     └─ RequestHandler.Start()
   ├─ StorageService.Start()
   │  ├─ StorageService.WithStartupLogic()
   │  ├─ Database.Start()
   │  ├─ Cache.Start()
   │  └─ Replicator.Start()
   └─ MonitoringService.Start()
      └─ Metrics.Start()

READY FLOW (Bottom-Up):
1. Leaf components signal ready:
   - ConnPool.Ready() ✓
   - RateLimiter.Ready() ✓
   - Auth.Ready() ✓
   - RequestHandler.Ready() ✓
   - Database.Ready() ✓
   - Cache.Ready() ✓
   - Replicator.Ready() ✓
   - Metrics.Ready() ✓

2. Parent components signal ready after ALL children:
   - TCPListener.Ready() ✓ (after ConnPool & RateLimiter)
   - GRPCServer.Ready() ✓ (after Auth & RequestHandler)
   - NetworkService.Ready() ✓ (after TCPListener & GRPCServer)
   - StorageService.Ready() ✓ (after DB, Cache & Replicator)
   - MonitoringService.Ready() ✓ (after Metrics)

3. Root signals ready:
   - Application.Ready() ✓ (after all services)

SHUTDOWN FLOW (Top-Down with Bottom-Up completion):
1. Context cancelled → Application.WithShutdownLogic()
2. Shutdown propagates down the tree
3. Each component waits for children to be Done before signaling Done
4. Application.Done() ✓ (after entire tree shutdown)
```

### Example with Timing Visualization

```go
// Simulated component initialization times
type DatabaseComponent struct {
    *component.Manager
}

func NewDatabaseComponent() *DatabaseComponent {
    db := &DatabaseComponent{}
    db.Manager = component.NewManager(
        component.WithStartupLogic(func(ctx modules.ThrowableContext) {
            time.Sleep(2 * time.Second) // Simulate slow DB connection
            log.Println("[DB] Connected to database")
        }),
    )
    return db
}

type CacheComponent struct {
    *component.Manager
}

func NewCacheComponent() *CacheComponent {
    cache := &CacheComponent{}
    cache.Manager = component.NewManager(
        component.WithStartupLogic(func(ctx modules.ThrowableContext) {
            time.Sleep(500 * time.Millisecond) // Fast cache init
            log.Println("[Cache] Initialized in-memory cache")
        }),
    )
    return cache
}

// Timeline visualization of startup:
//
// T+0ms    : Application.Start() called
// T+0ms    : ├─ StorageService.Start() called
// T+0ms    : │  ├─ Database.Start() called (begins 2s init)
// T+0ms    : │  └─ Cache.Start() called (begins 500ms init)
// T+500ms  : │     └─ Cache becomes READY ✓
// T+2000ms : │     └─ Database becomes READY ✓
// T+2000ms : └─ StorageService becomes READY ✓ (all children ready)
// T+2000ms : Application becomes READY ✓ (all services ready)
//
// Total startup time: MAX(child startup times) = 2000ms
```

### Dependency Benefits

1. **Automatic Dependency Management**: Parents automatically wait for dependencies (children)
2. **Parallel Initialization**: Sibling components start concurrently
3. **Graceful Degradation**: If a component fails, parents are aware
4. **Clean Shutdown**: Dependencies are properly cleaned up in reverse order

## Lifecycle Flow in Embedded Components

When using the embedding pattern, the lifecycle flows through the tree structure:

### Start Phase
1. Parent's Start() is called
2. Parent's WithStartupLogic executes
3. All child components' Start() methods are called
4. Process continues recursively down the tree

### Ready Phase
1. Leaf components (no children) signal ready first
2. Parent components wait for ALL children to be ready
3. Once all children are ready, parent signals ready
4. Process bubbles up to the root

### Shutdown Phase
1. Context cancellation triggers shutdown
2. Parent's WithShutdownLogic executes
3. Parent waits for ALL children to signal done
4. Process ensures graceful shutdown of entire tree

## Implementation Notes

- Components are added during manager creation using the options pattern
- Each component can only be started once
- The same component cannot be added multiple times to a manager
- Empty managers (with no components) immediately signal ready and done
- The embedding pattern allows any struct to become a full-fledged Component
- Embedded Managers can manage sub-components, creating recursive tree structures
- Startup and shutdown logic hooks provide clean initialization and cleanup points
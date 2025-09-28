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

## Implementation Notes

- Components are added during manager creation using the options pattern
- Each component can only be started once
- The same component cannot be added multiple times to a manager
- Empty managers (with no components) immediately signal ready and done
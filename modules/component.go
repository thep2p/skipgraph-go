package modules

// ReadyDoneAware is implemented by components that have a ready and done state.
// It provides channels to signal when the component is ready and when it is done.
// A component is considered ready when it has completed its initialization and is ready to process requests.
// A component is considered done when it has completed its shutdown process and is no longer processing requests.
//
// Note that this is a signal-only interface, it indicates the state of the component but does not
// provide any methods to change the state of the component by itself.
type ReadyDoneAware interface {
	// Ready signals that the component is ready to process requests.
	// Ready must be able to be called multiple times, maybe by different entities,
	// it is just an indication of the state of the component.
	// The returned channel must be closed when the component is ready.
	// If the component is not ready, the channel must not be closed.
	// If the component is already ready, the channel must be closed immediately.
	Ready() <-chan interface{}

	// Done signals that the component is done processing requests.
	// Done must be able to be called multiple times, maybe by different entities,
	// it is just an indication of the state of the component.
	// The returned channel must be closed when the component is done.
	// If the component is not done, the channel must not be closed.
	// If the component is already done, the channel must be closed immediately.
	Done() <-chan interface{}
}

// Startable is implemented by components that can be started.
type Startable interface {
	// Start method starts the component.
	// If the component fails to start, it must call ctx.ThrowIrrecoverable(err)
	// to propagate the error up the context chain, and cause the application to terminate.
	// Start must be called only once during the lifetime of the component.
	// Calling Start multiple times must cause a panic.
	Start(ctx ThrowableContext)
}

// Component is a module that can be started and has ready and done states.
type Component interface {
	Startable
	ReadyDoneAware
}

// ComponentManager is a component that can have other components added to it.
// When the ComponentManager is started, it starts all its added components.
// Its ready channel is closed when all its added components are ready.
// Its done channel is closed when all its added components are done.
type ComponentManager interface {
	Component
	// Add method adds a component to the ComponentManager.
	// A ComponentManager can have multiple components added to it.
	// Adding a component after the ComponentManager has been started must cause a panic.
	// Adding the same component multiple times must cause a panic.
	Add(c Component)
}

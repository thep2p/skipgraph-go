package modules

import "context"

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

package component

import (
	"context"
	"fmt"
	"github.com/thep2p/skipgraph-go/modules"
	"sync"
)

type Manager struct {
	components    []modules.Component
	readyChan     chan interface{}               // closed when all components are ready
	doneChan      chan interface{}               // closed when all components are done
	startupLogic  func(modules.ThrowableContext) // startup logic to be executed on Start
	shutdownLogic func()                         // shutdown logic to be executed on Done
	startOnce     sync.Once                      // ensures Start is only called once
}

var _ modules.Component = (*Manager)(nil)

// Option is a functional option for configuring a Manager
type Option func(*Manager)

// WithStartupLogic adds startup logic to be executed when the manager starts
func WithStartupLogic(logic func(modules.ThrowableContext)) Option {
	return func(m *Manager) {
		m.startupLogic = logic
	}
}

// WithShutdownLogic adds shutdown logic to be executed when the manager stops
func WithShutdownLogic(logic func()) Option {
	return func(m *Manager) {
		m.shutdownLogic = logic
	}
}

// WithComponent adds a component to be managed
func WithComponent(c modules.Component) Option {
	return func(m *Manager) {
		// Check if component already exists
		for _, existing := range m.components {
			if existing == c {
				panic("cannot add the same component to Manager multiple times")
			}
		}
		m.components = append(m.components, c)
	}
}

// NewManager creates a new Manager with the given options
func NewManager(opts ...Option) *Manager {
	m := &Manager{
		components: make([]modules.Component, 0),
		readyChan:  make(chan interface{}),
		doneChan:   make(chan interface{}),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Start(ctx modules.ThrowableContext) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	started := false

	// Ensure Start is only called once even if called concurrently
	m.startOnce.Do(
		func() {
			started = true // Indicate that Start has been called
			if m.startupLogic != nil {
				m.startupLogic(ctx)
			}

			// Start all components in parallel
			var wg sync.WaitGroup
			wg.Add(len(m.components))
			for _, c := range m.components {
				go func(component modules.Component) {
					defer wg.Done()
					component.Start(ctx)
				}(c)
			}

			// Wait for all components to be started
			go func() {
				wg.Wait()
				// Now wait for all components to be ready
				m.waitForReady(ctx)
			}()

			// Wait for all components to be done in a separate goroutine
			go m.waitForDone(ctx)
		},
	)

	if !started {
		ctx.ThrowIrrecoverable(fmt.Errorf("start called multiple times on Manager"))
	}
}

func (m *Manager) Ready() <-chan interface{} {
	return m.readyChan
}

func (m *Manager) Done() <-chan interface{} {
	return m.doneChan
}

func (m *Manager) waitForReady(ctx context.Context) {
	// If no components, immediately close ready channel
	if len(m.components) == 0 {
		close(m.readyChan)
		return
	}

	// Wait for all components to be ready in parallel
	var wg sync.WaitGroup
	wg.Add(len(m.components))

	for _, component := range m.components {
		go func(c modules.Component) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return // Exit if context is done
			case <-c.Ready():
				// Component is ready
			}
		}(component)
	}

	// Wait for all goroutines to complete or context to be done
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return // Exit if context is done
	case <-done:
		// All components are ready
		close(m.readyChan)
	}
}

func (m *Manager) waitForDone(ctx context.Context) {
	<-ctx.Done()
	if m.shutdownLogic != nil {
		m.shutdownLogic()
	}

	// If no components, immediately close done channel
	if len(m.components) == 0 {
		close(m.doneChan)
		return
	}

	// Wait for all components to be done in parallel
	var wg sync.WaitGroup
	wg.Add(len(m.components))

	for _, component := range m.components {
		go func(c modules.Component) {
			defer wg.Done()
			<-c.Done()
		}(component)
	}

	// Wait for all components to finish
	wg.Wait()

	// Close the done channel
	close(m.doneChan)
}

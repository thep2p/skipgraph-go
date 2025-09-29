package component

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/thep2p/skipgraph-go/modules"
	"sync"
)

type Manager struct {
	logger        zerolog.Logger // structured logger for component events
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
// Args:
//   - logger: zerolog.Logger for logging component lifecycle events
//   - opts: variadic options for configuring the manager
//
// Returns initialized manager (not started).
func NewManager(logger zerolog.Logger, opts ...Option) *Manager {
	logger = logger.With().
		Str("component", "manager").
		Logger()

	m := &Manager{
		components: make([]modules.Component, 0),
		readyChan:  make(chan interface{}),
		doneChan:   make(chan interface{}),
		logger:     logger,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Start(ctx modules.ThrowableContext) {
	select {
	case <-ctx.Done():
		m.logger.Debug().Msg("Start called but context already done")
		return
	default:
	}

	started := false

	// Ensure Start is only called once even if called concurrently
	m.startOnce.Do(
		func() {
			started = true // Indicate that Start has been called
			m.logger.Info().Int("component_count", len(m.components)).Msg("Starting component manager")

			if m.startupLogic != nil {
				m.logger.Debug().Msg("Executing startup logic")
				m.startupLogic(ctx)
			}

			// Start all components in parallel
			m.logger.Debug().Msg("Starting all components in parallel")
			var wg sync.WaitGroup
			wg.Add(len(m.components))
			for i, c := range m.components {
				go func(index int, component modules.Component) {
					defer wg.Done()
					m.logger.Debug().Int("component_index", index).Msg("Starting component")
					component.Start(ctx)
				}(i, c)
			}

			// Wait for all components to be started
			go func() {
				wg.Wait()
				m.logger.Debug().Msg("All components started, waiting for ready")
				// Now wait for all components to be ready
				m.waitForReady(ctx)
			}()

			// Wait for all components to be done in a separate goroutine
			go m.waitForDone(ctx)
			m.logger.Debug().Msg("Component manager startup initiated")
		},
	)

	if !started {
		m.logger.Error().Msg("Component manager start called multiple times")
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
		m.logger.Debug().Msg("No components to wait for, marking ready immediately")
		close(m.readyChan)
		return
	}

	m.logger.Debug().Int("component_count", len(m.components)).Msg("Waiting for all components to be ready")

	// Wait for all components to be ready in parallel
	var wg sync.WaitGroup
	wg.Add(len(m.components))

	for i, component := range m.components {
		go func(index int, c modules.Component) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				m.logger.Warn().Int("component_index", index).Msg("Context cancelled while waiting for component ready")
				return // Exit if context is done
			case <-c.Ready():
				m.logger.Debug().Int("component_index", index).Msg("Component ready")
				// Component is ready
			}
		}(i, component)
	}

	// Wait for all goroutines to complete or context to be done
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		m.logger.Warn().Msg("Context cancelled while waiting for components to be ready")
		return // Exit if context is done
	case <-done:
		// All components are ready
		m.logger.Info().Msg("All components ready")
		close(m.readyChan)
	}
}

func (m *Manager) waitForDone(ctx context.Context) {
	<-ctx.Done()
	m.logger.Info().Msg("Context cancelled, initiating shutdown")

	if m.shutdownLogic != nil {
		m.logger.Debug().Msg("Executing shutdown logic")
		m.shutdownLogic()
	}

	// If no components, immediately close done channel
	if len(m.components) == 0 {
		m.logger.Debug().Msg("No components to wait for, marking done immediately")
		close(m.doneChan)
		return
	}

	m.logger.Debug().Int("component_count", len(m.components)).Msg("Waiting for all components to be done")

	// Wait for all components to be done in parallel
	var wg sync.WaitGroup
	wg.Add(len(m.components))

	for i, component := range m.components {
		go func(index int, c modules.Component) {
			defer wg.Done()
			m.logger.Debug().Int("component_index", index).Msg("Waiting for component to be done")
			<-c.Done()
			m.logger.Debug().Int("component_index", index).Msg("Component done")
		}(i, component)
	}

	// Wait for all components to finish
	wg.Wait()

	// Close the done channel
	m.logger.Info().Msg("All components done, shutdown complete")
	close(m.doneChan)
}

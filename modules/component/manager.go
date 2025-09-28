package component

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/thep2p/skipgraph-go/modules"
)

type Manager struct {
	components    []modules.Component
	started       chan interface{}               // closed when Start is called (the manager has started)
	readyChan     chan interface{}               // closed when all components are ready
	doneChan      chan interface{}               // closed when all components are done
	startupLogic  func(modules.ThrowableContext) // startup logic to be executed on Start
	shutdownLogic func()                         // shutdown logic to be executed on Done
	logger        zerolog.Logger                 // structured logger for component events
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
		started:    make(chan interface{}),
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
	case <-m.started:
		m.logger.Error().Msg("Component manager already started")
		ctx.ThrowIrrecoverable(fmt.Errorf("component manager already started"))
	default:
		m.logger.Info().Int("component_count", len(m.components)).Msg("Starting component manager")
		close(m.started)
		if m.startupLogic != nil {
			m.logger.Debug().Msg("Executing startup logic")
			m.startupLogic(ctx)
		}
		// Start all components
		m.logger.Debug().Msg("Starting all components")
		for i, c := range m.components {
			m.logger.Debug().Int("component_index", i).Msg("Starting component")
			c.Start(ctx)
		}

		// Wait for all components to be ready in a separate goroutine
		go m.waitForReady(ctx)

		// Wait for all components to be done in a separate goroutine
		go m.waitForDone(ctx)
		m.logger.Debug().Msg("Component manager startup initiated")
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
	// Wait for all components to be ready
	for i, component := range m.components {
		select {
		case <-ctx.Done():
			m.logger.Warn().Msg("Context cancelled while waiting for components to be ready")
			return // Exit if context is done
		case <-component.Ready():
			m.logger.Debug().Int("component_index", i).Msg("Component ready")
			// Component is ready, continue to next
		}
	}

	// Close the ready channel
	m.logger.Info().Msg("All components ready")
	close(m.readyChan)
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
	// Wait for all components to be done
	for i, component := range m.components {
		m.logger.Debug().Int("component_index", i).Msg("Waiting for component to be done")
		<-component.Done()
		m.logger.Debug().Int("component_index", i).Msg("Component done")
	}

	// Close the done channel
	m.logger.Info().Msg("All components done, shutdown complete")
	close(m.doneChan)
}

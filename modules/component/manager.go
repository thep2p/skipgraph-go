package component

import (
	"context"
	"fmt"
	"github.com/thep2p/skipgraph-go/modules"
	"sync"
)

type Manager struct {
	mu            sync.RWMutex
	components    []modules.Component
	started       chan interface{}               // closed when Start is called (the manager has started)
	readyChan     chan interface{}               // closed when all components are ready
	doneChan      chan interface{}               // closed when all components are done
	startupLogic  func(modules.ThrowableContext) // startup logic to be executed on Start
	shutdownLogic func()                         // shutdown logic to be executed on Done
	readyOnce     sync.Once
	doneOnce      sync.Once
}

var _ modules.ComponentManager = (*Manager)(nil)

func NewManager() *Manager {
	return &Manager{
		components: make([]modules.Component, 0),
		readyChan:  make(chan interface{}),
		doneChan:   make(chan interface{}),
		started:    make(chan interface{}),
	}
}

func (m *Manager) Start(ctx modules.ThrowableContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return
	case <-m.started:
		ctx.ThrowIrrecoverable(fmt.Errorf("component manager already started"))
	}

	// Start all components
	for _, c := range m.components {
		c.Start(ctx)
	}

	// Wait for all components to be ready in a separate goroutine
	go m.waitForReady(ctx)

	// Wait for all components to be done in a separate goroutine
	go m.waitForDone()
}

func (m *Manager) Ready() <-chan interface{} {
	return m.readyChan
}

func (m *Manager) Done() <-chan interface{} {
	return m.doneChan
}

func (m *Manager) Add(c modules.Component) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.started:
		panic("cannot add component to Manager after it has started")
	default:
	}

	// Check if component already exists
	for _, component := range m.components {
		if component == c {
			panic("cannot add the same component to Manager multiple times")
		}
	}

	m.components = append(m.components, c)
}

func (m *Manager) waitForReady(ctx context.Context) {
	m.mu.RLock()
	components := make([]modules.Component, len(m.components))
	copy(components, m.components)
	m.mu.RUnlock()

	// If no components, immediately close ready channel
	if len(components) == 0 {
		m.readyOnce.Do(
			func() {
				close(m.readyChan)
			},
		)
		return
	}

	// Wait for all components to be ready
	for _, component := range components {
		select {
		case <-ctx.Done():
			return // Exit if context is done
		case <-component.Ready():
			// Component is ready, continue to next
		}
	}

	// Close the ready channel exactly once
	m.readyOnce.Do(
		func() {
			close(m.readyChan)
		},
	)
}

func (m *Manager) waitForDone() {
	m.mu.RLock()
	components := make([]modules.Component, len(m.components))
	copy(components, m.components)
	m.mu.RUnlock()

	// If no components, immediately close done channel
	if len(components) == 0 {
		m.doneOnce.Do(
			func() {
				close(m.doneChan)
			},
		)
		return
	}

	// Wait for all components to be done
	for _, component := range components {
		<-component.Done()
	}

	// Close the done channel exactly once
	m.doneOnce.Do(
		func() {
			close(m.doneChan)
		},
	)
}

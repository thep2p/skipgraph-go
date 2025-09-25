package component

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/modules"
)

// LifecycleManager is a simple implementation of modules.Component that tracks the lifecycle states
// (started, ready, done) using channels.
// Its primary purpose is to facilitate monitoring the lifecycle of components without boilerplate code.
type LifecycleManager struct {
	started       chan interface{}               // closed when Start is called
	readyChan     chan interface{}               // closed when SignalReady is called
	doneChan      chan interface{}               // closed when SignalDone is called
	startupLogic  func(modules.ThrowableContext) // startup logic to be executed on Start
	shutdownLogic func()                         // shutdown logic to be executed on Done
}

var _ modules.Component = (*LifecycleManager)(nil)

func NewLifecycleTracker(startupLogic func(modules.ThrowableContext), shutdownLogic func()) *LifecycleManager {
	return &LifecycleManager{
		started:       make(chan interface{}),
		readyChan:     make(chan interface{}),
		doneChan:      make(chan interface{}),
		startupLogic:  startupLogic,
		shutdownLogic: shutdownLogic,
	}
}

// Start marks the component as started by closing the started channel.
// It throws an irrecoverable error if called multiple times, as each component should only be started once.
func (l *LifecycleManager) Start(ctx modules.ThrowableContext) {
	select {
	case <-ctx.Done():
		return
	case <-l.started:
		ctx.ThrowIrrecoverable(fmt.Errorf("component already started"))
	default:
		close(l.started)
		l.startupLogic(ctx)
		close(l.readyChan)
		go func() {
			<-ctx.Done()
			l.shutdownLogic()
			close(l.doneChan)
		}()
	}
}

// Ready returns a channel that is closed when the component is ready.
// The channel is closed when SignalReady is called.
func (l *LifecycleManager) Ready() <-chan interface{} {
	return l.readyChan
}

// Done returns a channel that is closed when the component is done.
// The channel is closed when SignalDone is called.
func (l *LifecycleManager) Done() <-chan interface{} {
	return l.doneChan
}

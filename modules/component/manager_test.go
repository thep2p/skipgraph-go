package component_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
	"github.com/thep2p/skipgraph-go/modules/component"
	"github.com/thep2p/skipgraph-go/modules/throwable"
	"github.com/thep2p/skipgraph-go/unittest"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := component.NewManager()
	require.NotNil(t, manager)

	// Manager should not be nil and should implement ComponentManager interface
	var _ modules.ComponentManager = manager
}

func TestManager_Add(t *testing.T) {
	manager := component.NewManager()
	component1 := unittest.NewMockComponent(t)
	component2 := unittest.NewMockComponent(t)

	// Add components before starting
	require.NotPanics(
		t, func() {
			manager.Add(component1)
			manager.Add(component2)
		},
	)
}

func TestManager_Add_SameComponentTwice_ShouldPanic(t *testing.T) {
	manager := component.NewManager()
	component1 := unittest.NewMockComponent(t)

	manager.Add(component1)

	// Adding same component twice should panic
	require.Panics(
		t, func() {
			manager.Add(component1)
		},
	)
}

func TestManager_Add_AfterStart_ShouldPanic(t *testing.T) {
	manager := component.NewManager()
	component1 := unittest.NewMockComponent(t)
	component2 := unittest.NewMockComponent(t)

	manager.Add(component1)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Adding component after start should panic
	require.Panics(
		t, func() {
			manager.Add(component2)
		},
	)
}

func TestManager_Start_CalledTwice_ShouldPanic(t *testing.T) {
	manager := component.NewManager()
	component1 := unittest.NewMockComponent(t)

	manager.Add(component1)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Starting twice should panic
	require.Panics(
		t, func() {
			manager.Start(ctx)
		},
	)
}

func TestManager_Ready_Done_WaitsForAllComponents(t *testing.T) {
	manager := component.NewManager()
	component1 := unittest.NewMockComponent(t)
	component2 := unittest.NewMockComponent(t)

	manager.Add(component1)
	manager.Add(component2)

	ctx, cancel := context.WithCancel(context.Background())
	tCtx := throwable.NewContext(ctx)
	manager.Start(tCtx)

	// Components should be started and ready
	unittest.ChannelMustCloseWithinTimeout(t, component1.Ready(), 100*time.Millisecond, "component1 was not started")
	unittest.ChannelMustCloseWithinTimeout(t, component2.Ready(), 100*time.Millisecond, "component2 was not started")

	// When all components are ready, manager should be ready
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager was not ready after all components were ready")

	// Cancel context to signal components to be done
	cancel()

	// Manager should wait for all components to be done
	unittest.ChannelMustCloseWithinTimeout(t, component1.Done(), 200*time.Millisecond, "component1 was not done")
	unittest.ChannelMustCloseWithinTimeout(t, component2.Done(), 200*time.Millisecond, "component2 was not done")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 300*time.Millisecond, "manager was not done after all components were done")
}

func TestManager_WithNoComponents(t *testing.T) {
	manager := component.NewManager()

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// With no components, manager should be ready and done immediately
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager was not ready immediately")

	// Expected - since there are no components, manager should be done immediately
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager was not done immediately")
}

func TestManager_MultipleCalls(t *testing.T) {
	manager := component.NewManager()
	component1 := unittest.NewMockComponent(t)

	manager.Add(component1)

	ctx, cancel := context.WithCancel(context.Background())
	tCtx := throwable.NewContext(ctx)
	manager.Start(tCtx)

	// Multiple calls to Ready() and Done() should return the same channel
	readyChan1 := manager.Ready()
	readyChan2 := manager.Ready()
	require.Equal(t, readyChan1, readyChan2)

	doneChan1 := manager.Done()
	doneChan2 := manager.Done()
	require.Equal(t, doneChan1, doneChan2)

	// Both channels should be closed when component is ready and done
	unittest.ChannelMustCloseWithinTimeout(t, readyChan1, 100*time.Millisecond, "ready channel was not closed")
	unittest.ChannelMustCloseWithinTimeout(t, readyChan2, 100*time.Millisecond, "ready channel was not closed")

	// Cancel context to signal component to be done
	cancel()

	unittest.ChannelMustCloseWithinTimeout(t, doneChan1, 200*time.Millisecond, "done channel was not closed")
	unittest.ChannelMustCloseWithinTimeout(t, doneChan2, 200*time.Millisecond, "done channel was not closed")
}

func TestManager_NotReadyWhenComponentBlocksOnReady(t *testing.T) {
	manager := component.NewManager()

	// Create a blocking ready signal
	readySignal := make(chan struct{})
	blockingComponent := unittest.NewMockComponentWithLogic(
		t,
		func() { <-readySignal }, // Block until signal is sent
		func() {},                // Non-blocking done logic
	)

	// Create a non-blocking component for comparison
	normalComponent := unittest.NewMockComponent(t)

	manager.Add(blockingComponent)
	manager.Add(normalComponent)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tCtx := throwable.NewContext(ctx)
	manager.Start(tCtx)

	// Normal component should be ready quickly
	unittest.ChannelMustCloseWithinTimeout(t, normalComponent.Ready(), 100*time.Millisecond, "normal component should be ready")

	// Manager should NOT be ready while blocking component is not ready
	select {
	case <-manager.Ready():
		require.Fail(t, "manager should not be ready while blocking component is not ready")
	case <-time.After(200 * time.Millisecond):
		// Expected: manager is blocked
	}

	// Release the blocking component
	close(readySignal)

	// Now both components and manager should be ready
	unittest.ChannelMustCloseWithinTimeout(t, blockingComponent.Ready(), 100*time.Millisecond, "blocking component should be ready after signal")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager should be ready after all components are ready")
}

func TestManager_NotDoneWhenComponentBlocksOnDone(t *testing.T) {
	manager := component.NewManager()

	// Create a blocking done signal
	doneSignal := make(chan struct{})
	blockingComponent := unittest.NewMockComponentWithLogic(
		t,
		func() {},               // Non-blocking ready logic
		func() { <-doneSignal }, // Block until signal is sent
	)

	// Create a non-blocking component for comparison
	normalComponent := unittest.NewMockComponent(t)

	manager.Add(blockingComponent)
	manager.Add(normalComponent)

	ctx, cancel := context.WithCancel(context.Background())
	tCtx := throwable.NewContext(ctx)
	manager.Start(tCtx)

	// Both components and manager should be ready quickly
	unittest.ChannelMustCloseWithinTimeout(t, blockingComponent.Ready(), 100*time.Millisecond, "blocking component should be ready")
	unittest.ChannelMustCloseWithinTimeout(t, normalComponent.Ready(), 100*time.Millisecond, "normal component should be ready")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager should be ready")

	// Cancel context to trigger done state
	cancel()

	// Normal component should be done quickly
	unittest.ChannelMustCloseWithinTimeout(t, normalComponent.Done(), 200*time.Millisecond, "normal component should be done")

	// Manager should NOT be done while blocking component is not done
	select {
	case <-manager.Done():
		require.Fail(t, "manager should not be done while blocking component is not done")
	case <-time.After(300 * time.Millisecond):
		// Expected: manager is blocked
	}

	// Release the blocking component
	close(doneSignal)

	// Now blocking component and manager should be done
	unittest.ChannelMustCloseWithinTimeout(t, blockingComponent.Done(), 100*time.Millisecond, "blocking component should be done after signal")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager should be done after all components are done")
}

func TestManager_NeverReadyWhenContextCancelledDuringStartup(t *testing.T) {
	manager := component.NewManager()

	// Create a component that blocks on ready
	readySignal := make(chan struct{})
	slowComponent := unittest.NewMockComponentWithLogic(
		t,
		func() { <-readySignal }, // Block until signal is sent
		func() {},                // Non-blocking done logic
	)

	// Create another component that becomes ready quickly
	fastComponent := unittest.NewMockComponent(t)

	manager.Add(slowComponent)
	manager.Add(fastComponent)

	ctx, cancel := context.WithCancel(context.Background())
	tCtx := throwable.NewContext(ctx)
	manager.Start(tCtx)

	// Fast component should be ready quickly
	unittest.ChannelMustCloseWithinTimeout(t, fastComponent.Ready(), 100*time.Millisecond, "fast component should be ready")

	// Cancel the context while the slow component is still not ready
	cancel()

	// Manager should never become ready because context was cancelled
	// during the waitForReady loop before all components were ready
	select {
	case <-manager.Ready():
		require.Fail(t, "manager should never become ready when context is cancelled during startup")
	case <-time.After(500 * time.Millisecond):
		// Expected: manager never becomes ready
	}

	// Even if we release the slow component now, manager should still not be ready
	// because the waitForReady goroutine already returned early
	close(readySignal)
	unittest.ChannelMustCloseWithinTimeout(t, slowComponent.Ready(), 100*time.Millisecond, "slow component should be ready after signal")

	// Manager should still not be ready even after all components are ready
	select {
	case <-manager.Ready():
		require.Fail(t, "manager should still not be ready even after all components become ready")
	case <-time.After(200 * time.Millisecond):
		// Expected: manager remains not ready because waitForReady returned early
	}

	// Manager should eventually be done since context was cancelled
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 500*time.Millisecond, "manager should be done after context cancellation")
}

package component

import (
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
	"github.com/thep2p/skipgraph-go/unittest"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	require.NotNil(t, manager)

	// Manager should not be nil and should implement ComponentManager interface
	var _ modules.ComponentManager = manager
}

func TestManager_Add(t *testing.T) {
	manager := NewManager()
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
	manager := NewManager()
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
	manager := NewManager()
	component1 := unittest.NewMockComponent(t)
	component2 := unittest.NewMockComponent(t)

	manager.Add(component1)

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

	// Adding component after start should panic
	require.Panics(
		t, func() {
			manager.Add(component2)
		},
	)
}

func TestManager_Start_CalledTwice_ShouldPanic(t *testing.T) {
	manager := NewManager()
	component1 := unittest.NewMockComponent(t)

	manager.Add(component1)

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

	// Starting twice should trigger ThrowIrrecoverable
	var thrownErr error
	ctx2 := unittest.NewMockThrowableContext(
		t, unittest.WithThrowLogic(
			func(err error) {
				thrownErr = err
			},
		),
	)

	manager.Start(ctx2)

	require.NotNil(t, thrownErr)
	require.Contains(t, thrownErr.Error(), "already started")
}

func TestManager_Ready_Done_WaitsForAllComponents(t *testing.T) {
	manager := NewManager()

	// Create components with controlled done behavior
	doneSignal1 := make(chan struct{})
	doneSignal2 := make(chan struct{})

	component1 := unittest.NewMockComponentWithLogic(
		t,
		func() {},                // Non-blocking ready
		func() { <-doneSignal1 }, // Block until signal
	)

	component2 := unittest.NewMockComponentWithLogic(
		t,
		func() {},                // Non-blocking ready
		func() { <-doneSignal2 }, // Block until signal
	)

	manager.Add(component1)
	manager.Add(component2)

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

	// Components should be started and ready
	unittest.ChannelMustCloseWithinTimeout(t, component1.Ready(), 100*time.Millisecond, "component1 was not started")
	unittest.ChannelMustCloseWithinTimeout(t, component2.Ready(), 100*time.Millisecond, "component2 was not started")

	// When all components are ready, manager should be ready
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager was not ready after all components were ready")

	// Cancel context to signal components to start shutdown
	ctx.Cancel()

	// Verify manager is NOT done while components are still blocking
	require.Eventually(t, func() bool {
		select {
		case <-manager.Done():
			// Manager should NOT be done yet
			return false
		default:
			// Good, manager is waiting
			return true
		}
	}, 200*time.Millisecond, 10*time.Millisecond, "manager should wait for components to be done")

	// Release component1
	close(doneSignal1)
	unittest.ChannelMustCloseWithinTimeout(t, component1.Done(), 100*time.Millisecond, "component1 should be done after signal")

	// Manager should still be waiting for component2
	select {
	case <-manager.Done():
		require.Fail(t, "manager should not be done while component2 is still running")
	default:
		time.Sleep(100 * time.Millisecond)
		// Verify again that manager is still not done
		select {
		case <-manager.Done():
			require.Fail(t, "manager should not be done while component2 is still running")
		default:
			// Good, manager is still waiting
		}
	}

	// Release component2
	close(doneSignal2)
	unittest.ChannelMustCloseWithinTimeout(t, component2.Done(), 100*time.Millisecond, "component2 should be done after signal")

	// Now manager should be done
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager should be done after all components are done")
}

func TestManager_WithNoComponents(t *testing.T) {
	manager := NewManager()

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

	// With no components, manager should be ready immediately
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager was not ready immediately")

	// Cancel context to trigger done
	ctx.Cancel()

	// Expected - since there are no components, manager should be done immediately after cancellation
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager was not done immediately")
}

func TestManager_MultipleCalls(t *testing.T) {
	manager := NewManager()
	component1 := unittest.NewMockComponent(t)

	manager.Add(component1)

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

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
	ctx.Cancel()

	unittest.ChannelMustCloseWithinTimeout(t, doneChan1, 200*time.Millisecond, "done channel was not closed")
	unittest.ChannelMustCloseWithinTimeout(t, doneChan2, 200*time.Millisecond, "done channel was not closed")
}

func TestManager_NotReadyWhenComponentBlocksOnReady(t *testing.T) {
	manager := NewManager()

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

	ctx := unittest.NewMockThrowableContext(t)
	defer ctx.Cancel()
	manager.Start(ctx)

	// Normal component should be ready quickly
	unittest.ChannelMustCloseWithinTimeout(t, normalComponent.Ready(), 100*time.Millisecond, "normal component should be ready")

	// Manager should NOT be ready while blocking component is not ready
	select {
	case <-manager.Ready():
		require.Fail(t, "manager should not be ready while blocking component is not ready")
	default:
		time.Sleep(200 * time.Millisecond)
		// Verify again that manager is still not ready
		select {
		case <-manager.Ready():
			require.Fail(t, "manager should not be ready while blocking component is not ready")
		default:
			// Expected: manager is blocked
		}
	}

	// Release the blocking component
	close(readySignal)

	// Now both components and manager should be ready
	unittest.ChannelMustCloseWithinTimeout(t, blockingComponent.Ready(), 100*time.Millisecond, "blocking component should be ready after signal")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager should be ready after all components are ready")
}

func TestManager_NotDoneWhenComponentBlocksOnDone(t *testing.T) {
	manager := NewManager()

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

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

	// Both components and manager should be ready quickly
	unittest.ChannelMustCloseWithinTimeout(t, blockingComponent.Ready(), 100*time.Millisecond, "blocking component should be ready")
	unittest.ChannelMustCloseWithinTimeout(t, normalComponent.Ready(), 100*time.Millisecond, "normal component should be ready")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager should be ready")

	// Cancel context to trigger done state
	ctx.Cancel()

	// Normal component should be done quickly
	unittest.ChannelMustCloseWithinTimeout(t, normalComponent.Done(), 200*time.Millisecond, "normal component should be done")

	// Manager should NOT be done while blocking component is not done
	select {
	case <-manager.Done():
		require.Fail(t, "manager should not be done while blocking component is not done")
	default:
		time.Sleep(300 * time.Millisecond)
		// Verify again that manager is still not done
		select {
		case <-manager.Done():
			require.Fail(t, "manager should not be done while blocking component is not done")
		default:
			// Expected: manager is blocked
		}
	}

	// Release the blocking component
	close(doneSignal)

	// Now blocking component and manager should be done
	unittest.ChannelMustCloseWithinTimeout(t, blockingComponent.Done(), 100*time.Millisecond, "blocking component should be done after signal")
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager should be done after all components are done")
}

func TestNewManagerWithLifecycle(t *testing.T) {
	t.Run("successful lifecycle with startup and shutdown logic", func(t *testing.T) {
		var startupCalled, shutdownCalled bool

		manager := NewManagerWithLifecycle(
			func(ctx modules.ThrowableContext) {
				startupCalled = true
			},
			func() {
				shutdownCalled = true
			},
		)

		ctx := unittest.NewMockThrowableContext(t)
		manager.Start(ctx)

		require.True(t, startupCalled, "startup logic should be called")
		unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager should be ready")

		require.False(t, shutdownCalled, "shutdown logic should not be called yet")

		ctx.Cancel()

		unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager should be done")
		require.True(t, shutdownCalled, "shutdown logic should be called after context cancellation")
	})

	t.Run("with nil startup and shutdown logic", func(t *testing.T) {
		manager := NewManagerWithLifecycle(nil, nil)
		ctx := unittest.NewMockThrowableContext(t)

		require.NotPanics(t, func() {
			manager.Start(ctx)
		}, "should handle nil startup/shutdown logic gracefully")

		unittest.ChannelMustCloseWithinTimeout(t, manager.Ready(), 100*time.Millisecond, "manager should be ready")
		ctx.Cancel()
		unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 100*time.Millisecond, "manager should be done")
	})

	t.Run("double start should trigger ThrowIrrecoverable", func(t *testing.T) {
		manager := NewManagerWithLifecycle(
			func(ctx modules.ThrowableContext) {},
			func() {},
		)

		ctx := unittest.NewMockThrowableContext(t)
		manager.Start(ctx)

		var thrownErr error
		ctx2 := unittest.NewMockThrowableContext(
			t, unittest.WithThrowLogic(
				func(err error) {
					thrownErr = err
				},
			),
		)

		manager.Start(ctx2)

		require.NotNil(t, thrownErr)
		require.Contains(t, thrownErr.Error(), "already started")
	})
}

func TestManager_NeverReadyWhenContextCancelledDuringStartup(t *testing.T) {
	manager := NewManager()

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

	ctx := unittest.NewMockThrowableContext(t)
	manager.Start(ctx)

	// Fast component should be ready quickly
	unittest.ChannelMustCloseWithinTimeout(t, fastComponent.Ready(), 100*time.Millisecond, "fast component should be ready")

	// Cancel the context while the slow component is still not ready
	ctx.Cancel()

	// Manager should never become ready because context was cancelled
	// during the waitForReady loop before all components were ready
	select {
	case <-manager.Ready():
		require.Fail(t, "manager should never become ready when context is cancelled during startup")
	default:
		time.Sleep(500 * time.Millisecond)
		// Verify again that manager is still not ready
		select {
		case <-manager.Ready():
			require.Fail(t, "manager should never become ready when context is cancelled during startup")
		default:
			// Expected: manager never becomes ready
		}
	}

	// Even if we release the slow component now, manager should still not be ready
	// because the waitForReady goroutine already returned early
	close(readySignal)
	unittest.ChannelMustCloseWithinTimeout(t, slowComponent.Ready(), 100*time.Millisecond, "slow component should be ready after signal")

	// Manager should still not be ready even after all components are ready
	select {
	case <-manager.Ready():
		require.Fail(t, "manager should still not be ready even after all components become ready")
	default:
		time.Sleep(200 * time.Millisecond)
		// Verify again that manager is still not ready
		select {
		case <-manager.Ready():
			require.Fail(t, "manager should still not be ready even after all components become ready")
		default:
			// Expected: manager remains not ready because waitForReady returned early
		}
	}

	// Manager should eventually be done since context was cancelled
	unittest.ChannelMustCloseWithinTimeout(t, manager.Done(), 500*time.Millisecond, "manager should be done after context cancellation")
}

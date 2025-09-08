package component_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github/thep2p/skipgraph-go/modules"
	"github/thep2p/skipgraph-go/modules/component"
	"github/thep2p/skipgraph-go/modules/throwable"
	"github/thep2p/skipgraph-go/unittest"
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

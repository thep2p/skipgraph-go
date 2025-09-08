package component_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github/thep2p/skipgraph-go/modules"
	"github/thep2p/skipgraph-go/modules/component"
	"github/thep2p/skipgraph-go/modules/throwable"
	"sync"
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
	component1 := NewMockComponent()
	component2 := NewMockComponent()

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
	component1 := NewMockComponent()

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
	component1 := NewMockComponent()
	component2 := NewMockComponent()

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

func TestManager_Start_CallsStartOnAllComponents(t *testing.T) {
	manager := component.NewManager()
	component1 := NewMockComponent()
	component2 := NewMockComponent()

	manager.Add(component1)
	manager.Add(component2)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Verify all components were started
	require.True(t, component1.IsStartCalled())
	require.True(t, component2.IsStartCalled())
	require.Equal(t, 1, component1.GetStartCallCount())
	require.Equal(t, 1, component2.GetStartCallCount())
}

func TestManager_Start_CalledTwice_ShouldPanic(t *testing.T) {
	manager := component.NewManager()
	component1 := NewMockComponent()

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

func TestManager_Ready_WaitsForAllComponents(t *testing.T) {
	manager := component.NewManager()
	component1 := NewMockComponent()
	component2 := NewMockComponent()

	manager.Add(component1)
	manager.Add(component2)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Manager should not be ready yet since components aren't ready
	select {
	case <-manager.Ready():
		t.Fatal("Manager should not be ready yet")
	case <-time.After(50 * time.Millisecond):
		// Expected - manager not ready yet
	}

	// Mark first component as ready
	component1.MarkReady()

	// Manager should still not be ready since not all components are ready
	select {
	case <-manager.Ready():
		t.Fatal("Manager should not be ready yet - component2 not ready")
	case <-time.After(50 * time.Millisecond):
		// Expected - manager not ready yet
	}

	// Mark second component as ready
	component2.MarkReady()

	// Now manager should be ready
	select {
	case <-manager.Ready():
		// Expected - manager is now ready
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Manager should be ready now")
	}
}

func TestManager_Done_WaitsForAllComponents(t *testing.T) {
	manager := component.NewManager()
	component1 := NewMockComponent()
	component2 := NewMockComponent()

	manager.Add(component1)
	manager.Add(component2)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Mark components as ready first
	component1.MarkReady()
	component2.MarkReady()

	// Wait for manager to be ready
	<-manager.Ready()

	// Manager should not be done yet since components aren't done
	select {
	case <-manager.Done():
		t.Fatal("Manager should not be done yet")
	case <-time.After(50 * time.Millisecond):
		// Expected - manager not done yet
	}

	// Mark first component as done
	component1.MarkDone()

	// Manager should still not be done since not all components are done
	select {
	case <-manager.Done():
		t.Fatal("Manager should not be done yet - component2 not done")
	case <-time.After(50 * time.Millisecond):
		// Expected - manager not done yet
	}

	// Mark second component as done
	component2.MarkDone()

	// Now manager should be done
	select {
	case <-manager.Done():
		// Expected - manager is now done
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Manager should be done now")
	}
}

func TestManager_WithNoComponents(t *testing.T) {
	manager := component.NewManager()

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// With no components, manager should be ready and done immediately
	select {
	case <-manager.Ready():
		// Expected - manager with no components is immediately ready
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Manager with no components should be ready immediately")
	}

	select {
	case <-manager.Done():
		// Expected - manager with no components is immediately done
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Manager with no components should be done immediately")
	}
}

func TestManager_ComponentsWithDelays(t *testing.T) {
	manager := component.NewManager()

	// Component 1: Fast startup, slow to be ready
	component1 := NewMockComponentWithDelays(10*time.Millisecond, 100*time.Millisecond, 150*time.Millisecond)
	// Component 2: Slow startup, fast to be ready
	component2 := NewMockComponentWithDelays(50*time.Millisecond, 20*time.Millisecond, 200*time.Millisecond)

	manager.Add(component1)
	manager.Add(component2)

	ctx := throwable.NewContext(context.Background())
	startTime := time.Now()
	manager.Start(ctx)

	// Wait for manager to be ready - should take time for slowest component to be ready
	<-manager.Ready()
	readyTime := time.Now()

	// Should take at least the time for the slowest component to be ready
	// Component 1: 10ms start + 100ms ready = 110ms
	// Component 2: 50ms start + 20ms ready = 70ms
	// Manager should be ready after ~110ms
	require.True(t, readyTime.Sub(startTime) >= 100*time.Millisecond)

	// Wait for manager to be done
	<-manager.Done()
	doneTime := time.Now()

	// Should take time for the slowest component to be done
	// Component 1: 10ms start + 100ms ready + 150ms done = 260ms
	// Component 2: 50ms start + 20ms ready + 200ms done = 270ms
	// Manager should be done after ~270ms
	require.True(t, doneTime.Sub(startTime) >= 250*time.Millisecond)
}

func TestManager_Ready_MultipleCalls(t *testing.T) {
	manager := component.NewManager()
	component1 := NewMockComponent()

	manager.Add(component1)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Multiple calls to Ready() should return the same channel
	readyChan1 := manager.Ready()
	readyChan2 := manager.Ready()
	require.Equal(t, readyChan1, readyChan2)

	component1.MarkReady()

	// Both channels should be closed
	<-readyChan1
	<-readyChan2
}

func TestManager_Done_MultipleCalls(t *testing.T) {
	manager := component.NewManager()
	component1 := NewMockComponent()

	manager.Add(component1)

	ctx := throwable.NewContext(context.Background())
	manager.Start(ctx)

	// Multiple calls to Done() should return the same channel
	doneChan1 := manager.Done()
	doneChan2 := manager.Done()
	require.Equal(t, doneChan1, doneChan2)

	component1.MarkReady()
	component1.MarkDone()

	// Both channels should be closed
	<-doneChan1
	<-doneChan2
}

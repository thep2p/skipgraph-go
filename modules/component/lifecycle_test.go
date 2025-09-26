package component

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
	"github.com/thep2p/skipgraph-go/unittest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLifecycleManager_ImplementsComponent(t *testing.T) {
	lm := NewLifecycleTracker(
		func(ctx modules.ThrowableContext) {},
		func() {},
	)

	var _ modules.Component = lm
	assert.NotNil(t, lm, "LifecycleManager should not be nil")
}

func TestLifecycleManager_Start(t *testing.T) {
	t.Run(
		"successful start", func(t *testing.T) {
			var startupCalled atomic.Bool
			var shutdownCalled atomic.Bool

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					startupCalled.Store(true)
				},
				func() {
					shutdownCalled.Store(true)
				},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			assert.True(t, startupCalled.Load(), "startup logic should be called")

			unittest.RequireAllReady(t, lm)

			assert.False(t, shutdownCalled.Load(), "shutdown logic should not be called yet")

			ctx.Cancel()

			unittest.RequireAllDone(t, lm)

			assert.True(t, shutdownCalled.Load(), "shutdown logic should be called after context done")
		},
	)

	t.Run(
		"start with nil functions", func(t *testing.T) {
			lm := NewLifecycleTracker(nil, nil)
			ctx := unittest.NewMockThrowableContext(t)

			assert.NotPanics(
				t, func() {
					lm.Start(ctx)
				}, "Start should handle nil startup/shutdown functions gracefully",
			)

			unittest.RequireAllReady(t, lm)

			ctx.Cancel()

			unittest.RequireAllDone(t, lm)
		},
	)

	t.Run(
		"double start throws irrecoverable error", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			var throwCalled atomic.Bool
			var thrownErr error

			ctx := unittest.NewMockThrowableContext(
				t, unittest.WithThrowLogic(
					func(err error) {
						throwCalled.Store(true)
						thrownErr = err
					},
				),
			)

			lm.Start(ctx)

			lm.Start(ctx)

			assert.True(t, throwCalled.Load(), "ThrowIrrecoverable should be called on double start")
			assert.Error(t, thrownErr, "error should be thrown")
			assert.Contains(t, thrownErr.Error(), "already started", "error should indicate component already started")
		},
	)

	t.Run(
		"start after context cancelled", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					require.Fail(t, "startup logic should not be called when context is cancelled")
				},
				func() {
					require.Fail(t, "shutdown logic should not be called when start is skipped")
				},
			)

			ctx := unittest.NewMockThrowableContext(t)
			ctx.Cancel()

			lm.Start(ctx)

			select {
			case <-lm.Ready():
				t.Fatal("ready channel should not be closed when context is cancelled before start")
			case <-time.After(50 * time.Millisecond):
			}
		},
	)
}

func TestLifecycleManager_Ready(t *testing.T) {
	t.Run(
		"ready channel closed after start", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			select {
			case <-lm.Ready():
				t.Fatal("ready channel should not be closed before start")
			default:
			}

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			unittest.RequireAllReady(t, lm)
		},
	)

	t.Run(
		"ready channel is idempotent", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			readyChan1 := lm.Ready()
			readyChan2 := lm.Ready()
			readyChan3 := lm.Ready()

			assert.Equal(t, readyChan1, readyChan2, "Ready() should return the same channel")
			assert.Equal(t, readyChan2, readyChan3, "Ready() should return the same channel")
			unittest.ChannelsMustCloseWithinTimeout(
				t, unittest.DefaultReadyDoneTimeout, "all ready channels must be closed", readyChan1, readyChan2,
				readyChan3,
			)
		},
	)
}

func TestLifecycleManager_Done(t *testing.T) {
	t.Run(
		"done channel closed after context cancellation", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			doneChan := lm.Done()

			select {
			case <-doneChan:
				t.Fatal("done channel should not be closed before start")
			default:
			}

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			select {
			case <-doneChan:
				t.Fatal("done channel should not be closed before context cancellation")
			default:
			}

			ctx.Cancel()

			select {
			case <-doneChan:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("done channel should be closed after context cancellation")
			}
		},
	)

	t.Run(
		"done channel is idempotent", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			doneChan1 := lm.Done()
			doneChan2 := lm.Done()
			doneChan3 := lm.Done()

			assert.Equal(t, doneChan1, doneChan2, "Done() should return the same channel")
			assert.Equal(t, doneChan2, doneChan3, "Done() should return the same channel")

			ctx.Cancel()

			select {
			case <-doneChan1:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("all done channels should be closed")
			}

			select {
			case <-doneChan2:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("all done channels should be closed")
			}

			select {
			case <-doneChan3:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("all done channels should be closed")
			}
		},
	)
}

func TestLifecycleManager_StartupPanic(t *testing.T) {
	t.Run(
		"panic in startup logic propagates", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					panic("startup panic")
				},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(t)

			assert.Panics(
				t, func() {
					lm.Start(ctx)
				}, "panic in startup logic should propagate",
			)
		},
	)
}

func TestLifecycleManager_ShutdownPanic(t *testing.T) {
	t.Run(
		"panic in shutdown logic does not affect done channel", func(t *testing.T) {
			var panicOccurred atomic.Bool

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {
					panicOccurred.Store(true)
					panic("shutdown panic")
				},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			ctx.Cancel()

			time.Sleep(50 * time.Millisecond)

			assert.True(t, panicOccurred.Load(), "shutdown panic should have occurred")

			select {
			case <-lm.Done():
			case <-time.After(100 * time.Millisecond):
				t.Fatal("done channel should still be closed despite panic in shutdown")
			}
		},
	)
}

func TestLifecycleManager_ConcurrentOperations(t *testing.T) {
	t.Run(
		"concurrent Ready() calls", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					time.Sleep(10 * time.Millisecond)
				},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(t)

			var wg sync.WaitGroup
			const numGoroutines = 100

			readyChannels := make([]<-chan interface{}, numGoroutines)

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(idx int) {
					defer wg.Done()
					readyChannels[idx] = lm.Ready()
				}(i)
			}

			lm.Start(ctx)

			wg.Wait()

			for i, ch := range readyChannels {
				select {
				case <-ch:
				case <-time.After(100 * time.Millisecond):
					t.Fatalf("ready channel %d should be closed", i)
				}
			}
		},
	)

	t.Run(
		"concurrent Done() calls", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {
					time.Sleep(10 * time.Millisecond)
				},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			var wg sync.WaitGroup
			const numGoroutines = 100

			doneChannels := make([]<-chan interface{}, numGoroutines)

			wg.Add(numGoroutines)
			for i := 0; i < numGoroutines; i++ {
				go func(idx int) {
					defer wg.Done()
					doneChannels[idx] = lm.Done()
				}(i)
			}

			ctx.Cancel()

			wg.Wait()

			for i, ch := range doneChannels {
				select {
				case <-ch:
				case <-time.After(200 * time.Millisecond):
					t.Fatalf("done channel %d should be closed", i)
				}
			}
		},
	)
}

func TestLifecycleManager_StartupError(t *testing.T) {
	t.Run(
		"error during startup triggers ThrowIrrecoverable", func(t *testing.T) {
			expectedErr := errors.New("startup failed")
			var thrownErr error

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					ctx.ThrowIrrecoverable(expectedErr)
				},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(
				t, unittest.WithThrowLogic(
					func(err error) {
						thrownErr = err
					},
				),
			)

			lm.Start(ctx)

			assert.Equal(t, expectedErr, thrownErr, "the thrown error should match the expected error")
		},
	)
}

func TestLifecycleManager_LongRunningOperations(t *testing.T) {
	t.Run(
		"long-running startup does not block Ready", func(t *testing.T) {
			startupCompleted := make(chan struct{})

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					time.Sleep(50 * time.Millisecond)
					close(startupCompleted)
				},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			select {
			case <-lm.Ready():
			case <-time.After(100 * time.Millisecond):
				t.Fatal("ready channel should be closed after startup completes")
			}

			select {
			case <-startupCompleted:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("startup should complete")
			}
		},
	)

	t.Run(
		"long-running shutdown does not block Done", func(t *testing.T) {
			shutdownCompleted := make(chan struct{})

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {
					time.Sleep(50 * time.Millisecond)
					close(shutdownCompleted)
				},
			)

			ctx := unittest.NewMockThrowableContext(t)
			lm.Start(ctx)

			ctx.Cancel()

			select {
			case <-lm.Done():
			case <-time.After(100 * time.Millisecond):
				t.Fatal("done channel should be closed after shutdown completes")
			}

			select {
			case <-shutdownCompleted:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("shutdown should complete")
			}
		},
	)
}

func TestLifecycleManager_ContextPropagation(t *testing.T) {
	t.Run(
		"context is properly propagated to startup logic", func(t *testing.T) {
			var capturedCtx modules.ThrowableContext

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					capturedCtx = ctx
				},
				func() {},
			)

			originalCtx := unittest.NewMockThrowableContext(t)
			lm.Start(originalCtx)

			assert.Equal(t, originalCtx, capturedCtx, "context should be propagated to startup logic")
		},
	)

	t.Run(
		"context cancellation triggers shutdown", func(t *testing.T) {
			shutdownCalled := make(chan struct{})

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {
					close(shutdownCalled)
				},
			)

			parentCtx, cancel := context.WithCancel(context.Background())
			ctx := unittest.NewMockThrowableContext(t)
			ctx.Context = parentCtx

			lm.Start(ctx)

			cancel()

			select {
			case <-shutdownCalled:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("shutdown should be called after context cancellation")
			}

			select {
			case <-lm.Done():
			case <-time.After(100 * time.Millisecond):
				t.Fatal("done channel should be closed after context cancellation")
			}
		},
	)
}

func TestLifecycleManager_EdgeCases(t *testing.T) {
	t.Run(
		"multiple starts with different contexts", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			ctx1 := unittest.NewMockThrowableContext(t)
			lm.Start(ctx1)

			var throwCalled atomic.Bool
			ctx2 := unittest.NewMockThrowableContext(
				t, unittest.WithThrowLogic(
					func(err error) {
						throwCalled.Store(true)
					},
				),
			)

			lm.Start(ctx2)

			assert.True(t, throwCalled.Load(), "second start should throw error regardless of context")
		},
	)

	t.Run(
		"ready and done channels behavior before start", func(t *testing.T) {
			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {},
				func() {},
			)

			readyChan := lm.Ready()
			doneChan := lm.Done()

			select {
			case <-readyChan:
				t.Fatal("ready channel should not be closed before start")
			case <-doneChan:
				t.Fatal("done channel should not be closed before start")
			default:
			}
		},
	)
}

func TestLifecycleManager_RealWorldScenarios(t *testing.T) {
	t.Run(
		"simulating a server component lifecycle", func(t *testing.T) {
			serverStarted := make(chan struct{})
			serverStopped := make(chan struct{})

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					close(serverStarted)
				},
				func() {
					time.Sleep(10 * time.Millisecond)
					close(serverStopped)
				},
			)

			ctx := unittest.NewMockThrowableContext(t)

			go func() {
				lm.Start(ctx)
			}()

			select {
			case <-serverStarted:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("server should start")
			}

			select {
			case <-lm.Ready():
			case <-time.After(100 * time.Millisecond):
				t.Fatal("server should be ready")
			}

			ctx.Cancel()

			select {
			case <-serverStopped:
			case <-time.After(100 * time.Millisecond):
				t.Fatal("server should stop")
			}

			select {
			case <-lm.Done():
			case <-time.After(100 * time.Millisecond):
				t.Fatal("server should be done")
			}
		},
	)

	t.Run(
		"component with initialization failure", func(t *testing.T) {
			initErr := errors.New("failed to initialize database connection")
			var capturedErr error

			lm := NewLifecycleTracker(
				func(ctx modules.ThrowableContext) {
					ctx.ThrowIrrecoverable(initErr)
				},
				func() {},
			)

			ctx := unittest.NewMockThrowableContext(
				t, unittest.WithThrowLogic(
					func(err error) {
						capturedErr = err
					},
				),
			)

			lm.Start(ctx)

			assert.Equal(t, initErr, capturedErr, "initialization error should be propagated")

			// Note: In a real scenario, ThrowIrrecoverable would terminate the process,
			// but in tests it just records the error. The component will still become ready
			// unless the actual process terminates.
			select {
			case <-lm.Ready():
				// This is expected behavior in test context
			case <-time.After(50 * time.Millisecond):
				t.Fatal("ready channel should be closed even when ThrowIrrecoverable is called in test context")
			}
		},
	)
}

package throwable_test

import (
	"errors"
	"testing"
	"time"

	"github.com/thep2p/skipgraph-go/modules/throwable"
	"github.com/thep2p/skipgraph-go/unittest"
)

// TestThrowIrrecoverableWithMockContext verifies that ThrowIrrecoverable
// properly delegates to parent contexts that implement the ThrowableContext
// interface (not just concrete *Context types).
func TestThrowIrrecoverableWithMockContext(t *testing.T) {
	t.Parallel()

	errThrown := make(chan error, 1)
	mockCtx := unittest.NewMockThrowableContext(t, unittest.WithThrowLogic(func(err error) {
		errThrown <- err
		close(errThrown)
	}))

	// Wrap the mock context with throwable.NewContext
	ctx := throwable.NewContext(mockCtx)

	// Call ThrowIrrecoverable - this should delegate to mockCtx's implementation
	testErr := errors.New("test error")
	ctx.ThrowIrrecoverable(testErr)

	// Verify the custom throw logic was called (not panic)
	select {
	case receivedErr := <-errThrown:
		if receivedErr.Error() != testErr.Error() {
			t.Fatalf("expected error %q, got %q", testErr.Error(), receivedErr.Error())
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected custom throw logic to be called, but it was not")
	}
}

// TestThrowIrrecoverableWithNestedContext verifies that ThrowIrrecoverable
// propagates through nested throwable.Context instances and eventually
// reaches a custom ThrowableContext implementation at the root.
func TestThrowIrrecoverableWithNestedContext(t *testing.T) {
	t.Parallel()

	errThrown := make(chan error, 1)
	mockCtx := unittest.NewMockThrowableContext(t, unittest.WithThrowLogic(func(err error) {
		errThrown <- err
		close(errThrown)
	}))

	// Create nested throwable contexts
	ctx1 := throwable.NewContext(mockCtx)
	ctx2 := throwable.NewContext(ctx1)
	ctx3 := throwable.NewContext(ctx2)

	// Call ThrowIrrecoverable on the deepest context
	testErr := errors.New("nested test error")
	ctx3.ThrowIrrecoverable(testErr)

	// Verify it propagated all the way to the mock
	select {
	case receivedErr := <-errThrown:
		if receivedErr.Error() != testErr.Error() {
			t.Fatalf("expected error %q, got %q", testErr.Error(), receivedErr.Error())
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected custom throw logic to be called after propagation through nested contexts")
	}
}

package unittest

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// CallMustReturnWithinTimeout is a test helper that invokes the given function and fails the test if the invocation
// does not return prior to the given timeout.
func CallMustReturnWithinTimeout(t *testing.T, f func(), timeout time.Duration, failureMsg string) {
	done := make(chan interface{})

	go func() {
		f()

		close(done)
	}()

	ChannelMustCloseWithinTimeout(t, done, timeout, fmt.Sprintf("function did not return on time: %s", failureMsg))
}

// ChannelMustCloseWithinTimeout is a test helper that fails the test if the channel does not close prior to the given timeout.
func ChannelMustCloseWithinTimeout(t *testing.T, c <-chan interface{}, timeout time.Duration, failureMsg string) {
	select {
	case <-c:
		return
	case <-time.After(timeout):
		require.Fail(t, fmt.Sprintf("channel did not close on time: %s", failureMsg))
	}
}

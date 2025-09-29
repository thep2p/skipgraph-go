package unittest

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
	"sync"
	"testing"
	"time"
)

const DefaultReadyDoneTimeout = 100 * time.Millisecond

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

// ChannelsMustCloseWithinTimeout is a test helper that fails the test if any of the given channels do not close prior to the given timeout.
func ChannelsMustCloseWithinTimeout(t *testing.T, timeout time.Duration, failureMsg string, channels ...<-chan interface{}) {
	wg := sync.WaitGroup{}
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(ch <-chan interface{}) {
			<-ch
			wg.Done()
		}(ch)
	}

	CallMustReturnWithinTimeout(t, wg.Wait, timeout, failureMsg)
}

// RequireAllReady is a test helper that fails the test if any of the given components do not become ready within the default timeout.
func RequireAllReady(t *testing.T, components ...modules.Component) {
	readyChans := make([]<-chan interface{}, len(components))
	for i, c := range components {
		readyChans[i] = c.Ready()
	}
	ChannelsMustCloseWithinTimeout(t, DefaultReadyDoneTimeout, "not all components became ready on time", readyChans...)
}

// RequireAllDone is a test helper that fails the test if any of the given components do not become done within the default timeout.
func RequireAllDone(t *testing.T, components ...modules.Component) {
	doneChans := make([]<-chan interface{}, len(components))
	for i, c := range components {
		doneChans[i] = c.Done()
	}
	ChannelsMustCloseWithinTimeout(t, DefaultReadyDoneTimeout, "not all components became done on time", doneChans...)
}

// ChannelMustNotCloseWithinTimeout is a test helper that fails the test if the channel closes before the given timeout.
// This is useful for testing that a channel remains open when it should not close yet.
func ChannelMustNotCloseWithinTimeout(t *testing.T, c <-chan interface{}, timeout time.Duration, failureMsg string) {
	select {
	case <-c:
		require.Fail(t, fmt.Sprintf("channel closed when it should not have: %s", failureMsg))
	case <-time.After(timeout):
		// Expected behavior - channel did not close within timeout
		return
	}
}

package internal

import (
	"context"
	"github/thep2p/skipgraph-go/model/skipgraph"
)

// Connection represents a connection to a remote peer.
type Connection interface {
	// RemoteAddr returns the remote address of the connection.
	// It returns an empty string if the connection is closed, otherwise it returns the remote address,
	// which is the address of the peer that the connection is connected to.
	RemoteAddr() string

	// Send sends a message to the remote peer, returning an error if the message could not be sent.
	// Send is a blocking operation, and it will block until the message is sent.
	// It returns an error if the message could not be sent.
	// It returns io.EOF if the connection is closed.
	// Errors from this method are expected to be treated as benign, i.e., don't panic on error.
	Send([]byte) error

	// Next returns the next message received from the remote peer.
	// It is a blocking operation, and it will block until a message is received.
	// It returns io.EOF if the connection is closed.
	// It returns an error if the message could not be read.
	// Errors from this method are expected to be treated as benign, i.e., don't panic on error.
	Next() ([]byte, error)

	// Close gracefully closes the connection. Blocking until the connection is closed.
	// Errors from this method can be considered as benign, i.e., don't panic on error.
	Close() error
}

// ConnectionManager establishes and maintains connections.
type ConnectionManager interface {
	// Connect establishes a connection to a remote peer and locally caches and returns the connection.
	// If the connection is already established, it returns the cached connection.
	// The cardinal assumption is there is always at most one connection to a remote peer.
	// Errors from this method are expected to be treated as benign, i.e., don't panic on error.
	Connect(context.Context, skipgraph.Identifier) (Connection, error)

	// Close closes all connections.
	// Errors from this method can be considered as benign, i.e., don't panic on error.
	Close() error
}

package internal

import "github.com/thep2p/skipgraph-go/core/model"

// NodeReference represents a reference to a node in the array using both identifier and array index.
// This is used for testing purposes to enable graph traversal validation.
type NodeReference struct {
	Identifier model.Identifier
	ArrayIndex int
}

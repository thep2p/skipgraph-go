package model

import "github/thep2p/skipgraph-go/model/skipgraph"

type LookupTableUpdater interface {
	Update(dir skipgraph.Direction, level skipgraph.Level, identity skipgraph.Identity) error
}

package node

import (
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
)

type SkipGraphNode struct {
	id model.Identity
	lt core.MutableLookupTable
}

func NewSkipGraphNode(id model.Identity, lt core.MutableLookupTable) *SkipGraphNode {
	return &SkipGraphNode{id: id, lt: lt}
}

func (n *SkipGraphNode) Identifier() model.Identifier {
	return n.id.GetIdentifier()
}

func (n *SkipGraphNode) MembershipVector() model.MembershipVector {
	return n.id.GetMembershipVector()
}

func (n *SkipGraphNode) GetNeighbor(dir core.Direction, level core.Level) (model.Identity, error) {
	return n.lt.GetEntry(dir, level)
}

func (n *SkipGraphNode) SetNeighbor(dir core.Direction, level core.Level, neighbor model.Identity) error {
	return n.lt.AddEntry(dir, level, neighbor)
}

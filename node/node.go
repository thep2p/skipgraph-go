package node

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/core/types"
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

func (n *SkipGraphNode) GetNeighbor(dir types.Direction, level types.Level) (*model.Identity, error) {
	return n.lt.GetEntry(dir, level)
}

func (n *SkipGraphNode) SetNeighbor(dir types.Direction, level types.Level, neighbor model.Identity) error {
	return n.lt.AddEntry(dir, level, neighbor)
}

// SearchByID searches for an identifier in the lookup table in the given direction up to the given level.
//
// Algorithm (corresponds to Algorithm 1 from Skip Graph paper):
// 1. Collects neighbors from levels 0 to req.Level() in req.Direction()
// 2. Filters candidates based on direction:
//   - Left: smallest identifier >= target
//   - Right: greatest identifier <= target
//
// 3. Returns the best match, or falls back to own identifier at level 0 if no match found
//
// Returns error if lookup table access fails at any level.
func (n *SkipGraphNode) SearchByID(req model.IdSearchReq) (model.IdSearchRes, error) {
	// Step 1: Collect candidates from levels 0 to req.Level()
	type candidate struct {
		id    model.Identifier
		level types.Level
	}

	var candidates []candidate

	for level := types.Level(0); level <= req.Level(); level++ {
		identity, err := n.lt.GetEntry(req.Direction(), level)
		if err != nil {
			return model.IdSearchRes{}, fmt.Errorf("error while searching by id in level %d: %w", level, err)
		}
		if identity != nil {
			candidates = append(candidates, candidate{
				id:    identity.GetIdentifier(),
				level: level,
			})
		}
	}

	// Step 2: Filter candidates based on direction
	var bestCandidate *candidate

	target := req.Target()

	switch req.Direction() {
	case types.DirectionLeft:
		// Left: find smallest ID >= target
		for i := range candidates {
			c := &candidates[i]
			cmp := c.id.Compare(&target)
			if cmp.GetComparisonResult() == model.CompareGreater || cmp.GetComparisonResult() == model.CompareEqual {
				if bestCandidate == nil {
					bestCandidate = c
				} else {
					bestCmp := c.id.Compare(&bestCandidate.id)
					if bestCmp.GetComparisonResult() == model.CompareLess {
						bestCandidate = c
					}
				}
			}
		}

	case types.DirectionRight:
		// Right: find greatest ID <= target
		for i := range candidates {
			c := &candidates[i]
			cmp := c.id.Compare(&target)
			if cmp.GetComparisonResult() == model.CompareLess || cmp.GetComparisonResult() == model.CompareEqual {
				if bestCandidate == nil {
					bestCandidate = c
				} else {
					bestCmp := c.id.Compare(&bestCandidate.id)
					if bestCmp.GetComparisonResult() == model.CompareGreater {
						bestCandidate = c
					}
				}
			}
		}
	}

	// Step 3: Return result or fallback
	if bestCandidate != nil {
		return model.NewIdSearchRes(req.Target(), bestCandidate.level, bestCandidate.id), nil
	}

	// Fallback: return own identifier at level 0
	return model.NewIdSearchRes(req.Target(), 0, n.Identifier()), nil
}

package node

import (
	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
	"github.com/pd241008/sentinelmesh/simulator/internal/scorer"
)

type Digest struct {
	NodeID   int
	Score    float64
	Category string
	Round    int // Discrete gossip round at which this digest was generated
}

type Node struct {
	ID          int
	Scorer      *scorer.Scorer
	Flows       []dataset.Flow
	DigestCache map[int]Digest // PeerNodeID -> latest Digest
	CurrentFlow int
}

func New(id int, flows []dataset.Flow, alpha float64) *Node {
	return &Node{
		ID:          id,
		Scorer:      scorer.New(alpha),
		Flows:       flows,
		DigestCache: make(map[int]Digest),
	}
}

func (n *Node) ProcessFlowsForRound(round int) (*Digest, bool) {
	if n.CurrentFlow >= len(n.Flows) {
		return nil, false
	}

	flow := n.Flows[n.CurrentFlow]
	n.CurrentFlow++

	score := n.Scorer.ScoreFlow(flow)

	digest := Digest{
		NodeID:   n.ID,
		Score:    score,
		Category: flow.Category,
		Round:    round,
	}

	n.DigestCache[n.ID] = digest

	return &digest, true
}

func (n *Node) ReceiveDigest(d Digest) {
	if existing, ok := n.DigestCache[d.NodeID]; !ok || d.Round > existing.Round {
		n.DigestCache[d.NodeID] = d
	}
}

func (n *Node) GetCache() map[int]Digest {
	return n.DigestCache
}

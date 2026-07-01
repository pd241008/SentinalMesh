package gossip

import (
	"testing"

	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
)

func TestNew(t *testing.T) {
	nodes := []*node.Node{node.New(0, nil, 0.3)}
	g := New(nodes, 2, 5, 42)
	if g == nil {
		t.Fatal("New returned nil")
	}
	if g.fanout != 2 || g.window != 5 {
		t.Fatalf("bad params: %+v", g)
	}
}

func makeTestNodes(t *testing.T, n int, flowsPerNode int) []*node.Node {
	t.Helper()
	nodes := make([]*node.Node, n)
	for i := 0; i < n; i++ {
		flows := make([]dataset.Flow, flowsPerNode)
		for j := 0; j < flowsPerNode; j++ {
			flows[j] = dataset.Flow{
				ID:       i*flowsPerNode + j + 1,
				Category: "normal",
				Sbytes:   100,
				Dbytes:   200,
				Spkts:    2,
				Dpkts:    4,
				Rate:     10.0,
			}
		}
		nodes[i] = node.New(i, flows, 0.3)
	}
	return nodes
}

func TestSelectPeers(t *testing.T) {
	nodes := makeTestNodes(t, 8, 5)
	g := New(nodes, 3, 5, 42)

	peers := g.selectPeers(0)
	if len(peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(peers))
	}
	for _, p := range peers {
		if p.ID == 0 {
			t.Fatal("should not select self as peer")
		}
	}

	seen := make(map[int]bool)
	for _, p := range peers {
		if seen[p.ID] {
			t.Fatal("duplicate peer selected")
		}
		seen[p.ID] = true
	}
}

func TestSelectPeersSingleNode(t *testing.T) {
	nodes := makeTestNodes(t, 1, 5)
	g := New(nodes, 2, 5, 42)
	peers := g.selectPeers(0)
	if len(peers) != 0 {
		t.Fatalf("expected 0 peers for single node, got %d", len(peers))
	}
}

func TestRound(t *testing.T) {
	nodes := makeTestNodes(t, 4, 10)
	g := New(nodes, 2, 5, 42)

	g.Round(1)

	for _, n := range nodes {
		cache := n.GetCache()
		if len(cache) == 0 {
			t.Fatalf("node %d should have cache entries after round", n.ID)
		}
	}
}

func TestEvictStale(t *testing.T) {
	nodes := makeTestNodes(t, 2, 10)
	g := New(nodes, 2, 2, 42)

	g.Round(1)
	if len(nodes[0].GetCache()) == 0 {
		t.Fatal("expected cache entries after round 1")
	}

	g.Round(4)
	if len(nodes[1].GetCache()) == 0 {
		t.Fatal("expected some cache entries after round 4")
	}
}

func TestDeterministicPeers(t *testing.T) {
	nodes1 := makeTestNodes(t, 16, 5)
	nodes2 := makeTestNodes(t, 16, 5)

	g1 := New(nodes1, 3, 5, 99)
	g2 := New(nodes2, 3, 5, 99)

	p1 := g1.selectPeers(0)
	p2 := g2.selectPeers(0)

	if len(p1) != len(p2) {
		t.Fatalf("different peer count: %d vs %d", len(p1), len(p2))
	}
	for i := range p1 {
		if p1[i].ID != p2[i].ID {
			t.Fatalf("different peer at index %d: %d vs %d", i, p1[i].ID, p2[i].ID)
		}
	}
}

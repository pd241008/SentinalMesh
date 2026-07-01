package baseline

import (
	"testing"

	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
)

func makeBaselineNodes(t *testing.T, n int, flowsPerNode int, alpha float64) []*node.Node {
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
		nodes[i] = node.New(i, flows, alpha)
	}
	return nodes
}

func TestRunIndependent(t *testing.T) {
	nodes := makeBaselineNodes(t, 4, 10, 0.3)
	alerts := RunIndependent(nodes, 0.5, 2, 5, 5)
	if alerts == nil {
		t.Fatal("expected non-nil alerts")
	}
}

func TestRunIndependentNoAlerts(t *testing.T) {
	nodes := makeBaselineNodes(t, 4, 5, 0.3)
	alerts := RunIndependent(nodes, 0.99, 2, 5, 5)
	for round, cats := range alerts {
		t.Errorf("unexpected alert at round %d: %v", round, cats)
	}
}

func TestRunCentralized(t *testing.T) {
	nodes := makeBaselineNodes(t, 4, 10, 0.3)
	alerts := RunCentralized(nodes, 0.5, 2, 5, 5)
	if alerts == nil {
		t.Fatal("expected non-nil alerts")
	}
}

func TestRunCentralizedNoAlerts(t *testing.T) {
	nodes := makeBaselineNodes(t, 4, 5, 0.3)
	alerts := RunCentralized(nodes, 0.99, 2, 5, 5)
	for round, cats := range alerts {
		t.Errorf("unexpected alert at round %d: %v", round, cats)
	}
}

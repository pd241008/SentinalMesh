package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pd241008/sentinelmesh/simulator/internal/baseline"
	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
	"github.com/pd241008/sentinelmesh/simulator/internal/fragment"
	"github.com/pd241008/sentinelmesh/simulator/internal/gossip"
	"github.com/pd241008/sentinelmesh/simulator/internal/metrics"
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
	"github.com/pd241008/sentinelmesh/simulator/internal/quorum"
	"github.com/pd241008/sentinelmesh/simulator/internal/sweep"
)

var testdataPath = filepath.Join("..", "testdata", "testdata.csv")

func TestFullPipeline(t *testing.T) {
	allFlows, err := dataset.LoadCSV(testdataPath)
	if err != nil {
		t.Fatalf("LoadCSV: %v", err)
	}
	if len(allFlows) != 15 {
		t.Fatalf("expected 15 flows, got %d", len(allFlows))
	}

	attackCategories := []string{"fuzzers", "reconnaissance", "dos", "exploits", "generic", "analysis", "backdoor"}
	partitions := fragment.DistributeFlows(allFlows, 4, 2, attackCategories)
	if len(partitions) != 4 {
		t.Fatalf("expected 4 partitions, got %d", len(partitions))
	}

	totalRounds := 0
	for _, p := range partitions {
		if len(p) > totalRounds {
			totalRounds = len(p)
		}
	}
	t.Logf("totalRounds: %d", totalRounds)

	nodes := make([]*node.Node, 4)
	for i := 0; i < 4; i++ {
		nodes[i] = node.New(i, partitions[i], 0.3)
	}

	g := gossip.New(nodes, 2, 5, 42)
	alerts := make(baseline.AlertTimeline)

	roundAlerts := 0
	for round := 1; round <= totalRounds; round++ {
		g.Round(round)
		for _, n := range nodes {
			cats := quorum.Evaluate(n.GetCache(), 0.3, 2, 5, round)
			for _, cat := range cats {
				if alerts[round] == nil {
					alerts[round] = make(map[string]bool)
				}
				alerts[round][cat] = true
				roundAlerts++
			}
		}
	}

	t.Logf("Total round alerts: %d", roundAlerts)

	result := metrics.Compute(allFlows, alerts, totalRounds*4*2)
	t.Logf("Recall: %.4f, Bandwidth: %d, Latency: %.2f", result.Recall, result.Bandwidth, result.AvgLatency)
}

func TestSweepIntegration(t *testing.T) {
	allFlows, err := dataset.LoadCSV(testdataPath)
	if err != nil {
		t.Fatalf("LoadCSV: %v", err)
	}

	cfg := &sweep.Config{}
	cfg.Sweep.N = []int{4}
	cfg.Sweep.F = []int{2}
	cfg.Sweep.Q = []int{2}
	cfg.Sweep.K = []int{2}
	cfg.Sweep.W = 3

	outputDir := t.TempDir()
	err = sweep.Run(cfg, allFlows, 0.3, 0.3, 42, outputDir)
	if err != nil {
		t.Fatalf("sweep.Run: %v", err)
	}

	csvPath := filepath.Join(outputDir, "sweep_results.csv")
	data, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("CSV is empty")
	}
	t.Logf("CSV output:\n%s", string(data))
}

func TestEndToEndWithBaselines(t *testing.T) {
	allFlows, err := dataset.LoadCSV(testdataPath)
	if err != nil {
		t.Fatalf("LoadCSV: %v", err)
	}

	attackCategories := []string{"fuzzers", "reconnaissance", "dos", "exploits", "generic", "analysis", "backdoor"}
	partitions := fragment.DistributeFlows(allFlows, 4, 2, attackCategories)

	totalRounds := 0
	for _, p := range partitions {
		if len(p) > totalRounds {
			totalRounds = len(p)
		}
	}

	makeNodes := func() []*node.Node {
		nodes := make([]*node.Node, 4)
		for i := 0; i < 4; i++ {
			nodes[i] = node.New(i, partitions[i], 0.3)
		}
		return nodes
	}

	gossipAlerts := runGossipSim(t, makeNodes(), totalRounds)
	indepAlerts := baseline.RunIndependent(makeNodes(), 0.3, 2, 5, totalRounds)
	centAlerts := baseline.RunCentralized(makeNodes(), 0.3, 2, 5, totalRounds)

	gResult := metrics.Compute(allFlows, gossipAlerts, totalRounds*4*2)
	iResult := metrics.Compute(allFlows, indepAlerts, 0)
	cResult := metrics.Compute(allFlows, centAlerts, 0)

	t.Logf("Gossip:      Recall=%.4f  BW=%d  Lat=%.2f", gResult.Recall, gResult.Bandwidth, gResult.AvgLatency)
	t.Logf("Independent: Recall=%.4f  BW=%d  Lat=%.2f", iResult.Recall, iResult.Bandwidth, iResult.AvgLatency)
	t.Logf("Centralized: Recall=%.4f  BW=%d  Lat=%.2f", cResult.Recall, cResult.Bandwidth, cResult.AvgLatency)
}

func runGossipSim(t *testing.T, nodes []*node.Node, totalRounds int) baseline.AlertTimeline {
	t.Helper()
	g := gossip.New(nodes, 2, 5, 42)
	alerts := make(baseline.AlertTimeline)
	for round := 1; round <= totalRounds; round++ {
		g.Round(round)
		for _, n := range nodes {
			for _, cat := range quorum.Evaluate(n.GetCache(), 0.3, 2, 5, round) {
				if alerts[round] == nil {
					alerts[round] = make(map[string]bool)
				}
				alerts[round][cat] = true
			}
		}
	}
	return alerts
}

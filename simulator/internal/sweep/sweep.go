package sweep

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/pd241008/sentinelmesh/simulator/internal/baseline"
	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
	"github.com/pd241008/sentinelmesh/simulator/internal/fragment"
	"github.com/pd241008/sentinelmesh/simulator/internal/gossip"
	"github.com/pd241008/sentinelmesh/simulator/internal/metrics"
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
	"github.com/pd241008/sentinelmesh/simulator/internal/quorum"
	"gopkg.in/yaml.v3"
)

var AttackCategories = []string{
	"analysis", "backdoor", "dos", "exploits", "fuzzers",
	"generic", "reconnaissance", "shellcode", "worms",
}

type Config struct {
	Sweep struct {
		N []int `yaml:"N"`
		F []int `yaml:"f"`
		Q []int `yaml:"q"`
		K []int `yaml:"k"`
		W int   `yaml:"W"`
	} `yaml:"sweep"`
}

type RunResult struct {
	N              int
	F              int
	Q              int
	K              int
	W              int
	GossipRecall   float64
	GossipBandwidth int
	GossipLatency  float64
	IndepRecall    float64
	IndepBandwidth int
	IndepLatency   float64
	CentRecall     float64
	CentBandwidth  int
	CentLatency    float64
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Run(cfg *Config, allFlows []dataset.Flow, alpha float64, threshold float64, seed int64, outputDir string) error {
	var results []RunResult
	idx := int64(0)

	for _, N := range cfg.Sweep.N {
		for _, k := range cfg.Sweep.K {
			partitions := fragment.DistributeFlows(allFlows, N, k, AttackCategories)
			totalRounds := 0
			for _, p := range partitions {
				if len(p) > totalRounds {
					totalRounds = len(p)
				}
			}

			for _, f := range cfg.Sweep.F {
				for _, q := range cfg.Sweep.Q {
					idx++
					result := RunResult{
						N: N, F: f, Q: q, K: k, W: cfg.Sweep.W,
					}

					gossipNodes := makeNodes(N, partitions, alpha)
					gResult := runGossip(gossipNodes, f, cfg.Sweep.W, threshold, q, totalRounds, seed+idx)
					gMetrics := metrics.Compute(allFlows, gResult.alerts, gResult.totalDigests)
					result.GossipRecall = gMetrics.Recall
					result.GossipBandwidth = gMetrics.Bandwidth
					result.GossipLatency = gMetrics.AvgLatency

					indepNodes := makeNodes(N, partitions, alpha)
					iAlerts := baseline.RunIndependent(indepNodes, threshold, q, cfg.Sweep.W, totalRounds)
					iMetrics := metrics.Compute(allFlows, iAlerts, 0)
					result.IndepRecall = iMetrics.Recall
					result.IndepBandwidth = 0
					result.IndepLatency = iMetrics.AvgLatency

					centNodes := makeNodes(N, partitions, alpha)
					cAlerts := baseline.RunCentralized(centNodes, threshold, q, cfg.Sweep.W, totalRounds)
					cMetrics := metrics.Compute(allFlows, cAlerts, 0)
					result.CentRecall = cMetrics.Recall
					result.CentBandwidth = 0
					result.CentLatency = cMetrics.AvgLatency

					results = append(results, result)
				}
			}
		}
	}

	return writeCSV(results, outputDir)
}

type gossipResult struct {
	alerts        baseline.AlertTimeline
	totalDigests  int
}

func runGossip(nodes []*node.Node, fanout int, window int, threshold float64, quorumThreshold int, totalRounds int, seed int64) gossipResult {
	g := gossip.New(nodes, fanout, window, seed)
	alerts := make(baseline.AlertTimeline)
	totalDigests := 0

	for round := 1; round <= totalRounds; round++ {
		g.Round(round)
		totalDigests += fanout * len(nodes)

		for _, n := range nodes {
			for _, cat := range quorum.Evaluate(n.GetCache(), threshold, quorumThreshold, window, round) {
				if alerts[round] == nil {
					alerts[round] = make(map[string]bool)
				}
				alerts[round][cat] = true
			}
		}
	}

	return gossipResult{alerts: alerts, totalDigests: totalDigests}
}

func makeNodes(N int, partitions [][]dataset.Flow, alpha float64) []*node.Node {
	nodes := make([]*node.Node, N)
	for i := 0; i < N; i++ {
		nodes[i] = node.New(i, partitions[i], alpha)
	}
	return nodes
}

func writeCSV(results []RunResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(outputDir, "sweep_results.csv")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"N", "f", "q", "k", "W",
		"gossip_recall", "gossip_bandwidth", "gossip_latency",
		"indep_recall", "indep_bandwidth", "indep_latency",
		"cent_recall", "cent_bandwidth", "cent_latency",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{
			intStr(r.N), intStr(r.F), intStr(r.Q), intStr(r.K), intStr(r.W),
			floatStr(r.GossipRecall), intStr(r.GossipBandwidth), floatStr(r.GossipLatency),
			floatStr(r.IndepRecall), intStr(r.IndepBandwidth), floatStr(r.IndepLatency),
			floatStr(r.CentRecall), intStr(r.CentBandwidth), floatStr(r.CentLatency),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	fmt.Printf("Wrote %d results to %s\n", len(results), path)
	return nil
}

func intStr(i int) string {
	return fmt.Sprintf("%d", i)
}

func floatStr(f float64) string {
	if math.IsNaN(f) {
		return "0"
	}
	return fmt.Sprintf("%.4f", f)
}

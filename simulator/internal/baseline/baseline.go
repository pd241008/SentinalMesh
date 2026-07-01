package baseline

import (
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
	"github.com/pd241008/sentinelmesh/simulator/internal/quorum"
)

type AlertTimeline map[int]map[string]bool

func addAlert(alerts AlertTimeline, round int, category string) {
	if alerts[round] == nil {
		alerts[round] = make(map[string]bool)
	}
	alerts[round][category] = true
}

func RunIndependent(nodes []*node.Node, threshold float64, quorumThreshold int, window int, totalRounds int) AlertTimeline {
	alerts := make(AlertTimeline)
	for round := 1; round <= totalRounds; round++ {
		for _, n := range nodes {
			d, ok := n.ProcessFlowsForRound(round)
			if !ok {
				continue
			}
			if d.Score > threshold {
				n.DigestCache[n.ID] = *d
			}
		}
		for _, n := range nodes {
			for _, cat := range quorum.Evaluate(n.GetCache(), threshold, quorumThreshold, window, round) {
				addAlert(alerts, round, cat)
			}
		}
	}
	return alerts
}

func RunCentralized(nodes []*node.Node, threshold float64, quorumThreshold int, window int, totalRounds int) AlertTimeline {
	centralCache := make(map[int]node.Digest)
	alerts := make(AlertTimeline)

	for round := 1; round <= totalRounds; round++ {
		for _, n := range nodes {
			d, ok := n.ProcessFlowsForRound(round)
			if !ok {
				continue
			}
			if d.Score > threshold {
				centralCache[n.ID] = *d
			}
		}
		for _, cat := range quorum.Evaluate(centralCache, threshold, quorumThreshold, window, round) {
			addAlert(alerts, round, cat)
		}
		for id := range centralCache {
			if round-centralCache[id].Round > window {
				delete(centralCache, id)
			}
		}
	}
	return alerts
}

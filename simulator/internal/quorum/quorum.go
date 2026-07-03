package quorum

import (
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
)

func Evaluate(cache map[int][]node.Digest, quorum int, window int, currentRound int) map[string][]int {
	categoryCount := make(map[string]int)

	for _, digests := range cache {
		// Dedup node's votes: a node gets maximum 1 vote per category within the window W.
		nodeCats := make(map[string]bool)
		for _, d := range digests {
			if currentRound-d.Round <= window {
				nodeCats[d.Category] = true
			}
		}
		
		// Increment global count only once per category per source node
		for cat := range nodeCats {
			categoryCount[cat]++
		}
	}

	alerts := make(map[string][]int)
	for cat, count := range categoryCount {
		if count >= quorum {
			var nodeIDs []int
			for srcID, digests := range cache {
				for _, d := range digests {
					if d.Category == cat && currentRound-d.Round <= window {
						nodeIDs = append(nodeIDs, srcID)
						break
					}
				}
			}
			alerts[cat] = nodeIDs
		}
	}

	return alerts
}

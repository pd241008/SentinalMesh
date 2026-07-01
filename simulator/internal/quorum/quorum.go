package quorum

import (
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
)

func Evaluate(cache map[int]node.Digest, threshold float64, quorum int, window int, currentRound int) []string {
	categoryCount := make(map[string]int)

	for _, d := range cache {
		if d.Score > threshold && currentRound-d.Round <= window {
			categoryCount[d.Category]++
		}
	}

	var alerts []string
	for cat, count := range categoryCount {
		if count >= quorum {
			alerts = append(alerts, cat)
		}
	}

	return alerts
}

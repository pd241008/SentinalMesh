package metrics

import (
	"math"

	"github.com/pd241008/sentinelmesh/simulator/internal/baseline"
	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
)

type Result struct {
	Recall     float64
	Bandwidth  int
	AvgLatency float64
}

func Compute(allFlows []dataset.Flow, alerts baseline.AlertTimeline, totalDigests int) Result {
	alertedCategories := make(map[string]bool)
	firstAlertRound := make(map[string]int)

	maxRound := 0
	for r := range alerts {
		if r > maxRound {
			maxRound = r
		}
	}
	for round := 1; round <= maxRound; round++ {
		if cats, ok := alerts[round]; ok {
			for cat := range cats {
				if !alertedCategories[cat] {
					alertedCategories[cat] = true
					firstAlertRound[cat] = round
				}
			}
		}
	}

	totalAttackFlows := 0
	detectedAttackFlows := 0

	for _, flow := range allFlows {
		if !flow.IsAttack {
			continue
		}
		totalAttackFlows++
		if alertedCategories[flow.Category] {
			detectedAttackFlows++
		}
	}

	var recall float64
	if totalAttackFlows > 0 {
		recall = float64(detectedAttackFlows) / float64(totalAttackFlows)
	}

	var avgLatency float64
	if len(firstAlertRound) > 0 {
		avgLatency = float64(maxRound) / float64(len(firstAlertRound))
	}

	return Result{
		Recall:     math.Round(recall*10000) / 10000,
		Bandwidth:  totalDigests,
		AvgLatency: math.Round(avgLatency*100) / 100,
	}
}

package metrics

import (
	"fmt"
	"math"
	"sort"

	"github.com/pd241008/sentinelmesh/simulator/internal/baseline"
	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
	"github.com/pd241008/sentinelmesh/simulator/internal/fragment"
	"github.com/pd241008/sentinelmesh/simulator/internal/node"
)

type Result struct {
	FlowReconRecall   float64
	FlowDoSRecall     float64
	CorrectedFlowReconRecall float64
	CorrectedFlowDoSRecall   float64
	WindowReconRecall float64
	WindowDoSRecall   float64
	ReconFPR          float64
	DoSFPR            float64
	BandwidthKBps     float64
	AvgLatency        float64
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	sortedA := append([]int{}, a...)
	sortedB := append([]int{}, b...)
	sort.Ints(sortedA)
	sort.Ints(sortedB)
	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}
	return true
}

func Compute(nodes []*node.Node, allFlows []dataset.Flow, tAlerts baseline.AlertTimeline, cAlerts baseline.AlertTimeline, totalDigests int, window int, campaigns []fragment.Campaign, totalRounds int, numNodes int) Result {
	maxRound := totalRounds

	// Bandwidth
	digestSizeBytes := 32.0
	roundDurationSeconds := 1.0
	totalBytes := float64(totalDigests) * digestSizeBytes
	totalSeconds := float64(totalRounds) * roundDurationSeconds
	var bandwidthKBps float64
	if totalSeconds > 0 {
		bandwidthKBps = (totalBytes / float64(numNodes)) / totalSeconds / 1024.0
	}

	firstAlertRound := make(map[string]int)
	for r := 1; r <= maxRound; r++ {
		if cats, ok := tAlerts[r]; ok {
			for cat := range cats {
				if _, exists := firstAlertRound[cat]; !exists {
					firstAlertRound[cat] = r
				}
			}
		}
	}

	// Flow-level recall [r, r+W]
	totalReconFlows := 0
	detectedReconFlows := 0
	correctedDetectedReconFlows := 0
	totalDosFlows := 0
	detectedDosFlows := 0
	correctedDetectedDosFlows := 0

	var totalExactMatches int
	var totalNoiseCoincidences int

	for r := 1; r <= maxRound; r++ {
		for _, n := range nodes {
			if r-1 < len(n.Flows) {
				flow := n.Flows[r-1]
				if flow.Category == "reconnaissance" || flow.Category == "dos" {
					detected := false
					var escRound int

					for rr := r; rr <= min(r+window, maxRound); rr++ {
						if tAlerts[rr] != nil && len(tAlerts[rr][flow.Category]) > 0 {
							detected = true
							escRound = rr
							break
						}
					}
					
					if flow.Category == "reconnaissance" {
						totalReconFlows++
						if detected {
							detectedReconFlows++
							// Corrected Check
							if cAlerts != nil && cAlerts[escRound] != nil && len(cAlerts[escRound][flow.Category]) > 0 {
								tCorrobs := tAlerts[escRound][flow.Category]
								cCorrobs := cAlerts[escRound][flow.Category]
								if equalSlices(tCorrobs, cCorrobs) {
									totalExactMatches++
								} else {
									totalNoiseCoincidences++
								}
							} else {
								correctedDetectedReconFlows++
							}
						}
					} else if flow.Category == "dos" {
						totalDosFlows++
						if detected {
							detectedDosFlows++
							// Corrected Check
							if cAlerts != nil && cAlerts[escRound] != nil && len(cAlerts[escRound][flow.Category]) > 0 {
								tCorrobs := tAlerts[escRound][flow.Category]
								cCorrobs := cAlerts[escRound][flow.Category]
								if equalSlices(tCorrobs, cCorrobs) {
									totalExactMatches++
								} else {
									totalNoiseCoincidences++
								}
							} else {
								correctedDetectedDosFlows++
							}
						}
					}
				}
			}
		}
	}

	if (detectedReconFlows > 0 || detectedDosFlows > 0) && cAlerts != nil {
		fmt.Printf("DEBUG: DetectedRecon=%d/%d, CorrectedRecon=%d/%d, CorrectedDoS=%d/%d, ExactMatches=%d, NoiseCoincidences=%d\n", 
			detectedReconFlows, totalReconFlows, correctedDetectedReconFlows, totalReconFlows, correctedDetectedDosFlows, totalDosFlows, totalExactMatches, totalNoiseCoincidences)
	}

	var flowReconRecall float64
	var correctedFlowReconRecall float64
	if totalReconFlows > 0 {
		flowReconRecall = float64(detectedReconFlows) / float64(totalReconFlows)
		correctedFlowReconRecall = float64(correctedDetectedReconFlows) / float64(totalReconFlows)
	}
	
	var flowDosRecall float64
	var correctedFlowDosRecall float64
	if totalDosFlows > 0 {
		flowDosRecall = float64(detectedDosFlows) / float64(totalDosFlows)
		correctedFlowDosRecall = float64(correctedDetectedDosFlows) / float64(totalDosFlows)
	}

	// Window-level Recall & FPR
	totalReconActiveWindows := 0
	truePosReconWindows := 0
	totalReconNormalWindows := 0
	falsePosReconWindows := 0

	totalDoSActiveWindows := 0
	truePosDoSWindows := 0
	totalDoSNormalWindows := 0
	falsePosDoSWindows := 0

	for r := 1; r <= maxRound; r++ {
		hasReconFlow := false
		hasDoSFlow := false
		for _, n := range nodes {
			if r-1 < len(n.Flows) {
				cat := n.Flows[r-1].Category
				if cat == "reconnaissance" {
					hasReconFlow = true
				}
				if cat == "dos" {
					hasDoSFlow = true
				}
			}
		}
		
		if hasReconFlow {
			totalReconActiveWindows++
			if tAlerts[r] != nil && len(tAlerts[r]["reconnaissance"]) > 0 {
				truePosReconWindows++
			}
		} else {
			totalReconNormalWindows++
			if tAlerts[r] != nil && len(tAlerts[r]["reconnaissance"]) > 0 {
				falsePosReconWindows++
			}
		}
		
		if hasDoSFlow {
			totalDoSActiveWindows++
			if tAlerts[r] != nil && len(tAlerts[r]["dos"]) > 0 {
				truePosDoSWindows++
			}
		} else {
			totalDoSNormalWindows++
			if tAlerts[r] != nil && len(tAlerts[r]["dos"]) > 0 {
				falsePosDoSWindows++
			}
		}
	}
	
	var windowReconRecall float64
	if totalReconActiveWindows > 0 {
		windowReconRecall = float64(truePosReconWindows) / float64(totalReconActiveWindows)
	}
	var windowDosRecall float64
	if totalDoSActiveWindows > 0 {
		windowDosRecall = float64(truePosDoSWindows) / float64(totalDoSActiveWindows)
	}

	var reconFPR float64
	if totalReconNormalWindows > 0 {
		reconFPR = float64(falsePosReconWindows) / float64(totalReconNormalWindows)
	}
	var dosFPR float64
	if totalDoSNormalWindows > 0 {
		dosFPR = float64(falsePosDoSWindows) / float64(totalDoSNormalWindows)
	}

	var avgLatency float64
	if len(firstAlertRound) > 0 {
		avgLatency = float64(maxRound) / float64(len(firstAlertRound))
	}

	return Result{
		FlowReconRecall:   math.Round(flowReconRecall*10000) / 10000,
		FlowDoSRecall:     math.Round(flowDosRecall*10000) / 10000,
		CorrectedFlowReconRecall: math.Round(correctedFlowReconRecall*10000) / 10000,
		CorrectedFlowDoSRecall:   math.Round(correctedFlowDosRecall*10000) / 10000,
		WindowReconRecall: math.Round(windowReconRecall*10000) / 10000,
		WindowDoSRecall:   math.Round(windowDosRecall*10000) / 10000,
		ReconFPR:          math.Round(reconFPR*10000) / 10000,
		DoSFPR:            math.Round(dosFPR*10000) / 10000,
		BandwidthKBps:     bandwidthKBps,
		AvgLatency:        math.Round(avgLatency*100) / 100,
	}
}

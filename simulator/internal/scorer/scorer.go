package scorer

import (
	"math"

	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
)

type FeatureEWMA struct {
	alpha    float64
	mean     float64
	variance float64
	count    int
}

func (f *FeatureEWMA) update(val float64) float64 {
	if f.count == 0 {
		f.mean = val
		f.variance = 0
		f.count++
		return 0
	}

	stddev := math.Sqrt(f.variance)
	var z float64
	if stddev > 0 {
		z = (val - f.mean) / stddev
	} else if val != f.mean {
		z = 3.0 // Arbitrary significant deviation if variance was 0
	}

	diff := val - f.mean
	
	// Fix: EWMA Contamination. Only update baseline if the observation looks "normal"
	// Changed threshold from 2.5 to 15.0 after applying log1p scaling (covers ~99% of normal traffic)
	// Must also update if variance is low, otherwise it gets permanently stuck at z=huge
	if math.Abs(z) < 15.0 || f.variance < 0.1 {
		f.mean += f.alpha * diff
		f.variance = (1-f.alpha)*(f.variance + f.alpha*diff*diff)
	}
	f.count++

	return math.Abs(z)
}

type Thresholds struct {
	Recon float64
	DoS   float64
	Ewma  float64
}

type Scorer struct {
	alpha      float64
	thresholds Thresholds
	features   map[string]*FeatureEWMA
}

func New(alpha float64, thresholds Thresholds) *Scorer {
	return &Scorer{
		alpha:      alpha,
		thresholds: thresholds,
		features: map[string]*FeatureEWMA{
			"Sbytes": {alpha: alpha},
			"Dbytes": {alpha: alpha},
			"Spkts":  {alpha: alpha},
			"Dpkts":  {alpha: alpha},
			"Rate":   {alpha: alpha},
		},
	}
}

func (s *Scorer) ScoreFlow(flow dataset.Flow) (float64, string, bool) {
	// 1. EWMA Z-Score (Generic)
	zSbytes := s.features["Sbytes"].update(math.Log1p(float64(flow.Sbytes)))
	zDbytes := s.features["Dbytes"].update(math.Log1p(float64(flow.Dbytes)))
	zSpkts := s.features["Spkts"].update(math.Log1p(float64(flow.Spkts)))
	zDpkts := s.features["Dpkts"].update(math.Log1p(float64(flow.Dpkts)))
	zRate := s.features["Rate"].update(math.Log1p(flow.Rate))

	ewmaScore := zSbytes
	if zDbytes > ewmaScore { ewmaScore = zDbytes }
	if zSpkts > ewmaScore { ewmaScore = zSpkts }
	if zDpkts > ewmaScore { ewmaScore = zDpkts }
	if zRate > ewmaScore { ewmaScore = zRate }
	ewmaScore = ewmaScore / 5.0 // Normalize a bit to roughly [0, 3+]

	// 2. Absolute Magnitude (DoS)
	dosScore := math.Log1p(flow.Rate)/14.0
	if v := math.Log1p(float64(flow.Sbytes))/21.0; v > dosScore { dosScore = v }
	if v := math.Log1p(float64(flow.Dbytes))/21.0; v > dosScore { dosScore = v }
	if v := math.Log1p(float64(flow.Spkts))/14.0; v > dosScore { dosScore = v }
	if v := math.Log1p(float64(flow.Dpkts))/14.0; v > dosScore { dosScore = v }

	// 3. Inverse Fan-Out (Recon)
	fanoutDst := float64(flow.CtDstLtm) / (float64(flow.CtDstSportLtm) + 1.0)
	reconScore := 1.0 / (fanoutDst + 0.1)

	// Threshold gating with priority-based tie-breaker.
	// DoS and Recon are the paper's measured categories and take priority.
	// EWMA/generic is a fallback for residual anomalies — it should never
	// steal a classification from a specific detector that also fired.
	dosMargin := dosScore / s.thresholds.DoS
	reconMargin := reconScore / s.thresholds.Recon

	maxMargin := 1.0 // strictly greater than 1.0 means it breached
	isAnomalous := false
	guessedCategory := "normal"
	finalScore := ewmaScore

	// Priority 1: DoS and Recon (paper's measured categories)
	if dosMargin > maxMargin {
		maxMargin = dosMargin
		guessedCategory = "dos"
		isAnomalous = true
		finalScore = dosScore
	}
	if reconMargin > maxMargin {
		maxMargin = reconMargin
		guessedCategory = "reconnaissance"
		isAnomalous = true
		finalScore = reconScore
	}

	// Priority 2: EWMA/generic — only if no specific detector fired
	if !isAnomalous {
		ewmaMargin := ewmaScore / s.thresholds.Ewma
		if ewmaMargin > 1.0 {
			guessedCategory = "generic"
			isAnomalous = true
			finalScore = ewmaScore
		}
	}

	return finalScore, guessedCategory, isAnomalous
}


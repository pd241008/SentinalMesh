package quorum

import (
	"testing"

	"github.com/pd241008/sentinelmesh/simulator/internal/node"
)

func TestEvaluateBelowThreshold(t *testing.T) {
	cache := map[int]node.Digest{
		1: {NodeID: 1, Score: 0.3, Category: "dos", Round: 1},
		2: {NodeID: 2, Score: 0.4, Category: "dos", Round: 1},
	}
	alerts := Evaluate(cache, 0.5, 2, 5, 2)
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %v", alerts)
	}
}

func TestEvaluateQuorumMet(t *testing.T) {
	cache := map[int]node.Digest{
		1: {NodeID: 1, Score: 0.7, Category: "dos", Round: 1},
		2: {NodeID: 2, Score: 0.8, Category: "dos", Round: 1},
		3: {NodeID: 3, Score: 0.6, Category: "dos", Round: 1},
	}
	alerts := Evaluate(cache, 0.5, 3, 5, 2)
	if len(alerts) != 1 || alerts[0] != "dos" {
		t.Fatalf("expected [dos], got %v", alerts)
	}
}

func TestEvaluateMultipleCategories(t *testing.T) {
	cache := map[int]node.Digest{
		1: {NodeID: 1, Score: 0.8, Category: "dos", Round: 1},
		2: {NodeID: 2, Score: 0.9, Category: "dos", Round: 1},
		3: {NodeID: 3, Score: 0.8, Category: "fuzzers", Round: 1},
		4: {NodeID: 4, Score: 0.9, Category: "fuzzers", Round: 1},
	}
	alerts := Evaluate(cache, 0.5, 2, 5, 2)
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %v", alerts)
	}
}

func TestEvaluateWindowExcluded(t *testing.T) {
	cache := map[int]node.Digest{
		1: {NodeID: 1, Score: 0.9, Category: "dos", Round: 1},
		2: {NodeID: 2, Score: 0.9, Category: "dos", Round: 1},
	}
	alerts := Evaluate(cache, 0.5, 2, 1, 10)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts (outside window), got %v", alerts)
	}
}

func TestEvaluateEmptyCache(t *testing.T) {
	alerts := Evaluate(map[int]node.Digest{}, 0.5, 2, 5, 1)
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts for empty cache, got %v", alerts)
	}
}

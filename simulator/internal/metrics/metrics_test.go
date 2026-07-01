package metrics

import (
	"testing"

	"github.com/pd241008/sentinelmesh/simulator/internal/baseline"
	"github.com/pd241008/sentinelmesh/simulator/internal/dataset"
)

func TestComputeFullRecall(t *testing.T) {
	flows := []dataset.Flow{
		{ID: 1, Category: "dos", IsAttack: true},
		{ID: 2, Category: "dos", IsAttack: true},
	}
	alerts := baseline.AlertTimeline{
		1: {"dos": true},
	}
	result := Compute(flows, alerts, 100)
	if result.Recall != 1.0 {
		t.Fatalf("expected recall 1.0, got %f", result.Recall)
	}
	if result.Bandwidth != 100 {
		t.Fatalf("expected bandwidth 100, got %d", result.Bandwidth)
	}
}

func TestComputeZeroRecall(t *testing.T) {
	flows := []dataset.Flow{
		{ID: 1, Category: "dos", IsAttack: true},
		{ID: 2, Category: "fuzzers", IsAttack: true},
	}
	alerts := baseline.AlertTimeline{}
	result := Compute(flows, alerts, 0)
	if result.Recall != 0.0 {
		t.Fatalf("expected recall 0.0, got %f", result.Recall)
	}
}

func TestComputePartialRecall(t *testing.T) {
	flows := []dataset.Flow{
		{ID: 1, Category: "dos", IsAttack: true},
		{ID: 2, Category: "dos", IsAttack: true},
		{ID: 3, Category: "fuzzers", IsAttack: true},
		{ID: 4, Category: "fuzzers", IsAttack: true},
	}
	alerts := baseline.AlertTimeline{
		1: {"dos": true},
	}
	result := Compute(flows, alerts, 50)
	if result.Recall != 0.5 {
		t.Fatalf("expected recall 0.5, got %f", result.Recall)
	}
}

func TestComputeNoAttackFlows(t *testing.T) {
	flows := []dataset.Flow{
		{ID: 1, Category: "normal", IsAttack: false},
	}
	alerts := baseline.AlertTimeline{}
	result := Compute(flows, alerts, 0)
	if result.Recall != 0.0 {
		t.Fatalf("expected recall 0.0 for no attacks, got %f", result.Recall)
	}
}

func TestComputeBandwidth(t *testing.T) {
	flows := []dataset.Flow{}
	alerts := baseline.AlertTimeline{}
	result := Compute(flows, alerts, 500)
	if result.Bandwidth != 500 {
		t.Fatalf("expected bandwidth 500, got %d", result.Bandwidth)
	}
}

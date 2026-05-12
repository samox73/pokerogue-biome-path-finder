package graph

import (
	"math"
	"testing"
)

func TestSameNode(t *testing.T) {
	g := New()
	r := g.ShortestPathGuaranteed("Town", "Town")
	if r == nil {
		t.Fatal("expected path for same node")
	}
	if r.TotalHops != 0 {
		t.Errorf("expected 0 hops, got %d", r.TotalHops)
	}
	if r.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", r.Probability)
	}

	r2 := g.ShortestPathWeighted("Town", "Town")
	if r2 == nil {
		t.Fatal("expected path for same node (weighted)")
	}
	if r2.TotalHops != 0 {
		t.Errorf("expected 0 hops, got %d", r2.TotalHops)
	}
}

func TestTownToGrassyFieldGuaranteed(t *testing.T) {
	g := New()
	r := g.ShortestPathGuaranteed("Town", "Grassy Field")
	if r == nil {
		t.Fatal("expected path from Town to Grassy Field")
	}
	// Town -> Plains -> Grassy Field = 2 hops
	if r.TotalHops != 2 {
		t.Errorf("expected 2 hops, got %d", r.TotalHops)
	}
	if r.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", r.Probability)
	}
	if r.WeightedLen != 2.0 {
		t.Errorf("expected weighted length 2.0, got %f", r.WeightedLen)
	}

	expected := []string{"Town", "Plains", "Grassy Field"}
	if len(r.Steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(r.Steps))
	}
	for i, s := range r.Steps {
		if s.Biome != expected[i] {
			t.Errorf("step %d: expected %q, got %q", i, expected[i], s.Biome)
		}
	}
}

func TestNoPathToTown(t *testing.T) {
	g := New()
	// Town has no inbound edges, so nothing can reach it.
	r := g.ShortestPathGuaranteed("Plains", "Town")
	if r != nil {
		t.Error("expected no guaranteed path to Town")
	}
	r2 := g.ShortestPathWeighted("Plains", "Town")
	if r2 != nil {
		t.Error("expected no weighted path to Town")
	}
}

func TestTownToSpaceNoGuaranteedPath(t *testing.T) {
	g := New()
	// All edges into Space are probabilistic (0.5 or 0.33),
	// so no guaranteed-only path to Space exists.
	r := g.ShortestPathGuaranteed("Town", "Space")
	if r != nil {
		t.Error("expected no guaranteed path to Space (all inbound edges are probabilistic)")
	}
}

func TestTownToSpaceWeighted(t *testing.T) {
	g := New()
	r := g.ShortestPathWeighted("Town", "Space")
	if r == nil {
		t.Fatal("expected a weighted path from Town to Space")
	}
	t.Logf("Weighted path Town->Space: %d hops, prob=%.4f, weight=%.2f", r.TotalHops, r.Probability, r.WeightedLen)
	for i, s := range r.Steps {
		if s.Edge != nil {
			t.Logf("  [%d] %s -> %s (%.0f%%)", i, s.Edge.From, s.Edge.To, s.Edge.Probability*100)
		} else {
			t.Logf("  [%d] %s (start)", i, s.Biome)
		}
	}
	// Path must use at least one probabilistic edge to reach Space.
	hasProbabilistic := false
	for _, s := range r.Steps {
		if s.Edge != nil && s.Edge.Probability < 1.0 {
			hasProbabilistic = true
		}
	}
	if !hasProbabilistic {
		t.Error("weighted path to Space should include a probabilistic edge")
	}
}

func TestTownToVolcanoGuaranteed(t *testing.T) {
	g := New()
	r := g.ShortestPathGuaranteed("Town", "Volcano")
	if r == nil {
		t.Fatal("expected guaranteed path from Town to Volcano")
	}
	if r.Probability != 1.0 {
		t.Errorf("expected probability 1.0, got %f", r.Probability)
	}
	t.Logf("Guaranteed path Town->Volcano: %d hops", r.TotalHops)
	for i, s := range r.Steps {
		if s.Edge != nil {
			t.Logf("  [%d] %s -> %s", i, s.Edge.From, s.Edge.To)
		}
	}
}

func TestWeightedBetterOrEqual(t *testing.T) {
	g := New()
	// For any pair with a guaranteed path, weighted should have <= weight.
	pairs := [][2]string{
		{"Town", "Volcano"},
		{"Town", "Grassy Field"},
		{"Town", "Beach"},
		{"Town", "Desert"},
	}
	for _, p := range pairs {
		guar := g.ShortestPathGuaranteed(p[0], p[1])
		weighted := g.ShortestPathWeighted(p[0], p[1])
		if guar == nil || weighted == nil {
			continue
		}
		if weighted.WeightedLen > guar.WeightedLen+0.001 {
			t.Errorf("%s->%s: weighted (%.2f) worse than guaranteed (%.2f)",
				p[0], p[1], weighted.WeightedLen, guar.WeightedLen)
		}
	}
}

func TestBiomeCount(t *testing.T) {
	if len(Biomes) != 34 {
		t.Errorf("expected 34 biomes, got %d", len(Biomes))
	}
}

func TestEdgeWeights(t *testing.T) {
	g := New()
	for from, edges := range g.Adj {
		for _, e := range edges {
			expectedWeight := 1.0 / e.Probability
			if math.Abs(e.Weight-expectedWeight) > 0.001 {
				t.Errorf("edge %s->%s: weight %.4f != 1/%.4f = %.4f",
					from, e.To, e.Weight, e.Probability, expectedWeight)
			}
		}
	}
}

func TestGuaranteedEdgesSubset(t *testing.T) {
	g := New()
	for from, edges := range g.AdjGuar {
		for _, e := range edges {
			if e.Probability != 1.0 {
				t.Errorf("guaranteed adj contains non-guaranteed edge %s->%s (%.2f)",
					from, e.To, e.Probability)
			}
		}
	}
}

func TestCyclePlains(t *testing.T) {
	g := New()
	// Plains -> Lake -> Swamp -> Tall Grass -> Forest -> Meadow -> Plains (6 hops, all guaranteed)
	r := g.ShortestCycleGuaranteed("Plains")
	if r == nil {
		t.Fatal("expected a guaranteed cycle from Plains")
	}
	if r.Steps[0].Biome != "Plains" || r.Steps[len(r.Steps)-1].Biome != "Plains" {
		t.Errorf("cycle should start and end at Plains, got %s...%s",
			r.Steps[0].Biome, r.Steps[len(r.Steps)-1].Biome)
	}
	if r.Probability != 1.0 {
		t.Errorf("guaranteed cycle should have prob 1.0, got %f", r.Probability)
	}
	t.Logf("Guaranteed cycle from Plains: %d hops, weight=%.2f", r.TotalHops, r.WeightedLen)
	for i, s := range r.Steps {
		if s.Edge != nil {
			t.Logf("  [%d] %s -> %s (%.0f%%)", i, s.Edge.From, s.Edge.To, s.Edge.Probability*100)
		} else {
			t.Logf("  [%d] %s (start)", i, s.Biome)
		}
	}
}

func TestCycleVolcano(t *testing.T) {
	g := New()
	// Volcano -> Beach -> Sea -> Seabed -> Cave -> Badlands -> Mountain -> Volcano
	// or a shorter one via Seabed -> Volcano (0.33) in weighted mode.
	r := g.ShortestCycleGuaranteed("Volcano")
	if r == nil {
		t.Fatal("expected a guaranteed cycle from Volcano")
	}
	if r.Steps[0].Biome != "Volcano" || r.Steps[len(r.Steps)-1].Biome != "Volcano" {
		t.Error("cycle should start and end at Volcano")
	}
	t.Logf("Guaranteed cycle from Volcano: %d hops", r.TotalHops)
	for i, s := range r.Steps {
		if s.Edge != nil {
			t.Logf("  [%d] %s -> %s", i, s.Edge.From, s.Edge.To)
		}
	}
}

func TestCycleWeightedVolcano(t *testing.T) {
	g := New()
	r := g.ShortestCycleWeighted("Volcano")
	if r == nil {
		t.Fatal("expected a weighted cycle from Volcano")
	}
	t.Logf("Weighted cycle from Volcano: %d hops, prob=%.4f, weight=%.2f",
		r.TotalHops, r.Probability, r.WeightedLen)
	for i, s := range r.Steps {
		if s.Edge != nil {
			t.Logf("  [%d] %s -> %s (%.0f%%)", i, s.Edge.From, s.Edge.To, s.Edge.Probability*100)
		}
	}

	// Weighted cycle should be no worse than guaranteed.
	guar := g.ShortestCycleGuaranteed("Volcano")
	if guar != nil && r.WeightedLen > guar.WeightedLen+0.001 {
		t.Errorf("weighted cycle (%.2f) worse than guaranteed (%.2f)",
			r.WeightedLen, guar.WeightedLen)
	}
}

func TestNoCycleTown(t *testing.T) {
	g := New()
	// Town has no inbound edges, so no cycle is possible.
	r := g.ShortestCycleGuaranteed("Town")
	if r != nil {
		t.Error("expected no guaranteed cycle from Town")
	}
	r2 := g.ShortestCycleWeighted("Town")
	if r2 != nil {
		t.Error("expected no weighted cycle from Town")
	}
}

func TestNoCycleSpace(t *testing.T) {
	g := New()
	// Space -> Ancient Ruins -> Mountain -> ... no guaranteed path back to Space.
	r := g.ShortestCycleGuaranteed("Space")
	if r != nil {
		t.Error("expected no guaranteed cycle from Space")
	}
	// But weighted cycle should exist via Mountain->Space (0.33) or other probabilistic edges.
	r2 := g.ShortestCycleWeighted("Space")
	if r2 == nil {
		t.Fatal("expected a weighted cycle from Space")
	}
	t.Logf("Weighted cycle from Space: %d hops, prob=%.4f, weight=%.2f",
		r2.TotalHops, r2.Probability, r2.WeightedLen)
	for i, s := range r2.Steps {
		if s.Edge != nil {
			t.Logf("  [%d] %s -> %s (%.0f%%)", i, s.Edge.From, s.Edge.To, s.Edge.Probability*100)
		}
	}
}

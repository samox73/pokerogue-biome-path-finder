package ui

import (
	"strings"
	"testing"

	"biome-path-finder/graph"
)

func TestBiomeToSlug(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Ancient Ruins", "ancient_ruins"},
		{"Grassy Field", "grassy_field"},
		{"Town", "town"},
		{"Ice Cave", "ice_cave"},
	}
	for _, tc := range tests {
		got := biomeToSlug(tc.input)
		if got != tc.want {
			t.Errorf("biomeToSlug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestBiomeLink(t *testing.T) {
	got := biomeLink("Ancient Ruins")
	want := "[[biomes:ancient_ruins|Ancient Ruins]]"
	if got != want {
		t.Errorf("biomeLink = %q, want %q", got, want)
	}
}

func TestFormatCycleDokuWiki(t *testing.T) {
	g := graph.New()
	guaranteed := g.ShortestCycleGuaranteed("Plains")
	risky := g.ShortestCycleWeighted("Plains")

	text := formatCycleDokuWiki(guaranteed, risky, "Plains")

	if !strings.Contains(text, "==== Cycles ====") {
		t.Error("missing Cycles header")
	}
	if !strings.Contains(text, "Guaranteed Cycle") {
		t.Error("missing Guaranteed Cycle section")
	}
	if !strings.Contains(text, "[[biomes:plains|Plains]]") {
		t.Error("missing Plains biome link")
	}
	if !strings.Contains(text, "**Hops:**") {
		t.Error("missing Hops stat")
	}
	// Guaranteed cycle should NOT have Probability or Expected transitions.
	guarSection := text[:strings.Index(text, "Alternative Cycle")]
	if strings.Contains(guarSection, "Probability") {
		t.Error("guaranteed cycle should not have Probability")
	}

	t.Log("DokuWiki output for Plains cycle:\n" + text)
}

func TestFormatCycleDokuWikiNoCycle(t *testing.T) {
	text := formatCycleDokuWiki(nil, nil, "Town")
	if !strings.Contains(text, "No cycles") {
		t.Errorf("expected no-cycle message, got: %s", text)
	}
}

func TestFormatCycleDokuWikiRiskyOnly(t *testing.T) {
	g := graph.New()
	guaranteed := g.ShortestCycleGuaranteed("Space")
	risky := g.ShortestCycleWeighted("Space")

	if guaranteed != nil {
		t.Skip("Space unexpectedly has a guaranteed cycle")
	}

	text := formatCycleDokuWiki(guaranteed, risky, "Space")
	if !strings.Contains(text, "Risky Cycle") {
		t.Error("missing Risky Cycle section")
	}
	if strings.Contains(text, "Guaranteed Cycle") {
		t.Error("should not have Guaranteed Cycle section")
	}
	if !strings.Contains(text, "**!!**") {
		t.Error("risky steps should be marked with !!")
	}

	t.Log("DokuWiki output for Space cycle:\n" + text)
}

func TestFormatCycleDokuWikiBothDifferent(t *testing.T) {
	g := graph.New()
	guaranteed := g.ShortestCycleGuaranteed("Volcano")
	risky := g.ShortestCycleWeighted("Volcano")

	text := formatCycleDokuWiki(guaranteed, risky, "Volcano")
	if !strings.Contains(text, "Guaranteed Cycle") {
		t.Error("missing Guaranteed Cycle")
	}
	if !strings.Contains(text, "Risky Cycle") {
		t.Error("missing Risky Cycle (should differ from guaranteed)")
	}

	t.Log("DokuWiki output for Volcano cycles:\n" + text)
}

func TestFormatRoutesDokuWiki(t *testing.T) {
	g := graph.New()
	guaranteed := g.ShortestPathGuaranteed("Town", "Volcano")
	risky := g.ShortestPathWeighted("Town", "Volcano")

	text := formatRoutesDokuWiki(guaranteed, risky)

	if !strings.Contains(text, "Guaranteed Route") {
		t.Error("missing Guaranteed Route section")
	}
	if !strings.Contains(text, "**Hops:**") {
		t.Error("missing Hops stat")
	}
	t.Log("DokuWiki output for Town->Volcano routes:\n" + text)
}

func TestFormatRoutesDokuWikiNoPath(t *testing.T) {
	text := formatRoutesDokuWiki(nil, nil)
	if !strings.Contains(text, "No path") {
		t.Errorf("expected no-path message, got: %s", text)
	}
}

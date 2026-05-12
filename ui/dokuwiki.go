package ui

import (
	"fmt"
	"strings"

	"biome-path-finder/graph"
)

func biomeToSlug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func biomeLink(name string) string {
	return fmt.Sprintf("[[biomes:%s|%s]]", biomeToSlug(name), name)
}

// formatCycleDokuWiki generates DokuWiki markup for the cycles section of a biome page.
func formatCycleDokuWiki(guaranteed, risky *graph.PathResult, biome string) string {
	if guaranteed == nil && risky == nil {
		return fmt.Sprintf("No cycles exist for %s.\n", biome)
	}

	var b strings.Builder

	b.WriteString("<WRAP centeralign>\n")
	b.WriteString("==== Cycles ====\n")
	b.WriteString("</WRAP>\n\n")

	if guaranteed != nil {
		writeGuaranteedDokuWiki(&b, guaranteed, "Guaranteed Cycle")
	}

	if risky != nil && !samePath(guaranteed, risky) {
		if guaranteed != nil {
			b.WriteString("\n")
		}
		label := riskyDokuWikiLabel(risky, true)
		writeRiskyDokuWiki(&b, risky, label)
	}

	return b.String()
}

// formatRoutesDokuWiki generates DokuWiki markup for path results (src != dst).
func formatRoutesDokuWiki(guaranteed, risky *graph.PathResult) string {
	if guaranteed == nil && risky == nil {
		return "No path found.\n"
	}

	var b strings.Builder

	if guaranteed != nil {
		writeGuaranteedDokuWiki(&b, guaranteed, "Guaranteed Route")
	}

	if risky != nil && !samePath(guaranteed, risky) {
		if guaranteed != nil {
			b.WriteString("\n")
		}
		label := riskyDokuWikiLabel(risky, false)
		writeRiskyDokuWiki(&b, risky, label)
	}

	return b.String()
}

// writeGuaranteedDokuWiki outputs just the route and hop count.
func writeGuaranteedDokuWiki(b *strings.Builder, result *graph.PathResult, title string) {
	b.WriteString(fmt.Sprintf("=== %s ===\n\n", title))

	var biomes []string
	for _, s := range result.Steps {
		biomes = append(biomes, biomeLink(s.Biome))
	}
	b.WriteString(strings.Join(biomes, " -> "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("**Hops:** %d\n", result.TotalHops))
}

// writeRiskyDokuWiki outputs the full detail with per-step probabilities.
func writeRiskyDokuWiki(b *strings.Builder, result *graph.PathResult, title string) {
	b.WriteString(fmt.Sprintf("=== %s ===\n\n", title))

	var biomes []string
	for _, s := range result.Steps {
		biomes = append(biomes, biomeLink(s.Biome))
	}
	b.WriteString(strings.Join(biomes, " -> "))
	b.WriteString("\n\n")

	for _, s := range result.Steps {
		if s.Edge == nil {
			continue
		}
		prob := s.Edge.Probability * 100
		marker := ""
		if s.Edge.Probability < 1.0 {
			marker = " **!!**"
		}
		b.WriteString(fmt.Sprintf("  - %s -> %s (%.0f%%)%s\n",
			biomeLink(s.Edge.From),
			biomeLink(s.Edge.To),
			prob,
			marker,
		))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("**Hops:** %d  **Probability:** %.1f%%  **Expected transitions:** %.2f\n",
		result.TotalHops,
		result.Probability*100,
		result.WeightedLen,
	))
}

func riskyDokuWikiLabel(result *graph.PathResult, isCycle bool) string {
	if hasRiskyEdge(result) {
		if isCycle {
			return "Risky Cycle"
		}
		return "Risky Route"
	}
	if isCycle {
		return "Alternative Cycle"
	}
	return "Alternative Route"
}

// samePath returns true if two results follow the same sequence of biomes.
func samePath(a, b *graph.PathResult) bool {
	if a == nil || b == nil {
		return false
	}
	if len(a.Steps) != len(b.Steps) {
		return false
	}
	for i := range a.Steps {
		if a.Steps[i].Biome != b.Steps[i].Biome {
			return false
		}
	}
	return true
}

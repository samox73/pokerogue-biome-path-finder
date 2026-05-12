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
func formatCycleDokuWiki(guaranteed, risky []*graph.PathResult, biome string) string {
	if len(guaranteed) == 0 && len(risky) == 0 {
		return fmt.Sprintf("No cycles exist for %s.\n", biome)
	}

	var b strings.Builder

	b.WriteString("<WRAP centeralign>\n")
	b.WriteString("==== Cycles ====\n")
	b.WriteString("</WRAP>\n\n")

	if len(guaranteed) > 0 {
		label := "Guaranteed Cycle"
		if len(guaranteed) > 1 {
			label += "s"
		}
		b.WriteString(fmt.Sprintf("=== %s ===\n\n", label))
		for i, r := range guaranteed {
			if i > 0 {
				b.WriteString("\n")
			}
			writeGuaranteedDokuWiki(&b, r, len(guaranteed) > 1, i+1)
		}
	}

	uniqueRisky := filterUnique(risky, guaranteed)

	if len(uniqueRisky) > 0 {
		if len(guaranteed) > 0 {
			b.WriteString("\n")
		}
		label := riskyDokuWikiLabel(uniqueRisky, true)
		b.WriteString(fmt.Sprintf("=== %s ===\n\n", label))
		for i, r := range uniqueRisky {
			if i > 0 {
				b.WriteString("\n")
			}
			writeRiskyDokuWiki(&b, r, len(uniqueRisky) > 1, i+1)
		}
	}

	return b.String()
}

// formatRoutesDokuWiki generates DokuWiki markup for path results (src != dst).
func formatRoutesDokuWiki(guaranteed, risky []*graph.PathResult) string {
	if len(guaranteed) == 0 && len(risky) == 0 {
		return "No path found.\n"
	}

	var b strings.Builder

	if len(guaranteed) > 0 {
		label := "Guaranteed Route"
		if len(guaranteed) > 1 {
			label += "s"
		}
		b.WriteString(fmt.Sprintf("=== %s ===\n\n", label))
		for i, r := range guaranteed {
			if i > 0 {
				b.WriteString("\n")
			}
			writeGuaranteedDokuWiki(&b, r, len(guaranteed) > 1, i+1)
		}
	}

	uniqueRisky := filterUnique(risky, guaranteed)

	if len(uniqueRisky) > 0 {
		if len(guaranteed) > 0 {
			b.WriteString("\n")
		}
		label := riskyDokuWikiLabel(uniqueRisky, false)
		b.WriteString(fmt.Sprintf("=== %s ===\n\n", label))
		for i, r := range uniqueRisky {
			if i > 0 {
				b.WriteString("\n")
			}
			writeRiskyDokuWiki(&b, r, len(uniqueRisky) > 1, i+1)
		}
	}

	return b.String()
}

// writeGuaranteedDokuWiki outputs the compact route and hop count.
func writeGuaranteedDokuWiki(b *strings.Builder, result *graph.PathResult, numbered bool, num int) {
	if numbered {
		b.WriteString(fmt.Sprintf("**#%d** ", num))
	}
	var biomes []string
	for _, s := range result.Steps {
		biomes = append(biomes, biomeLink(s.Biome))
	}
	b.WriteString(strings.Join(biomes, " -> "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("**Hops:** %d\n", result.TotalHops))
}

// writeRiskyDokuWiki outputs the full detail with per-step probabilities.
func writeRiskyDokuWiki(b *strings.Builder, result *graph.PathResult, numbered bool, num int) {
	if numbered {
		b.WriteString(fmt.Sprintf("**#%d** ", num))
	}
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

func riskyDokuWikiLabel(results []*graph.PathResult, isCycle bool) string {
	anyRisky := false
	for _, r := range results {
		if hasRiskyEdge(r) {
			anyRisky = true
			break
		}
	}
	noun := "Route"
	if isCycle {
		noun = "Cycle"
	}
	prefix := "Alternative"
	if anyRisky {
		prefix = "Risky"
	}
	label := prefix + " " + noun
	if len(results) > 1 {
		label += "s"
	}
	return label
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

package ui

import (
	"fmt"
	"strings"

	"biome-path-finder/graph"
)

func renderAllResults(guaranteed, risky []*graph.PathResult, isCycle bool) string {
	if len(guaranteed) == 0 && len(risky) == 0 {
		label := "No path found between these biomes."
		if isCycle {
			label = "No cycle exists for this biome."
		}
		return noPathStyle.Render(label)
	}

	var b strings.Builder

	if len(guaranteed) > 0 {
		label := "GUARANTEED ROUTE"
		if isCycle {
			label = "GUARANTEED CYCLE"
		}
		if len(guaranteed) > 1 {
			label += "S"
		}
		b.WriteString(sectionTitleStyle.Render(label))
		b.WriteString("\n\n")
		for i, r := range guaranteed {
			if i > 0 {
				b.WriteString("\n\n")
			}
			if len(guaranteed) > 1 {
				b.WriteString(statValueStyle.Render(fmt.Sprintf("#%d ", i+1)))
			}
			writeGuaranteedSummary(&b, r)
		}
	}

	// Filter risky results that duplicate a guaranteed result.
	uniqueRisky := filterUnique(risky, guaranteed)

	if len(uniqueRisky) > 0 {
		if len(guaranteed) > 0 {
			b.WriteString("\n\n")
		}
		label := riskySetLabel(uniqueRisky, isCycle)
		if len(uniqueRisky) > 1 {
			label += "S"
		}
		b.WriteString(sectionTitleStyle.Render(strings.ToUpper(label)))
		b.WriteString("\n\n")
		for i, r := range uniqueRisky {
			if i > 0 {
				b.WriteString("\n\n")
			}
			if len(uniqueRisky) > 1 {
				b.WriteString(statValueStyle.Render(fmt.Sprintf("#%d ", i+1)))
			}
			writeRiskyDetails(&b, r)
		}
	}

	return b.String()
}

func writeGuaranteedSummary(b *strings.Builder, result *graph.PathResult) {
	var biomes []string
	for _, s := range result.Steps {
		biomes = append(biomes, pathBiomeStyle.Render(s.Biome))
	}
	b.WriteString(strings.Join(biomes, arrowStyle.Render(" -> ")))
	b.WriteString("\n\n")
	b.WriteString(statLabelStyle.Render("Hops: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%d", result.TotalHops)))
}

func writeRiskyDetails(b *strings.Builder, result *graph.PathResult) {
	var biomes []string
	for _, s := range result.Steps {
		biomes = append(biomes, pathBiomeStyle.Render(s.Biome))
	}
	b.WriteString(strings.Join(biomes, arrowStyle.Render(" -> ")))
	b.WriteString("\n\n")

	stepNum := 0
	for _, s := range result.Steps {
		if s.Edge == nil {
			continue
		}
		stepNum++
		prob := s.Edge.Probability * 100
		probStr := fmt.Sprintf("%.0f%%", prob)
		pStyle := probNormalStyle
		if s.Edge.Probability < 1.0 {
			pStyle = probRiskyStyle
		}
		b.WriteString(stepStyle.Render(fmt.Sprintf("[%d] %s -> %s  ", stepNum, s.Edge.From, s.Edge.To)))
		b.WriteString(pStyle.Render(probStr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(statLabelStyle.Render("Hops: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%d", result.TotalHops)))
	b.WriteString("  ")
	b.WriteString(statLabelStyle.Render("Prob: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%.1f%%", result.Probability*100)))
}

func hasRiskyEdge(result *graph.PathResult) bool {
	for _, s := range result.Steps {
		if s.Edge != nil && s.Edge.Probability < 1.0 {
			return true
		}
	}
	return false
}

// riskySetLabel returns the appropriate header label.
func riskySetLabel(results []*graph.PathResult, isCycle bool) string {
	anyRisky := false
	for _, r := range results {
		if hasRiskyEdge(r) {
			anyRisky = true
			break
		}
	}
	if anyRisky {
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

// filterUnique returns results from candidates that don't match any path in exclude.
func filterUnique(candidates, exclude []*graph.PathResult) []*graph.PathResult {
	var out []*graph.PathResult
	for _, c := range candidates {
		if !pathInSet(c, exclude) {
			out = append(out, c)
		}
	}
	return out
}

func pathInSet(result *graph.PathResult, set []*graph.PathResult) bool {
	for _, r := range set {
		if samePath(result, r) {
			return true
		}
	}
	return false
}

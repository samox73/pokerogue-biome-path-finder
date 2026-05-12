package ui

import (
	"fmt"
	"strings"

	"biome-path-finder/graph"

	"github.com/charmbracelet/lipgloss"
)

func renderAllResults(guaranteed, risky *graph.PathResult, isCycle bool) string {
	if guaranteed == nil && risky == nil {
		label := "No path found between these biomes."
		if isCycle {
			label = "No cycle exists for this biome."
		}
		return noPathStyle.Render(label)
	}

	var b strings.Builder

	if guaranteed != nil {
		label := "GUARANTEED ROUTE"
		if isCycle {
			label = "GUARANTEED CYCLE"
		}
		b.WriteString(resultTitleStyle.Render(label))
		b.WriteString("\n\n")
		writeGuaranteedSummary(&b, guaranteed)
	}

	if risky != nil && !samePath(guaranteed, risky) {
		if guaranteed != nil {
			b.WriteString("\n\n")
		}
		label := riskyRouteLabel(risky, isCycle)
		style := riskyTitleStyle
		if !hasRiskyEdge(risky) {
			style = resultTitleStyle
		}
		b.WriteString(style.Render(strings.ToUpper(label)))
		b.WriteString("\n\n")
		writeRiskyDetails(&b, risky)
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
		var style lipgloss.Style
		if s.Edge.Probability < 1.0 {
			style = riskyStyle
			probStr = fmt.Sprintf("%.0f%% !", prob)
		} else {
			style = guaranteedStyle
		}
		b.WriteString(fmt.Sprintf("[%d] %s -> %s  %s\n",
			stepNum, s.Edge.From, s.Edge.To, style.Render(probStr)))
	}

	b.WriteString("\n")
	b.WriteString(statLabelStyle.Render("Hops: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%d", result.TotalHops)))
	b.WriteString("  ")
	b.WriteString(statLabelStyle.Render("Prob: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%.1f%%", result.Probability*100)))
	b.WriteString("  ")
	b.WriteString(statLabelStyle.Render("Expected transitions: "))
	b.WriteString(statValueStyle.Render(fmt.Sprintf("%.2f", result.WeightedLen)))
}

func hasRiskyEdge(result *graph.PathResult) bool {
	for _, s := range result.Steps {
		if s.Edge != nil && s.Edge.Probability < 1.0 {
			return true
		}
	}
	return false
}

func riskyRouteLabel(result *graph.PathResult, isCycle bool) string {
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

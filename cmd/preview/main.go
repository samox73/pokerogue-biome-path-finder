package main

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"biome-path-finder/graph"
)

func main() {
	g := graph.New()

	var sections []string

	for _, biome := range g.Biomes {
		guaranteed := g.ShortestCycleGuaranteed(biome)
		risky := g.ShortestCycleWeighted(biome)

		if len(guaranteed) == 0 && len(risky) == 0 {
			continue
		}

		wiki := formatCycleDokuWiki(guaranteed, risky, biome)
		rendered := dokuwikiToHTML(wiki)

		sections = append(sections, fmt.Sprintf(
			"<div class=\"biome-section\">\n<h2>%s</h2>\n%s\n</div>",
			html.EscapeString(biome),
			rendered,
		))
	}

	page := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>PokéRogue Biome Cycles Preview</title>
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    max-width: 900px;
    margin: 2em auto;
    padding: 0 1em;
    background: #1a1a2e;
    color: #e0e0e0;
  }
  h1 { color: #ff6600; text-align: center; }
  h2 { color: #aaaaff; border-bottom: 1px solid #333; padding-bottom: 0.3em; }
  h3 { color: #00cc88; }
  h3.risky { color: #ff8800; }
  h3.alternative { color: #00cc88; }
  .biome-section {
    background: #16213e;
    border: 1px solid #333;
    border-radius: 8px;
    padding: 1em 1.5em;
    margin-bottom: 1.5em;
  }
  a { color: #6699ff; text-decoration: none; }
  a:hover { text-decoration: underline; }
  .route { font-size: 1.05em; margin: 0.5em 0; }
  ul { list-style: none; padding-left: 1em; }
  ul li::before { content: "→ "; color: #666; }
  .warn { color: #ff4444; font-weight: bold; }
  .stats { color: #888; margin-top: 0.5em; }
  .stats b { color: #fff; }
  .centered { text-align: center; }
</style>
</head>
<body>
<h1>PokéRogue Biome Cycles Preview</h1>
<p style="text-align:center;color:#888;">This is how the DokuWiki cycle sections would render on the wiki.</p>
%s
</body>
</html>`, strings.Join(sections, "\n"))

	outPath := "preview.html"
	if err := os.WriteFile(outPath, []byte(page), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}
	fmt.Printf("Preview written to %s\n", outPath)
}

// --- DokuWiki formatting (duplicated from ui package to avoid import issues) ---

func biomeToSlug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func biomeLink(name string) string {
	return fmt.Sprintf("[[biomes:%s|%s]]", biomeToSlug(name), name)
}

func formatCycleDokuWiki(guaranteed, risky []*graph.PathResult, biome string) string {
	if len(guaranteed) == 0 && len(risky) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<WRAP centeralign>\n==== Cycles ====\n</WRAP>\n\n")

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
			if len(guaranteed) > 1 {
				b.WriteString(fmt.Sprintf("**#%d** ", i+1))
			}
			var biomes []string
			for _, s := range r.Steps {
				biomes = append(biomes, biomeLink(s.Biome))
			}
			b.WriteString(strings.Join(biomes, " -> "))
			b.WriteString("\n\n")
			b.WriteString(fmt.Sprintf("**Hops:** %d\n", r.TotalHops))
		}
	}

	uniqueRisky := filterUniquePreview(risky, guaranteed)

	if len(uniqueRisky) > 0 {
		if len(guaranteed) > 0 {
			b.WriteString("\n")
		}
		label := riskyLabel(uniqueRisky, true)
		b.WriteString(fmt.Sprintf("=== %s ===\n\n", label))
		for i, r := range uniqueRisky {
			if i > 0 {
				b.WriteString("\n")
			}
			if len(uniqueRisky) > 1 {
				b.WriteString(fmt.Sprintf("**#%d** ", i+1))
			}
			var biomes []string
			for _, s := range r.Steps {
				biomes = append(biomes, biomeLink(s.Biome))
			}
			b.WriteString(strings.Join(biomes, " -> "))
			b.WriteString("\n\n")
			for _, s := range r.Steps {
				if s.Edge == nil {
					continue
				}
				prob := s.Edge.Probability * 100
				b.WriteString(fmt.Sprintf("  - %s -> %s (%.0f%%)\n",
					biomeLink(s.Edge.From), biomeLink(s.Edge.To), prob))
			}
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("**Hops:** %d  **Probability:** %.1f%%  **Expected transitions:** %.2f\n",
				r.TotalHops, r.Probability*100, r.WeightedLen))
		}
	}

	return b.String()
}

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

func filterUniquePreview(candidates, exclude []*graph.PathResult) []*graph.PathResult {
	var out []*graph.PathResult
	for _, c := range candidates {
		found := false
		for _, e := range exclude {
			if samePath(c, e) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, c)
		}
	}
	return out
}

func riskyLabel(results []*graph.PathResult, isCycle bool) string {
	anyRisky := false
	for _, r := range results {
		for _, s := range r.Steps {
			if s.Edge != nil && s.Edge.Probability < 1.0 {
				anyRisky = true
				break
			}
		}
		if anyRisky {
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

// --- DokuWiki to HTML converter ---

var (
	reBiomeLink = regexp.MustCompile(`\[\[biomes:([^|]+)\|([^\]]+)\]\]`)
	reBold      = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reH4        = regexp.MustCompile(`(?m)^====\s*(.+?)\s*====$`)
	reH3        = regexp.MustCompile(`(?m)^===\s*(.+?)\s*===$`)
	reListItem  = regexp.MustCompile(`(?m)^\s+-\s+(.+)$`)
	reWrapCA    = regexp.MustCompile(`(?s)<WRAP centeralign>\n?(.*?)\n?</WRAP>`)
)

func dokuwikiToHTML(wiki string) string {
	s := html.EscapeString(wiki)

	// Unescape the DokuWiki syntax characters we need to parse.
	s = strings.ReplaceAll(s, "&#34;", "\"")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")

	// Biome links: [[biomes:slug|Name]] -> <a>Name</a>
	s = reBiomeLink.ReplaceAllString(s, `<a href="#$1">$2</a>`)

	// Bold: **text** -> <b>text</b>
	// Handle !! specially for warning markers.
	s = reBold.ReplaceAllStringFunc(s, func(m string) string {
		inner := reBold.FindStringSubmatch(m)[1]
		if inner == "!!" {
			return `<span class="warn">!!</span>`
		}
		return "<b>" + inner + "</b>"
	})

	// Headings (must process h4 before h3).
	s = reH4.ReplaceAllString(s, `<h3 class="centered">$1</h3>`)
	s = reH3.ReplaceAllStringFunc(s, func(m string) string {
		inner := reH3.FindStringSubmatch(m)[1]
		cls := ""
		lower := strings.ToLower(inner)
		if strings.Contains(lower, "risky") {
			cls = ` class="risky"`
		} else if strings.Contains(lower, "alternative") {
			cls = ` class="alternative"`
		}
		return fmt.Sprintf("<h3%s>%s</h3>", cls, inner)
	})

	// WRAP centeralign.
	s = reWrapCA.ReplaceAllString(s, `<div class="centered">$1</div>`)

	// List items.
	s = reListItem.ReplaceAllString(s, `<li>$1</li>`)
	// Wrap consecutive <li> in <ul>.
	s = regexp.MustCompile(`(?s)(<li>.*?</li>\n?)+`).ReplaceAllStringFunc(s, func(m string) string {
		return "<ul>\n" + m + "</ul>\n"
	})

	// Route lines (lines with biome links joined by ->).
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ` -> <a href=`) && !strings.HasPrefix(trimmed, "<") {
			lines[i] = `<p class="route">` + trimmed + "</p>"
		}
		// Stats lines.
		if strings.HasPrefix(trimmed, "<b>Hops:</b>") {
			lines[i] = `<p class="stats">` + trimmed + "</p>"
		}
	}
	s = strings.Join(lines, "\n")

	return s
}

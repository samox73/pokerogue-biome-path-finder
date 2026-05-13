package ui

import "github.com/charmbracelet/lipgloss"

// Two-tone palette: orange accent + grayscale.
const (
	colorAccent = "#FF6600" // orange — the single accent color
	colorFg     = "#E0E0E0" // light gray — normal text
	colorDim    = "#666666" // dim — secondary text, arrows, labels
	colorBorder = "#444444" // dark — borders, separators
	colorDimmer = "#444444" // very dim — dimmed items during filter
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(1, 2)

	// Section headers in results panel.
	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(colorAccent))

	// Biome names in path overview.
	pathBiomeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFg)).
			Bold(true)

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	// Stat labels and values.
	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	statValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorFg)).
			Bold(true)

	// Per-step probability: normal vs risky.
	probNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	probRiskyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true)

	// Step detail lines.
	stepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	// Borders for the two list panels.
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorAccent)).
				Padding(1, 2)

	unfocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorder)).
				Padding(1, 2)

	noPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim)).
			Padding(1, 0)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true).
			Padding(1, 0)

	// List title styles (the "Source" / "Destination" headers).
	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorAccent)).
			Padding(0, 0)

	listTitleBarStyle = lipgloss.NewStyle().
				Padding(0, 0, 1, 2)

	// List item styles.
	listNormalTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorFg)).
				Padding(0, 0, 0, 2)

	listSelectedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent)).
				Bold(true).
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.Color(colorAccent)).
				Padding(0, 0, 0, 1)

	listDimmedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorDimmer)).
				Padding(0, 0, 0, 2)

	listFilterMatchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent)).
				Bold(true)

	listFilterPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent))

	listFilterCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorAccent))
)

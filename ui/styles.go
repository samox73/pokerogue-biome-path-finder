package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6600")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1, 2)

	resultTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00CC88"))

	guaranteedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CC88"))

	riskyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true)

	pathBiomeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAFF"))

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	statValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FF6600")).
				Padding(1, 2)

	unfocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444444")).
				Padding(1, 2)

	noPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true).
			Padding(1, 0)

	riskyTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF8800"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CC88")).
			Bold(true).
			Padding(1, 0)
)

package ui

import (
	"fmt"
	"time"

	"biome-path-finder/graph"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focus int

const (
	focusSource focus = iota
	focusDest
)

type biomeItem string

func (b biomeItem) Title() string       { return string(b) }
func (b biomeItem) Description() string { return "" }
func (b biomeItem) FilterValue() string { return string(b) }

type clearStatusMsg struct{}

type Model struct {
	graph      *graph.Graph
	sourceList list.Model
	destList   list.Model
	focus      focus
	locked     bool // when true, both lists scroll together (cycle mode)

	guaranteed []*graph.PathResult
	risky      []*graph.PathResult
	isCycle    bool

	statusMsg string
	width     int
	height    int
}

func NewModel() Model {
	g := graph.New()

	items := make([]list.Item, len(g.Biomes))
	for i, b := range g.Biomes {
		items[i] = biomeItem(b)
	}

	srcDelegate := styledDelegate()
	srcList := list.New(items, srcDelegate, 30, 20)
	srcList.Title = "Source"
	applyListStyles(&srcList)

	destItems := make([]list.Item, len(items))
	copy(destItems, items)
	destDelegate := styledDelegate()
	destList := list.New(destItems, destDelegate, 30, 20)
	destList.Title = "Destination"
	applyListStyles(&destList)

	m := Model{
		graph:      g,
		sourceList: srcList,
		destList:   destList,
		focus:      focusSource,
	}
	m.recalculate()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := msg.Height - 10
		if listHeight < 5 {
			listHeight = 5
		}
		listWidth := msg.Width*2/5 - 8
		if listWidth < 20 {
			listWidth = 20
		}
		halfList := listWidth / 2
		m.sourceList.SetSize(halfList, listHeight)
		m.destList.SetSize(halfList, listHeight)
		return m, nil

	case tea.KeyMsg:
		if m.activeList().FilterState() == list.Filtering {
			if msg.String() == "esc" {
				l := m.activeListPtr()
				*l, _ = l.Update(msg)
				return m, nil
			}
			l := m.activeListPtr()
			var cmd tea.Cmd
			*l, cmd = l.Update(msg)
			m.recalculate()
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if !m.locked {
				if m.focus == focusSource {
					m.focus = focusDest
				} else {
					m.focus = focusSource
				}
			}
			return m, nil
		case "shift+tab":
			if !m.locked {
				if m.focus == focusDest {
					m.focus = focusSource
				} else {
					m.focus = focusDest
				}
			}
			return m, nil
		case "l":
			m.locked = !m.locked
			if m.locked {
				// Sync dest to match source cursor.
				m.destList.Select(m.sourceList.Index())
				m.recalculate()
			}
			return m, nil
		case "c":
			return m, m.copyToClipboard()
		}
	}

	if m.locked {
		// Forward navigation to both lists.
		var cmd tea.Cmd
		m.sourceList, cmd = m.sourceList.Update(msg)
		m.destList.Select(m.sourceList.Index())
		m.recalculate()
		return m, cmd
	}

	l := m.activeListPtr()
	var cmd tea.Cmd
	*l, cmd = l.Update(msg)
	m.recalculate()
	return m, cmd
}

func (m *Model) copyToClipboard() tea.Cmd {
	if len(m.guaranteed) == 0 && len(m.risky) == 0 {
		m.statusMsg = "Nothing to copy"
		return clearStatusAfter(2 * time.Second)
	}

	var text string
	if m.isCycle {
		text = formatCycleDokuWiki(m.guaranteed, m.risky, m.selectedSource())
	} else {
		text = formatRoutesDokuWiki(m.guaranteed, m.risky)
	}

	if err := clipboard.WriteAll(text); err != nil {
		m.statusMsg = fmt.Sprintf("Copy failed: %v", err)
	} else {
		m.statusMsg = "Copied to clipboard!"
	}
	return clearStatusAfter(2 * time.Second)
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m *Model) activeList() *list.Model {
	if m.focus == focusSource {
		return &m.sourceList
	}
	return &m.destList
}

func (m *Model) activeListPtr() *list.Model {
	if m.focus == focusSource {
		return &m.sourceList
	}
	return &m.destList
}

func (m *Model) selectedSource() string {
	item := m.sourceList.SelectedItem()
	if item == nil {
		return ""
	}
	return string(item.(biomeItem))
}

func (m *Model) selectedDest() string {
	item := m.destList.SelectedItem()
	if item == nil {
		return ""
	}
	return string(item.(biomeItem))
}

func (m *Model) recalculate() {
	src := m.selectedSource()
	dst := m.selectedDest()
	if src == "" || dst == "" {
		m.guaranteed = nil
		m.risky = nil
		m.isCycle = false
		return
	}
	if src == dst {
		m.isCycle = true
		m.guaranteed = m.graph.ShortestCycleGuaranteed(src)
		m.risky = m.graph.ShortestCycleWeighted(src)
	} else {
		m.isCycle = false
		m.guaranteed = m.graph.ShortestPathGuaranteed(src, dst)
		m.risky = m.graph.ShortestPathWeighted(src, dst)
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render("POKEROGUE BIOME PATH FINDER")

	srcStyle := unfocusedBorderStyle
	dstStyle := unfocusedBorderStyle
	if m.locked {
		srcStyle = focusedBorderStyle
		dstStyle = focusedBorderStyle
	} else if m.focus == focusSource {
		srcStyle = focusedBorderStyle
	} else {
		dstStyle = focusedBorderStyle
	}

	leftWidth := m.width*2/5 - 4
	if leftWidth < 20 {
		leftWidth = 20
	}
	halfWidth := leftWidth/2 - 2
	if halfWidth < 12 {
		halfWidth = 12
	}

	srcPanel := srcStyle.Width(halfWidth).Render(m.sourceList.View())
	dstPanel := dstStyle.Width(halfWidth).Render(m.destList.View())

	lists := lipgloss.JoinHorizontal(lipgloss.Top, srcPanel, " ", dstPanel)

	lockLabel := "[l] lock lists"
	if m.locked {
		lockLabel = "[l] unlock lists"
	}
	help := helpStyle.Render("[tab] switch list  [/] filter  " + lockLabel + "  [c] copy  [q] quit")

	var statusLine string
	if m.statusMsg != "" {
		statusLine = statusStyle.Render(m.statusMsg)
	}

	leftParts := []string{title, "", lists, "", help}
	if statusLine != "" {
		leftParts = append(leftParts, statusLine)
	}
	leftPanel := lipgloss.JoinVertical(lipgloss.Left, leftParts...)

	rightWidth := m.width - leftWidth - 10
	if rightWidth < 20 {
		rightWidth = 20
	}

	resultsContent := renderAllResults(m.guaranteed, m.risky, m.isCycle)
	rightPanel := panelStyle.Width(rightWidth).Render(resultsContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
}

// styledDelegate creates a list.DefaultDelegate with our unified color scheme.
func styledDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetHeight(1)

	d.Styles.NormalTitle = listNormalTitleStyle
	d.Styles.SelectedTitle = listSelectedTitleStyle
	d.Styles.DimmedTitle = listDimmedTitleStyle
	d.Styles.FilterMatch = listFilterMatchStyle

	return d
}

// applyListStyles configures list-level styles (title, filter prompt, etc.).
func applyListStyles(l *list.Model) {
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = listTitleBarStyle
	l.Styles.FilterPrompt = listFilterPromptStyle
	l.Styles.FilterCursor = listFilterCursorStyle
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
}

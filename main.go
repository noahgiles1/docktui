package main

import (
	conatiners "docktui/views"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/container"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionState uint

const (
	containerView sessionState = iota
	containerView1
)

type model struct {
	Tabs       []string
	TabContent []tea.Model
	activeTab  int
	state      sessionState
}

type subModel struct {
	model struct{}
}

func newModel() model {
	m := model{state: containerView}
	m.Tabs = []string{"Containers", "Test"}
	m.TabContent = []tea.Model{conatiners.New(), conatiners.New()}
	return m
}

func (m model) Init() tea.Cmd {
	// Initialize all tab content models
	var cmds []tea.Cmd
	for _, tabModel := range m.TabContent {
		if cmd := tabModel.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			m.TabContent[m.activeTab], cmd = m.TabContent[m.activeTab].Update(msg)
			return m, cmd
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			m.TabContent[m.activeTab], cmd = m.TabContent[m.activeTab].Update(msg)
			return m, cmd
		}
	case []container.Summary:
		{
			for i, tab := range m.TabContent {
				m.TabContent[i], cmd = tab.Update(msg)
			}
		}
	}
	m.TabContent[m.activeTab], cmd = m.TabContent[m.activeTab].Update(msg)
	return m, cmd
}

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border).Width(50)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab].View()))
	doc.WriteString("\n")
	s := docStyle.Render(doc.String()) + helpStyle.Render(fmt.Sprintf("\ntab: focus next • ↑ tab: focus prev • q: exit\n"))
	return s
}

func main() {
	p := tea.NewProgram(newModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true).Foreground(lipgloss.Color("#7D56F4"))
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Align(lipgloss.Center).Padding(0, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

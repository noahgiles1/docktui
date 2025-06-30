package placeholder

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content string
}

var width int

func New(cnt string, w int) Model {
	width = w
	return Model{
		content: cnt,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return contentStyling.Width(width * 3).Render(m.content)
}

var (
	contentStyling = lipgloss.NewStyle().Padding(4, 2, 0, 2).Foreground(lipgloss.Color("50")).Align(lipgloss.Center)
)

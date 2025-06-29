package placeholder

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content string
}

func New(cnt string) Model {
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
	return contentStyling.Render(m.content)
}

var (
	contentStyling = lipgloss.NewStyle().Padding(4, 2, 4, 2).Foreground(lipgloss.Color("50"))
)

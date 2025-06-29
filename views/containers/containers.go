package containers

import (
	"context"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sync/atomic"
)

type Model struct {
	table   table.Model
	content []container.Summary
	err     error
}

func New() Model {
	// Initialize table
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Image", Width: 20},
		{Title: "State", Width: 10},
		{Title: "Status", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}), // Start with empty rows
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		table: t,
	}
}

func (m Model) Init() tea.Cmd {
	return getDockerContainers
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case []container.Summary:
		// Update content and rebuild table rows
		m.content = msg
		var rows []table.Row
		for _, ctr := range m.content {
			rows = append(rows, table.Row{
				ctr.Names[0],
				ctr.Image,
				ctr.State,
				ctr.Status,
			})
		}
		m.table.SetRows(rows)
		return m, nil
	}

	// Let the table handle all key events (including down/up arrows)
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	// Just render the table - don't recreate it!
	return baseStyle.Render(m.table.View())

}

func getDockerContainers() tea.Msg {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer func(cli *client.Client) {
		err := cli.Close()
		if err != nil {

		}
	}(cli)

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}
	return containers
}

var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.HiddenBorder()).
		BorderForeground(lipgloss.Color("240"))
)

var lastID int64

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}

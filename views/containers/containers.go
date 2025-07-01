package containers

import (
	"context"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sort"
	"time"
)

const refreshInterval = 5 * time.Second

type Model struct {
	table      table.Model
	containers []container.Summary
	err        error
	height     int
	width      int
}

type containerChangeMsg struct{}
type tickMsg time.Time

func New() Model {
	// Initialize table
	columns := []table.Column{
		{Title: "", Width: 1},
		{Title: "Name", Width: 20},
		{Title: "Image", Width: 20},
		{Title: "State", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}), // Start with empty rows
		table.WithFocused(true),
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
		table:  t,
		width:  80,
		height: 20,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		getDockerContainers,
		tick(), // Start the periodic refresh
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update dimensions and table height
		m.width = msg.Width
		m.height = msg.Height

		// Calculate table height based on available space
		tableHeight := m.height - 8 // Reserve space for headers, borders, etc.
		if tableHeight < 3 {
			tableHeight = 3
		}

		m.table.SetHeight(tableHeight - 3)
		return m, nil
	case []container.Summary:
		// Update content and rebuild table rows
		m.containers = msg
		var rows []table.Row
		for _, ctr := range m.containers {
			rows = append(rows, table.Row{
				"",
				ctr.Names[0],
				ctr.Image,
				ctr.State,
			})
		}
		m.table.SetRows(rows)
	case containerChangeMsg:
		cmd = getDockerContainers
		cmds = append(cmds, cmd)
	case tickMsg:
		cmd = tea.Batch(
			getDockerContainers,
			tick())
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			if len(m.containers) > 0 && m.table.Cursor() < len(m.containers) {
				cmd = runContainer(m.containers[m.table.Cursor()].ID)
				cmds = append(cmds, cmd)
			}
		case "delete", "backspace":
			if len(m.containers) > 0 && m.table.Cursor() < len(m.containers) {
				cmd = stopContainer(m.containers[m.table.Cursor()].ID)
				cmds = append(cmds, cmd)
			}
		}
	}
	// Let the table handle all key events (including down/up arrows)
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	views := []string{
		baseStyle.Render(m.table.View()), // Just render the table - don't recreate it!
		helpStyle.Render("test")}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

func executeContainerOperation(containerId string, operation func(*client.Client, string) error) tea.Cmd {
	return func() tea.Msg {
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}
		defer func(cli *client.Client) {
			err := cli.Close()
			if err != nil {
			}
		}(cli)

		err = operation(cli, containerId)
		if err != nil {
			panic(err)
		}
		return containerChangeMsg{}
	}
}

func runContainer(containerId string) tea.Cmd {
	return executeContainerOperation(containerId, func(cli *client.Client, id string) error {
		return cli.ContainerStart(context.Background(), id, container.StartOptions{})
	})
}

func stopContainer(containerId string) tea.Cmd {
	return executeContainerOperation(containerId, func(cli *client.Client, id string) error {
		return cli.ContainerStop(context.Background(), id, container.StopOptions{})
	})
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
	// Sort containers by running state
	sort.Slice(containers, func(i, j int) bool {
		if containers[i].State != containers[j].State {
			return containers[i].State == "running"
		}
		return i < j
	})

	return containers
}

func tick() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var (
	baseStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("240")).
			Border(lipgloss.NormalBorder(), false, true, false, false)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Bottom)
)

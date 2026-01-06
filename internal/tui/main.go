package tui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

type MainWindowModel struct {
	width  int
	height int
	t      []tasks.Task
	s      *data.Service
}

func NewMainWindowModel(s *data.Service) MainWindowModel {
	return MainWindowModel{s: s}
}

func (m MainWindowModel) LoadTasks() tea.Msg {
	taskslist, err := m.s.TasksRepo.GetAll()
	if err != nil {
		log.Fatalf("Error retrieving tasks: %v", err)
	}
	return tasksLoadedMsg{t: taskslist}
}

type tasksLoadedMsg struct {
	t []tasks.Task
}

// View implements tea.Model.
func (m MainWindowModel) View() string {
	style := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).Height(m.height)

	body := "Welcome to taskerr tui mode. Press ctrl+c to exit.\n\n\n"
	for _, task := range m.t {
		body += fmt.Sprintf("- [%v] %s\n", task.State, task.Description)
	}

	return style.Render(body)
}

func (m MainWindowModel) Init() tea.Cmd {
	return m.LoadTasks
}

func (m MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		m.t = msg.t
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

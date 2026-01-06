package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data"
)

type MainWindowModel struct {
	width  int
	height int
	t      []TaskModel
	s      *data.Service
}

func NewMainWindowModel(s *data.Service) *MainWindowModel {
	return &MainWindowModel{s: s}
}

// View implements tea.Model.
func (m *MainWindowModel) View() string {
	windowStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Width(m.width - 4). // Adjust for padding
		Height(m.height - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("2")).
		Padding(1)

	var taskViews string
	for _, task := range m.t {
		taskViews += task.View() + "\n"
	}

	return windowStyle.Render(taskViews)
}

func (m *MainWindowModel) Init() tea.Cmd {
	taskslist, err := m.s.TasksRepo.GetAll()
	if err == nil {
		log.Fatalf("Error retrieving tasks: %v", err)
	}
	m.t = make([]TaskModel, 0)
	for _, t := range taskslist {
		m.t = append(m.t, NewTaskModel(t))
	}
	return nil
}

func (m *MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

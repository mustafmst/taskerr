package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

type MainWindowModel struct {
	width         int
	height        int
	t             []tasks.Task
	s             *data.Service
	activeTask    int
	hideCompleted bool
}

func NewMainWindowModel(s *data.Service) MainWindowModel {
	return MainWindowModel{s: s, activeTask: -1, hideCompleted: true}
}

func (m MainWindowModel) LoadTasks() tea.Msg {
	taskslist, err := m.s.TasksRepo.GetAll()
	if err != nil {
		log.Fatalf("Error retrieving tasks: %v", err)
	}
	if m.hideCompleted {
		filtered := make([]tasks.Task, 0)
		for _, t := range taskslist {
			if !t.State {
				filtered = append(filtered, t)
			}
		}
		taskslist = filtered
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

	body := "Welcome to taskerr tui mode. Press ctrl+c or q to exit.\n"
	body += "Press h to toggle hide completed tasks.\n"
	body += "Press j/k or down/up to navigate, enter to toggle task state.\n\n"

	for i, task := range m.t {
		active := false
		if i == m.activeTask {
			active = true
		}
		body += NewTaskModel(task, m.width, active).View() + "\n"
	}

	return style.Render(body)
}

func (m MainWindowModel) Init() tea.Cmd {
	return m.LoadTasks
}

func fixactivetask(m *MainWindowModel) {
	if m.activeTask >= len(m.t) {
		m.activeTask = len(m.t) - 1
	} else if m.activeTask < 0 {
		m.activeTask = 0
	}
}

func (m MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		m.t = msg.t
		if m.activeTask == -1 && len(m.t) > 0 {
			m.activeTask = 0
		}
		fixactivetask(&m)
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "j" || msg.String() == "down" {
			m.activeTask++
			fixactivetask(&m)
			return m, nil
		}
		if msg.String() == "k" || msg.String() == "up" {
			m.activeTask--
			fixactivetask(&m)
			return m, nil
		}
		if msg.String() == "enter" || msg.String() == "return" {
			if m.activeTask < 0 || m.activeTask >= len(m.t) {
				return m, nil
			}
			t := m.t[m.activeTask]
			t.ToggleState()
			m.s.TasksRepo.Update(&t)
			return m, m.LoadTasks
		}
		if msg.String() == "h" {
			m.hideCompleted = !m.hideCompleted
			return m, m.LoadTasks
		}
	}
	return m, nil
}

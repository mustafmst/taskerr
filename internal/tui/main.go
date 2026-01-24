package tui

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

type MainWindowModel struct {
	width         int
	height        int
	tasks         []tasks.Task
	service       *data.Service
	activeTask    int
	hideCompleted bool
	lastDBState   tasks.DBState
	scrollOffset  int
}

func NewMainWindowModel(service *data.Service) MainWindowModel {
	return MainWindowModel{service: service, activeTask: -1, hideCompleted: true}
}

func (m MainWindowModel) LoadTasks() tea.Msg {
	tasksList, err := m.service.TasksRepo.GetAllWithTags()
	if err != nil {
		log.Fatalf("Error retrieving tasks: %v", err)
	}
	if m.hideCompleted {
		filtered := make([]tasks.Task, 0)
		for _, task := range tasksList {
			if !task.State {
				filtered = append(filtered, task)
			}
		}
		tasksList = filtered
	}
	return tasksLoadedMsg{tasks: tasksList}
}

type tasksLoadedMsg struct {
	tasks []tasks.Task
}

type tickMsg time.Time

type dbStateMsg struct {
	state tasks.DBState
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MainWindowModel) checkDBState() tea.Msg {
	state, err := m.service.TasksRepo.GetDBState()
	if err != nil {
		return nil
	}
	return dbStateMsg{state: state}
}

// headerHeight is the number of lines used by the header text
const headerHeight = 4

// taskHeight is the approximate height of each task card (border + content + spacing)
const taskHeight = 4

// View implements tea.Model.
func (m MainWindowModel) View() string {
	style := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).Height(m.height)

	body := "Welcome to taskerr tui mode. Press ctrl+c or q to exit.\n"
	body += "Press h to toggle hide completed tasks.\n"
	body += "Press j/k or down/up to navigate, enter to toggle task state.\n\n"

	// Calculate how many tasks can fit in the visible area
	availableHeight := m.height - headerHeight
	visibleTasks := availableHeight / taskHeight
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	// Determine the range of tasks to display
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleTasks
	if endIdx > len(m.tasks) {
		endIdx = len(m.tasks)
	}

	// Render visible tasks
	for i := startIdx; i < endIdx; i++ {
		active := i == m.activeTask
		body += NewTaskModel(m.tasks[i], m.width, active).View() + "\n"
	}

	// Show scroll indicator if there are more tasks
	if len(m.tasks) > visibleTasks {
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf("  [%d-%d of %d tasks]", startIdx+1, endIdx, len(m.tasks)))
		body += scrollInfo
	}

	return style.Render(body)
}

func (m MainWindowModel) Init() tea.Cmd {
	return tea.Batch(m.LoadTasks, tickCmd())
}

func fixActiveTask(m *MainWindowModel) {
	if m.activeTask >= len(m.tasks) {
		m.activeTask = len(m.tasks) - 1
	} else if m.activeTask < 0 {
		m.activeTask = 0
	}
}

// adjustScroll ensures the active task is visible within the scroll view
func (m *MainWindowModel) adjustScroll() {
	availableHeight := m.height - headerHeight
	visibleTasks := availableHeight / taskHeight
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	// If active task is above the visible area, scroll up
	if m.activeTask < m.scrollOffset {
		m.scrollOffset = m.activeTask
	}

	// If active task is below the visible area, scroll down
	if m.activeTask >= m.scrollOffset+visibleTasks {
		m.scrollOffset = m.activeTask - visibleTasks + 1
	}

	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	// Ensure we don't scroll past the end
	maxOffset := len(m.tasks) - visibleTasks
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

func (m MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		m.tasks = msg.tasks
		if m.activeTask == -1 && len(m.tasks) > 0 {
			m.activeTask = 0
		}
		fixActiveTask(&m)
		m.adjustScroll()
		return m, nil

	case tickMsg:
		return m, m.checkDBState

	case dbStateMsg:
		if msg.state.Count != m.lastDBState.Count ||
			!msg.state.LastUpdated.Equal(m.lastDBState.LastUpdated) {
			m.lastDBState = msg.state
			return m, tea.Batch(m.LoadTasks, tickCmd())
		}
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustScroll()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "j" || msg.String() == "down" {
			m.activeTask++
			fixActiveTask(&m)
			m.adjustScroll()
			return m, nil
		}
		if msg.String() == "k" || msg.String() == "up" {
			m.activeTask--
			fixActiveTask(&m)
			m.adjustScroll()
			return m, nil
		}
		if msg.String() == "enter" || msg.String() == "return" {
			if m.activeTask < 0 || m.activeTask >= len(m.tasks) {
				return m, nil
			}
			task := m.tasks[m.activeTask]
			task.ToggleState()
			m.service.TasksRepo.Update(&task)
			return m, m.LoadTasks
		}
		if msg.String() == "h" {
			m.hideCompleted = !m.hideCompleted
			return m, m.LoadTasks
		}
	}
	return m, nil
}

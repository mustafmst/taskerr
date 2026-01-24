package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// TasksPanelModel represents the tasks panel
type TasksPanelModel struct {
	tasks        []tasks.Task
	activeTask   int
	scrollOffset int
	width        int
	height       int
	focused      bool
}

// NewTasksPanelModel creates a new TasksPanelModel
func NewTasksPanelModel() TasksPanelModel {
	return TasksPanelModel{
		activeTask: -1,
	}
}

// Setters

// SetFocused sets the focus state of the panel
func (m *TasksPanelModel) SetFocused(focused bool) {
	m.focused = focused
}

// SetSize sets the dimensions of the panel
func (m *TasksPanelModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.adjustScroll()
}

// SetTasks sets the tasks and fixes the active task index
func (m *TasksPanelModel) SetTasks(tasksList []tasks.Task) {
	m.tasks = tasksList
	if m.activeTask == -1 && len(m.tasks) > 0 {
		m.activeTask = 0
	}
	m.fixActiveTask()
	m.adjustScroll()
}

// Getters

// IsFocused returns whether the panel is focused
func (m TasksPanelModel) IsFocused() bool {
	return m.focused
}

// ActiveTask returns the currently active task, or nil if none
func (m TasksPanelModel) ActiveTask() *tasks.Task {
	if m.activeTask >= 0 && m.activeTask < len(m.tasks) {
		task := m.tasks[m.activeTask]
		return &task
	}
	return nil
}

// Update handles key events when focused
func (m TasksPanelModel) Update(msg tea.Msg) (TasksPanelModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.activeTask++
			m.fixActiveTask()
			m.adjustScroll()
		case "k", "up":
			m.activeTask--
			m.fixActiveTask()
			m.adjustScroll()
		case " ", "enter": // Space or Enter - toggle task
			if m.activeTask >= 0 && m.activeTask < len(m.tasks) {
				task := m.tasks[m.activeTask]
				task.ToggleState()
				return m, func() tea.Msg {
					return TaskToggledMsg{Task: task}
				}
			}
		}
	}
	return m, nil
}

// View renders the tasks panel
func (m TasksPanelModel) View() string {
	panelStyle := PanelStyle(m.width, m.height, m.focused)

	// Header
	header := HeaderStyle.Render("Tasks")

	// Calculate visible tasks
	availableHeight := m.height - HeaderHeight - PanelPadding
	visibleTasks := availableHeight / TaskHeight
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	// Determine range of tasks to display
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleTasks
	if endIdx > len(m.tasks) {
		endIdx = len(m.tasks)
	}

	// Build content
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	if len(m.tasks) == 0 {
		lines = append(lines, DimStyle.Render("No tasks"))
	} else {
		// Calculate task card width (account for panel border and padding)
		taskCardWidth := m.width - 6

		for i := startIdx; i < endIdx; i++ {
			active := i == m.activeTask && m.focused
			taskView := NewTaskModel(m.tasks[i], taskCardWidth, active).View()
			lines = append(lines, taskView)
		}
	}

	// Scroll indicator
	if len(m.tasks) > visibleTasks {
		scrollInfo := DimStyle.Render(fmt.Sprintf("[%d-%d of %d tasks]", startIdx+1, endIdx, len(m.tasks)))
		lines = append(lines, scrollInfo)
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Render(content)
}

// fixActiveTask ensures activeTask is within bounds
func (m *TasksPanelModel) fixActiveTask() {
	if len(m.tasks) == 0 {
		m.activeTask = -1
		return
	}
	if m.activeTask >= len(m.tasks) {
		m.activeTask = len(m.tasks) - 1
	} else if m.activeTask < 0 {
		m.activeTask = 0
	}
}

// adjustScroll ensures the active task is visible within the scroll view
func (m *TasksPanelModel) adjustScroll() {
	availableHeight := m.height - HeaderHeight - PanelPadding
	visibleTasks := availableHeight / TaskHeight
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	if m.activeTask < m.scrollOffset {
		m.scrollOffset = m.activeTask
	}

	if m.activeTask >= m.scrollOffset+visibleTasks {
		m.scrollOffset = m.activeTask - visibleTasks + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	maxOffset := len(m.tasks) - visibleTasks
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

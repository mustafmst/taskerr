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

	// Calculate visible range using dynamic heights
	startIdx, endIdx := m.calculateVisibleRange()

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
	if endIdx < len(m.tasks) || startIdx > 0 {
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

// calculateVisibleRange determines which tasks fit in the viewport based on their actual heights
func (m TasksPanelModel) calculateVisibleRange() (startIdx, endIdx int) {
	if len(m.tasks) == 0 {
		return 0, 0
	}

	// Available height for task cards
	// Subtract: panel header (2), padding (2), scroll indicator (1)
	availableHeight := m.height - HeaderHeight - PanelPadding - 1
	taskCardWidth := m.width - 6

	startIdx = m.scrollOffset
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(m.tasks) {
		startIdx = len(m.tasks) - 1
	}

	endIdx = startIdx
	usedHeight := 0

	for i := startIdx; i < len(m.tasks); i++ {
		taskHeight := NewTaskModel(m.tasks[i], taskCardWidth, false).Height()
		if usedHeight+taskHeight > availableHeight && i > startIdx {
			break // Don't add this task, but ensure at least 1 is shown
		}
		usedHeight += taskHeight
		endIdx = i + 1
	}

	return startIdx, endIdx
}

// adjustScroll ensures the active task is visible within the scroll view
func (m *TasksPanelModel) adjustScroll() {
	if len(m.tasks) == 0 {
		m.scrollOffset = 0
		return
	}

	// Clamp scrollOffset to valid range
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if m.scrollOffset >= len(m.tasks) {
		m.scrollOffset = len(m.tasks) - 1
	}

	// Ensure active task is not above viewport
	if m.activeTask >= 0 && m.activeTask < m.scrollOffset {
		m.scrollOffset = m.activeTask
	}

	// Ensure active task is not below viewport
	for m.scrollOffset < len(m.tasks) && m.activeTask >= 0 {
		_, endIdx := m.calculateVisibleRange()
		if m.activeTask < endIdx {
			break // Active task is visible
		}
		m.scrollOffset++
	}

	// Clamp scrollOffset again after adjustments
	if m.scrollOffset >= len(m.tasks) {
		m.scrollOffset = len(m.tasks) - 1
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

type TaskModel struct {
	t      tasks.Task
	w      int
	active bool
}

func NewTaskModel(t tasks.Task, w int, active bool) TaskModel {
	return TaskModel{t: t, w: w, active: active}
}

func (t TaskModel) View() string {
	color := "#b3bfbc"
	if t.t.State {
		color = "#16af0e"
	}
	if t.active {
		color = "#e0c021"
	}
	style := lipgloss.NewStyle().
		// Padding(1).
		Width(t.w - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color))

	return style.Render(t.t.Description)
}

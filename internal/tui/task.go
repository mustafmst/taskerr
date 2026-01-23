package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

type TaskModel struct {
	task   tasks.Task
	width  int
	active bool
}

func NewTaskModel(task tasks.Task, width int, active bool) TaskModel {
	return TaskModel{task: task, width: width, active: active}
}

func (model TaskModel) View() string {
	color := "#b3bfbc"
	if model.task.State {
		color = "#16af0e"
	}
	if model.active {
		color = "#e0c021"
	}
	style := lipgloss.NewStyle().
		// Padding(1).
		Width(model.width - 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color))

	return style.Render(model.task.Description)
}

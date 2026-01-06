package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

type TaskModel struct {
	t tasks.Task
}

func NewTaskModel(t tasks.Task) TaskModel {
	return TaskModel{t: t}
}

func (t TaskModel) View() string {
	style := lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5"))
	return style.Render(t.t.Description)
}

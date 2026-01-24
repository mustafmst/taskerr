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

	contentWidth := model.width - 4

	// Build tag badges
	tagStr := ""
	for _, tag := range model.task.Tags {
		tagStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(tag.Color)).
			Padding(0, 1)
		if tagStr != "" {
			tagStr += " "
		}
		tagStr += tagStyle.Render(tag.Name)
	}

	// Build content: description on first line, tags on second line (right-aligned)
	content := model.task.Description

	tagsLineStyle := lipgloss.NewStyle().
		Width(contentWidth - 2). // Account for border padding
		Align(lipgloss.Right)

	if tagStr != "" {
		content += "\n" + tagsLineStyle.Render(tagStr)
	} else {
		noTagsText := lipgloss.NewStyle().
			Foreground(DimTextColor).
			Render("[no tags]")
		content += "\n" + tagsLineStyle.Render(noTagsText)
	}

	style := lipgloss.NewStyle().
		Width(contentWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color))

	return style.Render(content)
}

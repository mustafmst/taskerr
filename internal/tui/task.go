package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// TaskModel represents a single task card in the TUI
type TaskModel struct {
	task   tasks.Task
	width  int
	active bool
}

// NewTaskModel creates a new TaskModel
func NewTaskModel(task tasks.Task, width int, active bool) TaskModel {
	return TaskModel{task: task, width: width, active: active}
}

// Height returns the rendered height of the task card in lines
func (model TaskModel) Height() int {
	contentWidth := model.width - 4
	maxDescWidth := contentWidth - 2

	// Description: max 2 lines (after truncation)
	desc := truncateToLines(model.task.Description, maxDescWidth, 2)
	descLines := countWrappedLines(desc, maxDescWidth)
	if descLines > 2 {
		descLines = 2
	}

	// Tags: always 1 line (with [...] overflow indicator)
	tagsLine := 1

	// Border: 2 lines (top + bottom)
	border := 2

	return descLines + tagsLine + border
}

// View renders the task card
func (model TaskModel) View() string {
	color := "#b3bfbc"
	if model.task.State {
		color = "#16af0e"
	}
	if model.active {
		color = "#e0c021"
	}

	contentWidth := model.width - 4
	maxDescWidth := contentWidth - 2

	// Truncate description to max 2 lines
	desc := truncateToLines(model.task.Description, maxDescWidth, 2)

	// Build tag badges with overflow handling
	tagStr := model.buildTagsLine(maxDescWidth)

	// Build content: description on first line(s), tags on last line (right-aligned)
	tagsLineStyle := lipgloss.NewStyle().
		Width(contentWidth - 2). // Account for border padding
		Align(lipgloss.Right)

	content := desc
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

// buildTagsLine builds the tags string, adding [...] if they overflow
func (model TaskModel) buildTagsLine(maxWidth int) string {
	if len(model.task.Tags) == 0 {
		return ""
	}

	var result string
	moreIndicator := " [...]"
	moreIndicatorWidth := len(moreIndicator)

	for i, tag := range model.task.Tags {
		tagStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(tag.Color)).
			Padding(0, 1)

		rendered := tagStyle.Render(tag.Name)
		renderedWidth := lipgloss.Width(rendered)

		separator := ""
		separatorWidth := 0
		if result != "" {
			separator = " "
			separatorWidth = 1
		}

		currentWidth := lipgloss.Width(result)
		newWidth := currentWidth + separatorWidth + renderedWidth

		// Check if this tag fits (reserve space for [...] if more tags remain)
		hasMoreTags := i < len(model.task.Tags)-1
		reservedSpace := 0
		if hasMoreTags {
			reservedSpace = moreIndicatorWidth
		}

		if newWidth+reservedSpace > maxWidth && i > 0 {
			// Doesn't fit, add indicator
			result += lipgloss.NewStyle().Foreground(DimTextColor).Render(moreIndicator)
			break
		}

		result += separator + rendered
	}

	return result
}

// countWrappedLines returns how many lines text will occupy at given width
func countWrappedLines(text string, width int) int {
	if width <= 0 || len(text) == 0 {
		return 1
	}
	lines := (len(text) + width - 1) / width
	if lines < 1 {
		lines = 1
	}
	return lines
}

// truncateToLines truncates text to fit within maxLines at given width
func truncateToLines(text string, width, maxLines int) string {
	if width <= 0 || maxLines <= 0 {
		return text
	}
	maxChars := width * maxLines
	if len(text) <= maxChars {
		return text
	}
	if maxChars <= 3 {
		return text[:maxChars]
	}
	return text[:maxChars-3] + "..."
}

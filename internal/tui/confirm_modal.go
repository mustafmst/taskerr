package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModalModel represents a confirmation dialog
type ConfirmModalModel struct {
	visible     bool
	title       string
	message     string
	warning     string // Optional warning message (e.g., "Used by 3 tasks")
	width       int
	height      int
	selectedYes bool // false = No (default), true = Yes

	// Context for what we're deleting
	itemType string // "task" or "tag"
	itemID   uint
}

// NewConfirmModalModel creates a new ConfirmModalModel
func NewConfirmModalModel() ConfirmModalModel {
	return ConfirmModalModel{
		selectedYes: false, // Default to No (safer)
	}
}

// Show displays the confirmation modal
func (m *ConfirmModalModel) Show(itemType string, itemID uint, title, message, warning string) {
	m.visible = true
	m.itemType = itemType
	m.itemID = itemID
	m.title = title
	m.message = message
	m.warning = warning
	m.selectedYes = false // Reset to No
}

// SetSize sets the modal dimensions (uses parent window size)
func (m *ConfirmModalModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsVisible returns whether the modal is visible
func (m ConfirmModalModel) IsVisible() bool {
	return m.visible
}

// Reset clears the modal state
func (m *ConfirmModalModel) Reset() {
	m.visible = false
	m.title = ""
	m.message = ""
	m.warning = ""
	m.itemType = ""
	m.itemID = 0
	m.selectedYes = false
}

// Update handles key events
func (m ConfirmModalModel) Update(msg tea.Msg) (ConfirmModalModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "n":
			// Cancel / No
			m.Reset()
			return m, func() tea.Msg { return ModalClosedMsg{} }

		case "y":
			// Yes - confirm deletion
			itemType := m.itemType
			itemID := m.itemID
			m.Reset()
			return m, func() tea.Msg {
				return DeleteConfirmedMsg{ItemType: itemType, ItemID: itemID}
			}

		case "enter":
			if m.selectedYes {
				// Confirm deletion
				itemType := m.itemType
				itemID := m.itemID
				m.Reset()
				return m, func() tea.Msg {
					return DeleteConfirmedMsg{ItemType: itemType, ItemID: itemID}
				}
			} else {
				// Cancel
				m.Reset()
				return m, func() tea.Msg { return ModalClosedMsg{} }
			}

		case "left", "h":
			m.selectedYes = false
		case "right", "l":
			m.selectedYes = true
		case "tab":
			m.selectedYes = !m.selectedYes
		}
	}

	return m, nil
}

// View renders the modal
func (m ConfirmModalModel) View() string {
	if !m.visible {
		return ""
	}

	modalWidth := 50
	if m.width > 0 && m.width < 60 {
		modalWidth = m.width - 10
	}

	// Modal container style
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth-2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ActiveBorderColor).
		Padding(1, 2)

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Render(m.title)

	// Message
	message := lipgloss.NewStyle().
		Foreground(WhiteColor).
		Render(m.message)

	// Warning (if any)
	var warning string
	if m.warning != "" {
		warning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e74c3c")). // Red color for warning
			Render(m.warning)
	}

	// Buttons
	noStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())

	yesStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder())

	if m.selectedYes {
		yesStyle = yesStyle.BorderForeground(ActiveBorderColor).Bold(true)
		noStyle = noStyle.BorderForeground(InactiveBorderColor)
	} else {
		noStyle = noStyle.BorderForeground(ActiveBorderColor).Bold(true)
		yesStyle = yesStyle.BorderForeground(InactiveBorderColor)
	}

	noBtn := noStyle.Render("No")
	yesBtn := yesStyle.Render("Yes")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, noBtn, "  ", yesBtn)
	buttonLine := lipgloss.NewStyle().Width(modalWidth - 6).Align(lipgloss.Center).Render(buttons)

	// Help text
	help := lipgloss.NewStyle().
		Foreground(DimTextColor).
		Render("←/→: select | y/n | Enter | Esc")

	// Compose modal content
	var content string
	if warning != "" {
		content = lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			message,
			warning,
			"",
			buttonLine,
			"",
			help,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Center,
			title,
			"",
			message,
			"",
			buttonLine,
			"",
			help,
		)
	}

	modal := modalStyle.Render(content)

	// Center the modal on screen
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

// Helper function to truncate text if too long
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

// FormatTaskMessage formats the message for task deletion
func FormatTaskMessage(description string) string {
	truncated := truncateText(description, 35)
	return fmt.Sprintf("\"%s\"", truncated)
}

// FormatTagMessage formats the message for tag deletion
func FormatTagMessage(name string) string {
	return fmt.Sprintf("Tag: \"%s\"", name)
}

// FormatTagWarning formats the warning for tag deletion
func FormatTagWarning(taskCount int64) string {
	if taskCount == 0 {
		return ""
	}
	if taskCount == 1 {
		return "Used by 1 task"
	}
	return fmt.Sprintf("Used by %d tasks", taskCount)
}

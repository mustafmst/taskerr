package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// ModalField represents which field is focused in the modal
type ModalField int

const (
	FieldDescription ModalField = iota
	FieldTags
	FieldNewTags
)

// AddTaskModalModel represents the add task modal dialog
type AddTaskModalModel struct {
	visible     bool
	width       int
	height      int
	activeField ModalField

	// Fields
	descInput   TextInputModel
	newTagInput TextInputModel

	// Tags selection
	tags         []tasks.Tag
	selectedTags map[uint]bool
	activeTagIdx int
	tagScroll    int

	// Edit mode
	editMode   bool
	editTaskID uint
}

// NewAddTaskModalModel creates a new AddTaskModalModel
func NewAddTaskModalModel() AddTaskModalModel {
	return AddTaskModalModel{
		descInput:    NewTextInputModel("Task description..."),
		newTagInput:  NewTextInputModel("New tags (comma-separated)..."),
		selectedTags: make(map[uint]bool),
	}
}

// Setters

// SetVisible shows or hides the modal
func (m *AddTaskModalModel) SetVisible(visible bool) {
	m.visible = visible
	if visible {
		m.activeField = FieldDescription
		m.descInput.SetFocused(true)
		m.newTagInput.SetFocused(false)
	}
}

// SetSize sets the modal dimensions (uses parent window size)
func (m *AddTaskModalModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Set input widths (account for modal border, padding, and input border/padding)
	modalWidth := m.modalWidth()
	inputWidth := modalWidth - 10
	m.descInput.SetWidth(inputWidth)
	m.newTagInput.SetWidth(inputWidth)
}

// SetTags sets the available tags for selection
func (m *AddTaskModalModel) SetTags(tags []tasks.Tag) {
	m.tags = tags
}

// SetEditTask populates the modal with existing task data for editing
func (m *AddTaskModalModel) SetEditTask(task *tasks.Task) {
	m.editMode = true
	m.editTaskID = task.ID
	m.descInput.SetValue(task.Description)

	// Pre-select task's existing tags
	m.selectedTags = make(map[uint]bool)
	for _, tag := range task.Tags {
		m.selectedTags[tag.ID] = true
	}
}

// Getters

// IsVisible returns whether the modal is visible
func (m AddTaskModalModel) IsVisible() bool {
	return m.visible
}

// Reset clears the modal state
func (m *AddTaskModalModel) Reset() {
	m.descInput.Reset()
	m.newTagInput.Reset()
	m.selectedTags = make(map[uint]bool)
	m.activeTagIdx = 0
	m.tagScroll = 0
	m.activeField = FieldDescription
	m.editMode = false
	m.editTaskID = 0
}

// modalWidth returns the width of the modal
func (m AddTaskModalModel) modalWidth() int {
	w := m.width * 2 / 3
	if w < 50 {
		w = 50
	}
	if w > 80 {
		w = 80
	}
	return w
}

// modalHeight returns the height of the modal
func (m AddTaskModalModel) modalHeight() int {
	h := m.height * 2 / 3
	if h < 20 {
		h = 20
	}
	if h > 30 {
		h = 30
	}
	return h
}

// Update handles key events
func (m AddTaskModalModel) Update(msg tea.Msg) (AddTaskModalModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Reset()
			m.visible = false
			return m, func() tea.Msg { return ModalClosedMsg{} }

		case "enter":
			// Save task if description is not empty
			desc := strings.TrimSpace(m.descInput.Value())
			if desc == "" {
				return m, nil
			}

			// Collect selected tag IDs
			var tagIDs []uint
			for tagID, selected := range m.selectedTags {
				if selected {
					tagIDs = append(tagIDs, tagID)
				}
			}

			// Parse new tag names
			var newTagNames []string
			newTagsStr := strings.TrimSpace(m.newTagInput.Value())
			if newTagsStr != "" {
				parts := strings.Split(newTagsStr, ",")
				for _, part := range parts {
					name := strings.TrimSpace(part)
					if name != "" {
						newTagNames = append(newTagNames, name)
					}
				}
			}

			// Capture edit mode state before reset
			editMode := m.editMode
			editTaskID := m.editTaskID

			m.Reset()
			m.visible = false

			if editMode {
				return m, func() tea.Msg {
					return TaskUpdatedMsg{
						TaskID:      editTaskID,
						Description: desc,
						TagIDs:      tagIDs,
						NewTagNames: newTagNames,
					}
				}
			}

			return m, func() tea.Msg {
				return TaskCreatedMsg{
					Description: desc,
					TagIDs:      tagIDs,
					NewTagNames: newTagNames,
				}
			}

		case "tab":
			m.nextField()
			return m, nil

		case "shift+tab":
			m.prevField()
			return m, nil

		default:
			// Delegate to active field
			switch m.activeField {
			case FieldDescription:
				var cmd tea.Cmd
				m.descInput, cmd = m.descInput.Update(msg)
				return m, cmd

			case FieldTags:
				m.handleTagsInput(msg)
				return m, nil

			case FieldNewTags:
				var cmd tea.Cmd
				m.newTagInput, cmd = m.newTagInput.Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

// nextField moves to the next field
func (m *AddTaskModalModel) nextField() {
	m.descInput.SetFocused(false)
	m.newTagInput.SetFocused(false)

	switch m.activeField {
	case FieldDescription:
		m.activeField = FieldTags
	case FieldTags:
		m.activeField = FieldNewTags
		m.newTagInput.SetFocused(true)
	case FieldNewTags:
		m.activeField = FieldDescription
		m.descInput.SetFocused(true)
	}
}

// prevField moves to the previous field
func (m *AddTaskModalModel) prevField() {
	m.descInput.SetFocused(false)
	m.newTagInput.SetFocused(false)

	switch m.activeField {
	case FieldDescription:
		m.activeField = FieldNewTags
		m.newTagInput.SetFocused(true)
	case FieldTags:
		m.activeField = FieldDescription
		m.descInput.SetFocused(true)
	case FieldNewTags:
		m.activeField = FieldTags
	}
}

// handleTagsInput handles key events for the tags list
func (m *AddTaskModalModel) handleTagsInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "j", "down":
		if m.activeTagIdx < len(m.tags)-1 {
			m.activeTagIdx++
			m.adjustTagScroll()
		}
	case "k", "up":
		if m.activeTagIdx > 0 {
			m.activeTagIdx--
			m.adjustTagScroll()
		}
	case " ":
		if m.activeTagIdx >= 0 && m.activeTagIdx < len(m.tags) {
			tagID := m.tags[m.activeTagIdx].ID
			m.selectedTags[tagID] = !m.selectedTags[tagID]
			if !m.selectedTags[tagID] {
				delete(m.selectedTags, tagID)
			}
		}
	}
}

// adjustTagScroll ensures the active tag is visible
func (m *AddTaskModalModel) adjustTagScroll() {
	visibleTags := m.visibleTagCount()

	if m.activeTagIdx < m.tagScroll {
		m.tagScroll = m.activeTagIdx
	}

	if m.activeTagIdx >= m.tagScroll+visibleTags {
		m.tagScroll = m.activeTagIdx - visibleTags + 1
	}

	if m.tagScroll < 0 {
		m.tagScroll = 0
	}
}

// visibleTagCount returns how many tags can be displayed
func (m AddTaskModalModel) visibleTagCount() int {
	// Reserve space for header, inputs, labels, and padding
	available := m.modalHeight() - 12
	if available < 3 {
		available = 3
	}
	return available
}

// View renders the modal
func (m AddTaskModalModel) View() string {
	if !m.visible {
		return ""
	}

	modalWidth := m.modalWidth()
	modalHeight := m.modalHeight()

	// Modal container style
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth-2).
		Height(modalHeight-2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ActiveBorderColor).
		Padding(1, 2)

	// Title - changes based on mode
	titleText := "Add New Task"
	if m.editMode {
		titleText = "Edit Task"
	}
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Render(titleText)

	// Description field
	descLabel := m.fieldLabel("Description:", m.activeField == FieldDescription)
	descInput := m.descInput.View()

	// Tags field
	tagsLabel := m.fieldLabel("Tags:", m.activeField == FieldTags)
	tagsView := m.renderTagsList()

	// New tags field
	newTagsLabel := m.fieldLabel("New Tags:", m.activeField == FieldNewTags)
	newTagsInput := m.newTagInput.View()

	// Help
	help := lipgloss.NewStyle().
		Foreground(DimTextColor).
		Render("TAB: next field | SPACE: toggle tag | Enter: save | Esc: cancel")

	// Compose modal content
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		descLabel,
		descInput,
		"",
		tagsLabel,
		tagsView,
		"",
		newTagsLabel,
		newTagsInput,
		"",
		help,
	)

	modal := modalStyle.Render(content)

	// Center on screen
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

// fieldLabel renders a field label with focus indicator
func (m AddTaskModalModel) fieldLabel(label string, focused bool) string {
	style := lipgloss.NewStyle()
	if focused {
		style = style.Bold(true).Foreground(ActiveBorderColor)
		return style.Render("> " + label)
	}
	return style.Foreground(DimTextColor).Render("  " + label)
}

// renderTagsList renders the tags checklist
func (m AddTaskModalModel) renderTagsList() string {
	if len(m.tags) == 0 {
		return DimStyle.Render("  No tags available")
	}

	visibleTags := m.visibleTagCount()
	startIdx := m.tagScroll
	endIdx := startIdx + visibleTags
	if endIdx > len(m.tags) {
		endIdx = len(m.tags)
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		tag := m.tags[i]

		// Checkbox
		checkbox := "[ ]"
		if m.selectedTags[tag.ID] {
			checkbox = "[x]"
		}

		// Tag color dot
		colorDot := lipgloss.NewStyle().
			Foreground(lipgloss.Color(tag.Color)).
			Render("●")

		line := fmt.Sprintf("  %s %s %s", checkbox, colorDot, tag.Name)

		// Highlight if active and tags field is focused
		if i == m.activeTagIdx && m.activeField == FieldTags {
			line = lipgloss.NewStyle().
				Background(HighlightBgColor).
				Foreground(WhiteColor).
				Render(line)
		}

		lines = append(lines, line)
	}

	// Scroll indicator
	if len(m.tags) > visibleTags {
		scrollInfo := DimStyle.Render(fmt.Sprintf("  [%d-%d/%d]", startIdx+1, endIdx, len(m.tags)))
		lines = append(lines, scrollInfo)
	}

	return strings.Join(lines, "\n")
}

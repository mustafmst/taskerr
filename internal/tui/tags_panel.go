package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// TagsPanelModel represents the tags panel with filtering functionality
type TagsPanelModel struct {
	tags         []tasks.Tag
	selectedTags map[uint]bool
	activeTag    int
	scrollOffset int
	width        int
	height       int
	focused      bool
	filterMode   FilterMode
}

// NewTagsPanelModel creates a new TagsPanelModel
func NewTagsPanelModel() TagsPanelModel {
	return TagsPanelModel{
		selectedTags: make(map[uint]bool),
		activeTag:    -1,
		filterMode:   FilterOR,
	}
}

// Setters

// SetFocused sets the focus state of the panel
func (m *TagsPanelModel) SetFocused(focused bool) {
	m.focused = focused
}

// SetSize sets the dimensions of the panel
func (m *TagsPanelModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.adjustScroll()
}

// SetTags sets the tags and fixes the active tag index
func (m *TagsPanelModel) SetTags(tags []tasks.Tag) {
	m.tags = tags
	if m.activeTag == -1 && len(m.tags) > 0 {
		m.activeTag = 0
	}
	m.fixActiveTag()
	m.adjustScroll()
}

// SetFilterMode sets the filter mode
func (m *TagsPanelModel) SetFilterMode(mode FilterMode) {
	m.filterMode = mode
}

// Getters

// SelectedTags returns the map of selected tag IDs
func (m TagsPanelModel) SelectedTags() map[uint]bool {
	return m.selectedTags
}

// FilterMode returns the current filter mode
func (m TagsPanelModel) GetFilterMode() FilterMode {
	return m.filterMode
}

// Tags returns the list of tags
func (m TagsPanelModel) Tags() []tasks.Tag {
	return m.tags
}

// IsFocused returns whether the panel is focused
func (m TagsPanelModel) IsFocused() bool {
	return m.focused
}

// Update handles key events when focused
func (m TagsPanelModel) Update(msg tea.Msg) (TagsPanelModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.activeTag++
			m.fixActiveTag()
			m.adjustScroll()
		case "k", "up":
			m.activeTag--
			m.fixActiveTag()
			m.adjustScroll()
		case " ": // Space - toggle tag
			if m.activeTag >= 0 && m.activeTag < len(m.tags) {
				tagID := m.tags[m.activeTag].ID
				m.selectedTags[tagID] = !m.selectedTags[tagID]
				if !m.selectedTags[tagID] {
					delete(m.selectedTags, tagID)
				}
				selected := m.selectedTags[tagID]
				return m, func() tea.Msg {
					return TagToggledMsg{TagID: tagID, Selected: selected}
				}
			}
		case "m": // Toggle filter mode
			if m.filterMode == FilterOR {
				m.filterMode = FilterAND
			} else {
				m.filterMode = FilterOR
			}
			mode := m.filterMode
			return m, func() tea.Msg {
				return FilterModeChangedMsg{Mode: mode}
			}
		}
	}
	return m, nil
}

// View renders the tags panel
func (m TagsPanelModel) View() string {
	panelStyle := PanelStyle(m.width, m.height, m.focused)

	// Header
	header := HeaderStyle.Render(fmt.Sprintf("Tags [%s]", m.filterMode.String()))

	// Calculate visible tags
	availableHeight := m.height - HeaderHeight - PanelPadding
	visibleTags := availableHeight / TagHeight
	if visibleTags < 1 {
		visibleTags = 1
	}

	// Determine range of tags to display
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleTags
	if endIdx > len(m.tags) {
		endIdx = len(m.tags)
	}

	// Render tags
	var tagLines []string
	tagLines = append(tagLines, header)
	tagLines = append(tagLines, "")

	if len(m.tags) == 0 {
		tagLines = append(tagLines, DimStyle.Render("No tags"))
	} else {
		for i := startIdx; i < endIdx; i++ {
			tag := m.tags[i]

			// Checkbox
			checkbox := "[ ]"
			if m.selectedTags[tag.ID] {
				checkbox = "[x]"
			}

			// Tag name with color indicator
			colorDot := lipgloss.NewStyle().
				Foreground(lipgloss.Color(tag.Color)).
				Render("●")

			tagName := tag.Name

			// Build line
			line := fmt.Sprintf("%s %s %s", checkbox, colorDot, tagName)

			// Highlight if active and focused
			if i == m.activeTag && m.focused {
				line = lipgloss.NewStyle().
					Background(HighlightBgColor).
					Foreground(WhiteColor).
					Render(line)
			}

			tagLines = append(tagLines, line)
		}
	}

	// Scroll indicator
	if len(m.tags) > visibleTags {
		scrollInfo := DimStyle.Render(fmt.Sprintf("[%d-%d/%d]", startIdx+1, endIdx, len(m.tags)))
		tagLines = append(tagLines, "")
		tagLines = append(tagLines, scrollInfo)
	}

	content := strings.Join(tagLines, "\n")
	return panelStyle.Render(content)
}

// fixActiveTag ensures activeTag is within bounds
func (m *TagsPanelModel) fixActiveTag() {
	if len(m.tags) == 0 {
		m.activeTag = -1
		return
	}
	if m.activeTag >= len(m.tags) {
		m.activeTag = len(m.tags) - 1
	} else if m.activeTag < 0 {
		m.activeTag = 0
	}
}

// adjustScroll ensures the active tag is visible within the scroll view
func (m *TagsPanelModel) adjustScroll() {
	availableHeight := m.height - HeaderHeight - PanelPadding
	visibleTags := availableHeight / TagHeight
	if visibleTags < 1 {
		visibleTags = 1
	}

	if m.activeTag < m.scrollOffset {
		m.scrollOffset = m.activeTag
	}

	if m.activeTag >= m.scrollOffset+visibleTags {
		m.scrollOffset = m.activeTag - visibleTags + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	maxOffset := len(m.tags) - visibleTags
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

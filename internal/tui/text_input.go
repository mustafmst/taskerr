package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInputModel is a simple text input component
type TextInputModel struct {
	value       string
	placeholder string
	focused     bool
	width       int
	cursorPos   int
}

// NewTextInputModel creates a new TextInputModel
func NewTextInputModel(placeholder string) TextInputModel {
	return TextInputModel{
		placeholder: placeholder,
	}
}

// Setters

// SetFocused sets the focus state
func (m *TextInputModel) SetFocused(focused bool) {
	m.focused = focused
}

// SetWidth sets the display width
func (m *TextInputModel) SetWidth(width int) {
	m.width = width
}

// SetValue sets the input value
func (m *TextInputModel) SetValue(value string) {
	m.value = value
	m.cursorPos = len(value)
}

// Getters

// Value returns the current input value
func (m TextInputModel) Value() string {
	return m.value
}

// IsFocused returns whether the input is focused
func (m TextInputModel) IsFocused() bool {
	return m.focused
}

// Reset clears the input value
func (m *TextInputModel) Reset() {
	m.value = ""
	m.cursorPos = 0
}

// Update handles key events
func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			if len(m.value) > 0 && m.cursorPos > 0 {
				m.value = m.value[:m.cursorPos-1] + m.value[m.cursorPos:]
				m.cursorPos--
			}
		case tea.KeyDelete:
			if m.cursorPos < len(m.value) {
				m.value = m.value[:m.cursorPos] + m.value[m.cursorPos+1:]
			}
		case tea.KeyLeft:
			if m.cursorPos > 0 {
				m.cursorPos--
			}
		case tea.KeyRight:
			if m.cursorPos < len(m.value) {
				m.cursorPos++
			}
		case tea.KeyHome:
			m.cursorPos = 0
		case tea.KeyEnd:
			m.cursorPos = len(m.value)
		case tea.KeyRunes:
			// Insert character at cursor position
			char := string(msg.Runes)
			m.value = m.value[:m.cursorPos] + char + m.value[m.cursorPos:]
			m.cursorPos += len(char)
		}
	}

	return m, nil
}

// View renders the text input
func (m TextInputModel) View() string {
	width := m.width
	if width == 0 {
		width = 30
	}

	var display string
	if m.value == "" && !m.focused {
		// Show placeholder
		display = lipgloss.NewStyle().
			Foreground(DimTextColor).
			Render(m.placeholder)
	} else if m.focused {
		// Show value with cursor
		before := m.value[:m.cursorPos]
		after := m.value[m.cursorPos:]
		cursor := lipgloss.NewStyle().
			Background(WhiteColor).
			Foreground(lipgloss.Color("#000000")).
			Render(" ")
		display = before + cursor + after
	} else {
		display = m.value
	}

	// Pad or truncate to width
	displayLen := len(m.value)
	if m.focused {
		displayLen++ // for cursor
	}
	if displayLen < width-2 {
		display += strings.Repeat(" ", width-2-displayLen)
	}

	// Style based on focus
	style := lipgloss.NewStyle().
		Width(width-2).
		Padding(0, 1)

	if m.focused {
		style = style.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ActiveBorderColor)
	} else {
		style = style.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(InactiveBorderColor)
	}

	return style.Render(display)
}

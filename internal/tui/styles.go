package tui

import "github.com/charmbracelet/lipgloss"

// Layout constants
const (
	HeaderBarHeight = 1
	HeaderHeight    = 2 // Panel header height (title + blank line)
	FooterHeight    = 1
	TaskHeight      = 4
	TagHeight       = 1
	PanelPadding    = 2
)

// Colors
var (
	ActiveBorderColor   = lipgloss.Color("#e0c021")
	InactiveBorderColor = lipgloss.Color("#555555")
	SelectedTagColor    = lipgloss.Color("#2ecc71")
	DimTextColor        = lipgloss.Color("#888888")
	WhiteColor          = lipgloss.Color("#ffffff")
	HighlightBgColor    = lipgloss.Color("#444444")
	HeaderBarBgColor    = lipgloss.Color("#e0c021")
	HeaderBarFgColor    = lipgloss.Color("#000000")
)

// FilterMode represents how multiple selected tags filter tasks
type FilterMode int

const (
	FilterOR FilterMode = iota
	FilterAND
)

func (f FilterMode) String() string {
	if f == FilterOR {
		return "OR"
	}
	return "AND"
}

// Common styles
var (
	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(WhiteColor)
	DimStyle    = lipgloss.NewStyle().Foreground(DimTextColor)
)

// PanelStyle returns a style for a panel with the given dimensions and focus state
func PanelStyle(width, height int, focused bool) lipgloss.Style {
	borderColor := InactiveBorderColor
	if focused {
		borderColor = ActiveBorderColor
	}
	return lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)
}

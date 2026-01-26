package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// StatsModalModel represents the statistics modal
type StatsModalModel struct {
	visible bool
	width   int
	height  int

	// Task overview
	totalTasks      int
	completedTasks  int
	incompleteTasks int
	completionRate  float64

	// Time-based metrics
	todayCompleted    int
	weekCompleted     int
	monthCompleted    int
	avgCompletionTime time.Duration

	// Tag statistics
	totalTags int
	tagStats  []tasks.TagStat // sorted by count descending

	// Activity trend (last 7 days)
	dailyActivity [7]int
	dayLabels     [7]string
}

// NewStatsModalModel creates a new stats modal
func NewStatsModalModel() StatsModalModel {
	return StatsModalModel{}
}

// Show displays the modal with pre-computed statistics from the database
func (m *StatsModalModel) Show(stats *tasks.TaskStats) {
	m.visible = true
	m.applyStats(stats)
}

// applyStats applies the database-computed statistics to the modal
func (m *StatsModalModel) applyStats(stats *tasks.TaskStats) {
	m.totalTasks = int(stats.TotalTasks)
	m.completedTasks = int(stats.CompletedTasks)
	m.incompleteTasks = int(stats.IncompleteTasks)
	m.completionRate = stats.CompletionRate

	m.todayCompleted = int(stats.TodayCompleted)
	m.weekCompleted = int(stats.WeekCompleted)
	m.monthCompleted = int(stats.MonthCompleted)
	m.avgCompletionTime = stats.AvgCompletionTime

	// Copy daily activity
	for i := 0; i < 7; i++ {
		m.dailyActivity[i] = int(stats.DailyActivity[i])
		m.dayLabels[i] = stats.DayLabels[i]
	}

	m.totalTags = stats.TotalTags
	m.tagStats = stats.TagStats
}

// SetSize sets the modal dimensions
func (m *StatsModalModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsVisible returns whether the modal is visible
func (m StatsModalModel) IsVisible() bool {
	return m.visible
}

// Reset clears the modal state
func (m *StatsModalModel) Reset() {
	m.visible = false
}

// Update handles key events
func (m StatsModalModel) Update(msg tea.Msg) (StatsModalModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "s", "q":
			m.Reset()
			return m, func() tea.Msg { return ModalClosedMsg{} }
		}
	}
	return m, nil
}

// View renders the modal
func (m StatsModalModel) View() string {
	if !m.visible {
		return ""
	}

	// Calculate modal dimensions (adapt to terminal size)
	modalWidth := m.width - 10
	if modalWidth > 70 {
		modalWidth = 70
	}
	if modalWidth < 50 {
		modalWidth = 50
	}

	contentWidth := modalWidth - 6 // Account for borders and padding

	// Build sections
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ActiveBorderColor).
		Width(contentWidth).
		Align(lipgloss.Center).
		Render("Statistics")

	sections = append(sections, title, "")

	// Section 1: Task Overview
	sections = append(sections, m.renderSectionHeader("TASK OVERVIEW"))
	sections = append(sections, fmt.Sprintf("Total: %d    Completed: %d    Incomplete: %d",
		m.totalTasks, m.completedTasks, m.incompleteTasks))
	sections = append(sections, fmt.Sprintf("Completion: %s",
		renderPercentBar(m.completionRate, contentWidth-12)))
	sections = append(sections, "")

	// Section 2: Time Metrics
	sections = append(sections, m.renderSectionHeader("TIME METRICS"))
	sections = append(sections, fmt.Sprintf("Today: %d    This Week: %d    This Month: %d",
		m.todayCompleted, m.weekCompleted, m.monthCompleted))
	sections = append(sections, fmt.Sprintf("Avg Time to Complete: %s",
		formatDuration(m.avgCompletionTime)))
	sections = append(sections, "")

	// Section 3: Top Tags
	sections = append(sections, m.renderSectionHeader("TOP TAGS"))
	maxTags := 5
	if len(m.tagStats) < maxTags {
		maxTags = len(m.tagStats)
	}
	if maxTags == 0 {
		sections = append(sections, DimStyle.Render("No tags yet"))
	} else {
		maxCount := int64(0)
		for i := 0; i < maxTags; i++ {
			if m.tagStats[i].Count > maxCount {
				maxCount = m.tagStats[i].Count
			}
		}
		for i := 0; i < maxTags; i++ {
			tag := m.tagStats[i]
			sections = append(sections, renderTagBar(tag.Name, int(tag.Count), int(maxCount), contentWidth))
		}
	}
	sections = append(sections, "")

	// Section 4: Activity Trend
	sections = append(sections, m.renderSectionHeader("ACTIVITY (Last 7 Days)"))
	sections = append(sections, renderActivityChart(m.dailyActivity, m.dayLabels))
	sections = append(sections, "")

	// Help
	help := lipgloss.NewStyle().
		Foreground(DimTextColor).
		Width(contentWidth).
		Align(lipgloss.Center).
		Render("Press Esc or s to close")
	sections = append(sections, help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Modal container
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth-2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ActiveBorderColor).
		Padding(1, 2)

	modal := modalStyle.Render(content)

	// Center on screen
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

func (m StatsModalModel) renderSectionHeader(title string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(WhiteColor).
		Render(title)
}

// renderPercentBar renders a percentage bar
func renderPercentBar(percent float64, width int) string {
	barWidth := width - 5 // Leave room for "XX%"
	if barWidth < 10 {
		barWidth = 10
	}
	filled := int(percent / 100 * float64(barWidth))
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("%s %3.0f%%", bar, percent)
}

// renderTagBar renders a tag with its usage bar
func renderTagBar(name string, count, maxCount, width int) string {
	nameWidth := 12
	if len(name) > nameWidth {
		name = name[:nameWidth-1] + "…"
	}
	name = fmt.Sprintf("%-*s", nameWidth, name)

	barWidth := width - nameWidth - 6 // Room for name and count
	if barWidth < 5 {
		barWidth = 5
	}

	filled := 0
	if maxCount > 0 {
		filled = int(float64(count) / float64(maxCount) * float64(barWidth))
		if filled == 0 && count > 0 {
			filled = 1 // Show at least 1 block if count > 0
		}
	}

	bar := strings.Repeat("█", filled)
	return fmt.Sprintf("%s %s %d", name,
		lipgloss.NewStyle().Foreground(SelectedTagColor).Render(bar), count)
}

// renderActivityChart renders the 7-day activity chart
func renderActivityChart(daily [7]int, labels [7]string) string {
	blocks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	// Find max for scaling
	maxVal := 0
	for _, v := range daily {
		if v > maxVal {
			maxVal = v
		}
	}

	var bars []string
	var counts []string
	var dayLabels []string

	for i := 0; i < 7; i++ {
		level := 0
		if maxVal > 0 {
			level = int(float64(daily[i]) / float64(maxVal) * 8)
			if level > 8 {
				level = 8
			}
		}
		bars = append(bars, fmt.Sprintf(" %s ", blocks[level]))
		counts = append(counts, fmt.Sprintf("%3d", daily[i]))
		dayLabels = append(dayLabels, labels[i])
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(dayLabels, " "),
		strings.Join(bars, " "),
		strings.Join(counts, " "),
	)
}

// formatDuration formats a duration as human readable
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

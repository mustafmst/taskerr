package tui

import (
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// Panel represents which panel is currently active
type Panel int

const (
	TagsPanel Panel = iota
	TasksPanel
)

// FilterMode represents how multiple selected tags filter tasks
type FilterMode int

const (
	FilterOR FilterMode = iota
	FilterAND
)

// Layout constants
const (
	headerHeight = 2
	footerHeight = 1
	taskHeight   = 4
	tagHeight    = 1
	panelPadding = 2
)

// Colors
var (
	activeBorderColor   = lipgloss.Color("#e0c021")
	inactiveBorderColor = lipgloss.Color("#555555")
	selectedTagColor    = lipgloss.Color("#2ecc71")
	dimTextColor        = lipgloss.Color("#888888")
)

type MainWindowModel struct {
	width         int
	height        int
	tasks         []tasks.Task
	service       *data.Service
	activeTask    int
	hideCompleted bool
	lastDBState   tasks.DBState
	scrollOffset  int

	// Panel-related fields
	activePanel     Panel
	tags            []tasks.Tag
	selectedTags    map[uint]bool
	activeTag       int
	tagScrollOffset int
	filterMode      FilterMode
}

func NewMainWindowModel(service *data.Service) MainWindowModel {
	return MainWindowModel{
		service:       service,
		activeTask:    -1,
		activeTag:     -1,
		hideCompleted: true,
		activePanel:   TasksPanel,
		selectedTags:  make(map[uint]bool),
		filterMode:    FilterOR,
	}
}

// Message types
type tasksLoadedMsg struct {
	tasks []tasks.Task
}

type tagsLoadedMsg struct {
	tags []tasks.Tag
}

type tickMsg time.Time

type dbStateMsg struct {
	state tasks.DBState
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MainWindowModel) checkDBState() tea.Msg {
	state, err := m.service.TasksRepo.GetDBState()
	if err != nil {
		return nil
	}
	return dbStateMsg{state: state}
}

func (m MainWindowModel) LoadTasks() tea.Msg {
	tasksList, err := m.service.TasksRepo.GetAllWithTags()
	if err != nil {
		log.Fatalf("Error retrieving tasks: %v", err)
	}

	// Filter by completion state
	if m.hideCompleted {
		filtered := make([]tasks.Task, 0)
		for _, task := range tasksList {
			if !task.State {
				filtered = append(filtered, task)
			}
		}
		tasksList = filtered
	}

	// Filter by selected tags
	if len(m.selectedTags) > 0 {
		tasksList = m.filterByTags(tasksList)
	}

	return tasksLoadedMsg{tasks: tasksList}
}

func (m MainWindowModel) LoadTags() tea.Msg {
	tagsList, err := m.service.TagsRepo.GetAll()
	if err != nil {
		log.Fatalf("Error retrieving tags: %v", err)
	}
	return tagsLoadedMsg{tags: tagsList}
}

// filterByTags filters tasks based on selected tags and filter mode
func (m MainWindowModel) filterByTags(tasksList []tasks.Task) []tasks.Task {
	filtered := make([]tasks.Task, 0)

	for _, task := range tasksList {
		if m.filterMode == FilterOR {
			// OR mode: task must have at least one selected tag
			for _, tag := range task.Tags {
				if m.selectedTags[tag.ID] {
					filtered = append(filtered, task)
					break
				}
			}
		} else {
			// AND mode: task must have all selected tags
			hasAll := true
			for tagID := range m.selectedTags {
				if !m.selectedTags[tagID] {
					continue
				}
				found := false
				for _, tag := range task.Tags {
					if tag.ID == tagID {
						found = true
						break
					}
				}
				if !found {
					hasAll = false
					break
				}
			}
			if hasAll {
				filtered = append(filtered, task)
			}
		}
	}

	return filtered
}

func (m MainWindowModel) filterModeString() string {
	if m.filterMode == FilterOR {
		return "OR"
	}
	return "AND"
}

func (m MainWindowModel) Init() tea.Cmd {
	return tea.Batch(m.LoadTasks, m.LoadTags, tickCmd())
}

// View implements tea.Model.
func (m MainWindowModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate panel widths (25% for tags, 75% for tasks)
	tagsWidth := m.width / 4
	if tagsWidth < 15 {
		tagsWidth = 15
	}
	tasksWidth := m.width - tagsWidth

	// Calculate content height
	contentHeight := m.height - footerHeight

	// Render panels
	tagsPanel := m.renderTagsPanel(tagsWidth, contentHeight)
	tasksPanel := m.renderTasksPanel(tasksWidth, contentHeight)

	// Join panels horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, tagsPanel, tasksPanel)

	// Render footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

func (m MainWindowModel) renderTagsPanel(width, height int) string {
	// Determine border color based on active panel
	borderColor := inactiveBorderColor
	if m.activePanel == TagsPanel {
		borderColor = activeBorderColor
	}

	// Create panel style
	panelStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff"))
	header := headerStyle.Render(fmt.Sprintf("Tags [%s]", m.filterModeString()))

	// Calculate visible tags
	availableHeight := height - headerHeight - panelPadding
	visibleTags := availableHeight / tagHeight
	if visibleTags < 1 {
		visibleTags = 1
	}

	// Determine range of tags to display
	startIdx := m.tagScrollOffset
	endIdx := startIdx + visibleTags
	if endIdx > len(m.tags) {
		endIdx = len(m.tags)
	}

	// Render tags
	var tagLines []string
	tagLines = append(tagLines, header)
	tagLines = append(tagLines, "")

	if len(m.tags) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(dimTextColor)
		tagLines = append(tagLines, dimStyle.Render("No tags"))
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

			// Highlight if active
			line := fmt.Sprintf("%s %s %s", checkbox, colorDot, tagName)
			if i == m.activeTag && m.activePanel == TagsPanel {
				line = lipgloss.NewStyle().
					Background(lipgloss.Color("#444444")).
					Foreground(lipgloss.Color("#ffffff")).
					Render(line)
			}

			tagLines = append(tagLines, line)
		}
	}

	// Scroll indicator
	if len(m.tags) > visibleTags {
		scrollInfo := lipgloss.NewStyle().
			Foreground(dimTextColor).
			Render(fmt.Sprintf("[%d-%d/%d]", startIdx+1, endIdx, len(m.tags)))
		tagLines = append(tagLines, "")
		tagLines = append(tagLines, scrollInfo)
	}

	content := strings.Join(tagLines, "\n")
	return panelStyle.Render(content)
}

func (m MainWindowModel) renderTasksPanel(width, height int) string {
	// Determine border color based on active panel
	borderColor := inactiveBorderColor
	if m.activePanel == TasksPanel {
		borderColor = activeBorderColor
	}

	// Create panel style
	panelStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff"))
	header := headerStyle.Render("Tasks")

	// Calculate visible tasks
	availableHeight := height - headerHeight - panelPadding
	visibleTasks := availableHeight / taskHeight
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	// Determine range of tasks to display
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleTasks
	if endIdx > len(m.tasks) {
		endIdx = len(m.tasks)
	}

	// Build content
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	if len(m.tasks) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(dimTextColor)
		lines = append(lines, dimStyle.Render("No tasks"))
	} else {
		// Calculate task card width (account for panel border and padding)
		taskCardWidth := width - 6

		for i := startIdx; i < endIdx; i++ {
			active := i == m.activeTask && m.activePanel == TasksPanel
			taskView := NewTaskModel(m.tasks[i], taskCardWidth, active).View()
			lines = append(lines, taskView)
		}
	}

	// Scroll indicator
	if len(m.tasks) > visibleTasks {
		scrollInfo := lipgloss.NewStyle().
			Foreground(dimTextColor).
			Render(fmt.Sprintf("[%d-%d of %d tasks]", startIdx+1, endIdx, len(m.tasks)))
		lines = append(lines, scrollInfo)
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Render(content)
}

func (m MainWindowModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(dimTextColor).
		Width(m.width)

	hiddenStatus := "shown"
	if m.hideCompleted {
		hiddenStatus = "hidden"
	}

	footer := fmt.Sprintf(" TAB: switch panel | SPACE: toggle | m: filter mode (%s) | h: completed %s | q: quit",
		m.filterModeString(), hiddenStatus)

	return footerStyle.Render(footer)
}

func fixActiveTask(m *MainWindowModel) {
	if len(m.tasks) == 0 {
		m.activeTask = -1
		return
	}
	if m.activeTask >= len(m.tasks) {
		m.activeTask = len(m.tasks) - 1
	} else if m.activeTask < 0 {
		m.activeTask = 0
	}
}

func fixActiveTag(m *MainWindowModel) {
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

// adjustScroll ensures the active task is visible within the scroll view
func (m *MainWindowModel) adjustScroll() {
	contentHeight := m.height - footerHeight - headerHeight - panelPadding
	visibleTasks := contentHeight / taskHeight
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	if m.activeTask < m.scrollOffset {
		m.scrollOffset = m.activeTask
	}

	if m.activeTask >= m.scrollOffset+visibleTasks {
		m.scrollOffset = m.activeTask - visibleTasks + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	maxOffset := len(m.tasks) - visibleTasks
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
}

// adjustTagScroll ensures the active tag is visible within the scroll view
func (m *MainWindowModel) adjustTagScroll() {
	contentHeight := m.height - footerHeight - headerHeight - panelPadding
	visibleTags := contentHeight / tagHeight
	if visibleTags < 1 {
		visibleTags = 1
	}

	if m.activeTag < m.tagScrollOffset {
		m.tagScrollOffset = m.activeTag
	}

	if m.activeTag >= m.tagScrollOffset+visibleTags {
		m.tagScrollOffset = m.activeTag - visibleTags + 1
	}

	if m.tagScrollOffset < 0 {
		m.tagScrollOffset = 0
	}

	maxOffset := len(m.tags) - visibleTags
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.tagScrollOffset > maxOffset {
		m.tagScrollOffset = maxOffset
	}
}

func (m MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		m.tasks = msg.tasks
		if m.activeTask == -1 && len(m.tasks) > 0 {
			m.activeTask = 0
		}
		fixActiveTask(&m)
		m.adjustScroll()
		return m, nil

	case tagsLoadedMsg:
		m.tags = msg.tags
		if m.activeTag == -1 && len(m.tags) > 0 {
			m.activeTag = 0
		}
		fixActiveTag(&m)
		m.adjustTagScroll()
		return m, nil

	case tickMsg:
		return m, m.checkDBState

	case dbStateMsg:
		if msg.state.Count != m.lastDBState.Count ||
			!msg.state.LastUpdated.Equal(m.lastDBState.LastUpdated) {
			m.lastDBState = msg.state
			return m, tea.Batch(m.LoadTasks, m.LoadTags, tickCmd())
		}
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustScroll()
		m.adjustTagScroll()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			if m.activePanel == TagsPanel {
				m.activePanel = TasksPanel
			} else {
				m.activePanel = TagsPanel
			}
			return m, nil

		case "j", "down":
			if m.activePanel == TagsPanel {
				m.activeTag++
				fixActiveTag(&m)
				m.adjustTagScroll()
			} else {
				m.activeTask++
				fixActiveTask(&m)
				m.adjustScroll()
			}
			return m, nil

		case "k", "up":
			if m.activePanel == TagsPanel {
				m.activeTag--
				fixActiveTag(&m)
				m.adjustTagScroll()
			} else {
				m.activeTask--
				fixActiveTask(&m)
				m.adjustScroll()
			}
			return m, nil

		case " ": // Space
			if m.activePanel == TagsPanel {
				if m.activeTag >= 0 && m.activeTag < len(m.tags) {
					tagID := m.tags[m.activeTag].ID
					m.selectedTags[tagID] = !m.selectedTags[tagID]
					// Remove from map if false to keep it clean
					if !m.selectedTags[tagID] {
						delete(m.selectedTags, tagID)
					}
					return m, m.LoadTasks
				}
			} else {
				// Toggle task in tasks panel
				if m.activeTask >= 0 && m.activeTask < len(m.tasks) {
					task := m.tasks[m.activeTask]
					task.ToggleState()
					m.service.TasksRepo.Update(&task)
					return m, m.LoadTasks
				}
			}
			return m, nil

		case "enter":
			if m.activePanel == TasksPanel {
				if m.activeTask >= 0 && m.activeTask < len(m.tasks) {
					task := m.tasks[m.activeTask]
					task.ToggleState()
					m.service.TasksRepo.Update(&task)
					return m, m.LoadTasks
				}
			}
			return m, nil

		case "m":
			if m.filterMode == FilterOR {
				m.filterMode = FilterAND
			} else {
				m.filterMode = FilterOR
			}
			// Only reload if there are selected tags
			if len(m.selectedTags) > 0 {
				return m, m.LoadTasks
			}
			return m, nil

		case "h":
			m.hideCompleted = !m.hideCompleted
			return m, m.LoadTasks
		}
	}
	return m, nil
}

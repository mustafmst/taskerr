package tui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// Panel represents which panel is currently active
type Panel int

const (
	TagsPanelFocus Panel = iota
	TasksPanelFocus
)

// MainWindowModel is the root model that orchestrates the TUI panels
type MainWindowModel struct {
	width         int
	height        int
	service       *data.Service
	hideCompleted bool
	lastDBState   tasks.DBState

	activePanel Panel
	tagsPanel   TagsPanelModel
	tasksPanel  TasksPanelModel
}

// NewMainWindowModel creates a new MainWindowModel
func NewMainWindowModel(service *data.Service) MainWindowModel {
	tagsPanel := NewTagsPanelModel()
	tasksPanel := NewTasksPanelModel()
	tasksPanel.SetFocused(true) // Start with tasks focused

	return MainWindowModel{
		service:       service,
		hideCompleted: true,
		activePanel:   TasksPanelFocus,
		tagsPanel:     tagsPanel,
		tasksPanel:    tasksPanel,
	}
}

// Init initializes the model
func (m MainWindowModel) Init() tea.Cmd {
	return tea.Batch(m.loadTasks, m.loadTags, TickCmd())
}

// loadTasks loads tasks from the database with filtering
func (m MainWindowModel) loadTasks() tea.Msg {
	tasksList, err := m.service.TasksRepo.GetAllWithTags()
	if err != nil {
		log.Fatalf("Error retrieving tasks: %v", err)
	}

	// Filter by completion state
	if m.hideCompleted {
		tasksList = filterCompleted(tasksList)
	}

	// Filter by selected tags
	selectedTags := m.tagsPanel.SelectedTags()
	if len(selectedTags) > 0 {
		tasksList = filterByTags(tasksList, selectedTags, m.tagsPanel.GetFilterMode())
	}

	return TasksLoadedMsg{Tasks: tasksList}
}

// loadTags loads tags from the database
func (m MainWindowModel) loadTags() tea.Msg {
	tagsList, err := m.service.TagsRepo.GetAll()
	if err != nil {
		log.Fatalf("Error retrieving tags: %v", err)
	}
	return TagsLoadedMsg{Tags: tagsList}
}

// checkDBState checks the database state for changes
func (m MainWindowModel) checkDBState() tea.Msg {
	state, err := m.service.TasksRepo.GetDBState()
	if err != nil {
		return nil
	}
	return DBStateMsg{State: state}
}

// Update handles all messages
func (m MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TasksLoadedMsg:
		m.tasksPanel.SetTasks(msg.Tasks)
		return m, nil

	case TagsLoadedMsg:
		m.tagsPanel.SetTags(msg.Tags)
		return m, nil

	case TickMsg:
		return m, m.checkDBState

	case DBStateMsg:
		if msg.State.Count != m.lastDBState.Count ||
			!msg.State.LastUpdated.Equal(m.lastDBState.LastUpdated) {
			m.lastDBState = msg.State
			return m, tea.Batch(m.loadTasks, m.loadTags, TickCmd())
		}
		return m, TickCmd()

	case TagToggledMsg, FilterModeChangedMsg:
		// Tag selection or filter mode changed, reload tasks
		return m, m.loadTasks

	case TaskToggledMsg:
		// Update task in DB, then reload
		m.service.TasksRepo.Update(&msg.Task)
		return m, m.loadTasks

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.switchPanel()
			return m, nil
		case "h":
			m.hideCompleted = !m.hideCompleted
			return m, m.loadTasks
		default:
			// Delegate to active panel
			if m.activePanel == TagsPanelFocus {
				var cmd tea.Cmd
				m.tagsPanel, cmd = m.tagsPanel.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			} else {
				var cmd tea.Cmd
				m.tasksPanel, cmd = m.tasksPanel.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m MainWindowModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Render panels
	tagsView := m.tagsPanel.View()
	tasksView := m.tasksPanel.View()

	// Join horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, tagsView, tasksView)

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

// switchPanel toggles focus between panels
func (m *MainWindowModel) switchPanel() {
	if m.activePanel == TagsPanelFocus {
		m.activePanel = TasksPanelFocus
		m.tagsPanel.SetFocused(false)
		m.tasksPanel.SetFocused(true)
	} else {
		m.activePanel = TagsPanelFocus
		m.tagsPanel.SetFocused(true)
		m.tasksPanel.SetFocused(false)
	}
}

// updatePanelSizes calculates and sets panel sizes
func (m *MainWindowModel) updatePanelSizes() {
	tagsWidth := m.width / 4
	if tagsWidth < 15 {
		tagsWidth = 15
	}
	tasksWidth := m.width - tagsWidth
	contentHeight := m.height - FooterHeight

	m.tagsPanel.SetSize(tagsWidth, contentHeight)
	m.tasksPanel.SetSize(tasksWidth, contentHeight)
}

// renderFooter renders the help footer
func (m MainWindowModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(DimTextColor).
		Width(m.width)

	hiddenStatus := "shown"
	if m.hideCompleted {
		hiddenStatus = "hidden"
	}

	footer := fmt.Sprintf(" TAB: switch panel | SPACE: toggle | m: filter mode (%s) | h: completed %s | q: quit",
		m.tagsPanel.GetFilterMode().String(), hiddenStatus)

	return footerStyle.Render(footer)
}

// Helper functions

// filterCompleted filters out completed tasks
func filterCompleted(tasksList []tasks.Task) []tasks.Task {
	filtered := make([]tasks.Task, 0)
	for _, task := range tasksList {
		if !task.State {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// filterByTags filters tasks based on selected tags and filter mode
func filterByTags(tasksList []tasks.Task, selectedTags map[uint]bool, mode FilterMode) []tasks.Task {
	filtered := make([]tasks.Task, 0)

	for _, task := range tasksList {
		if mode == FilterOR {
			// OR mode: task must have at least one selected tag
			for _, tag := range task.Tags {
				if selectedTags[tag.ID] {
					filtered = append(filtered, task)
					break
				}
			}
		} else {
			// AND mode: task must have all selected tags
			hasAll := true
			for tagID := range selectedTags {
				if !selectedTags[tagID] {
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

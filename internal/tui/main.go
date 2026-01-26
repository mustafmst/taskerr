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

	activePanel  Panel
	tagsPanel    TagsPanelModel
	tasksPanel   TasksPanelModel
	addTaskModal AddTaskModalModel
	confirmModal ConfirmModalModel
	statsModal   StatsModalModel
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
		addTaskModal:  NewAddTaskModalModel(),
		confirmModal:  NewConfirmModalModel(),
		statsModal:    NewStatsModalModel(),
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

	case TaskCreatedMsg:
		// Create new task
		task := tasks.Task{Description: msg.Description}
		m.service.TasksRepo.Create(&task)

		// Attach existing tags
		for _, tagID := range msg.TagIDs {
			m.service.TagsRepo.AttachToTask(tagID, task.ID)
		}

		// Create and attach new tags
		for _, tagName := range msg.NewTagNames {
			tag, err := m.service.TagsRepo.GetOrCreate(tagName)
			if err == nil {
				m.service.TagsRepo.AttachToTask(tag.ID, task.ID)
			}
		}

		return m, tea.Batch(m.loadTasks, m.loadTags)

	case ModalClosedMsg:
		// Modal was cancelled, nothing to do
		return m, nil

	case DeleteConfirmedMsg:
		// Handle deletion based on item type
		if msg.ItemType == "task" {
			m.service.TasksRepo.Delete(msg.ItemID)
			return m, m.loadTasks
		} else if msg.ItemType == "tag" {
			m.service.TagsRepo.Delete(msg.ItemID)
			return m, tea.Batch(m.loadTasks, m.loadTags)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		return m, nil

	case tea.KeyMsg:
		// If stats modal is visible, delegate to it
		if m.statsModal.IsVisible() {
			var cmd tea.Cmd
			m.statsModal, cmd = m.statsModal.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// If confirm modal is visible, delegate to it
		if m.confirmModal.IsVisible() {
			var cmd tea.Cmd
			m.confirmModal, cmd = m.confirmModal.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		// If add task modal is visible, delegate to modal
		if m.addTaskModal.IsVisible() {
			var cmd tea.Cmd
			m.addTaskModal, cmd = m.addTaskModal.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.switchPanel()
			return m, nil
		case "h":
			m.hideCompleted = !m.hideCompleted
			return m, m.loadTasks
		case "n":
			// Open add task modal
			m.addTaskModal.SetTags(m.tagsPanel.Tags())
			m.addTaskModal.SetSize(m.width, m.height)
			m.addTaskModal.SetVisible(true)
			return m, nil
		case "s":
			// Open stats modal - get all tasks (unfiltered)
			allTasks, _ := m.service.TasksRepo.GetAllWithTags()
			m.statsModal.Show(allTasks, m.tagsPanel.Tags())
			m.statsModal.SetSize(m.width, m.height)
			return m, nil
		case "d":
			// Delete selected item
			m.confirmModal.SetSize(m.width, m.height)
			if m.activePanel == TasksPanelFocus {
				if task := m.tasksPanel.ActiveTask(); task != nil {
					m.confirmModal.Show(
						"task",
						task.ID,
						"Delete Task?",
						FormatTaskMessage(task.Description),
						"",
					)
				}
			} else {
				if tag := m.tagsPanel.ActiveTag(); tag != nil {
					// Count tasks using this tag
					taskCount, _ := m.service.TagsRepo.CountTasksWithTag(tag.ID)
					m.confirmModal.Show(
						"tag",
						tag.ID,
						"Delete Tag?",
						FormatTagMessage(tag.Name),
						FormatTagWarning(taskCount),
					)
				}
			}
			return m, nil
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

	// Render header
	header := m.renderHeader()

	// Render panels
	tagsView := m.tagsPanel.View()
	tasksView := m.tasksPanel.View()

	// Join horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, tagsView, tasksView)

	// Footer
	footer := m.renderFooter()

	view := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)

	// Overlay modal if visible
	if m.statsModal.IsVisible() {
		view = m.statsModal.View()
	} else if m.confirmModal.IsVisible() {
		view = m.confirmModal.View()
	} else if m.addTaskModal.IsVisible() {
		modalView := m.addTaskModal.View()
		// Place modal over the view using lipgloss.Place
		view = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modalView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
		)
	}

	return view
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
	contentHeight := m.height - FooterHeight - HeaderBarHeight

	m.tagsPanel.SetSize(tagsWidth, contentHeight)
	m.tasksPanel.SetSize(tasksWidth, contentHeight)
}

// renderHeader renders the app title bar
func (m MainWindowModel) renderHeader() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(HeaderBarFgColor).
		Background(HeaderBarBgColor)
	return style.Render("Taskerr")
}

// renderFooter renders the help footer
func (m MainWindowModel) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(DimTextColor).
		Width(m.width)

	hiddenStatus := "hide"
	if m.hideCompleted {
		hiddenStatus = "show"
	}

	footer := fmt.Sprintf(" [Navigate] TAB j/k │ [Edit] n:new d:del SPACE:toggle │ [View] m:filter(%s) h:%s s:stats │ q:quit",
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

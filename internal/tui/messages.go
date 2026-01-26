package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mustafmst/taskerr/internal/data/tasks"
)

// Data loading messages

// TasksLoadedMsg is sent when tasks are loaded from the database
type TasksLoadedMsg struct {
	Tasks []tasks.Task
}

// TagsLoadedMsg is sent when tags are loaded from the database
type TagsLoadedMsg struct {
	Tags []tasks.Tag
}

// Timer messages

// TickMsg is sent on each tick for auto-refresh
type TickMsg time.Time

// DBStateMsg contains the current database state for change detection
type DBStateMsg struct {
	State tasks.DBState
}

// TickCmd returns a command that sends a TickMsg after one second
func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Panel action messages

// TagToggledMsg is sent when a tag selection changes
type TagToggledMsg struct {
	TagID    uint
	Selected bool
}

// TaskToggledMsg is sent when a task completion state changes
type TaskToggledMsg struct {
	Task tasks.Task
}

// FilterModeChangedMsg is sent when the filter mode changes
type FilterModeChangedMsg struct {
	Mode FilterMode
}

// Modal messages

// TaskCreatedMsg is sent when a new task is created via the modal
type TaskCreatedMsg struct {
	Description string
	TagIDs      []uint
	NewTagNames []string
}

// TaskUpdatedMsg is sent when an existing task is updated via the modal
type TaskUpdatedMsg struct {
	TaskID      uint
	Description string
	TagIDs      []uint
	NewTagNames []string
}

// ModalClosedMsg is sent when the modal is closed without saving
type ModalClosedMsg struct{}

// DeleteConfirmedMsg is sent when user confirms deletion
type DeleteConfirmedMsg struct {
	ItemType string // "task" or "tag"
	ItemID   uint
}

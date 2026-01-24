package tasks

import (
	"time"
)

type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Description string     `gorm:"not null" json:"description"`
	State       bool       `gorm:"not null" json:"state"` // true for completed, false for not completed
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	FinishedAt  *time.Time `json:"finished_at"` // Nullable field
	Tags        []Tag      `gorm:"many2many:task_tags;" json:"tags,omitempty"`
}

func (t *Task) TableName() string {
	return "tasks"
}

func (t *Task) IsCompleted() bool {
	return t.State
}

func (t *Task) MarkAsCompleted() {
	t.State = true
	t.SetFinishedAt(time.Now())
}

func (t *Task) MarkAsNotCompleted() {
	t.State = false
	t.FinishedAt = nil
}

func (t *Task) ToggleState() {
	t.State = !t.State
	if t.State {
		t.SetFinishedAt(time.Now())
	}
}

func (t *Task) SetFinishedAt(finishedAt time.Time) {
	t.FinishedAt = &finishedAt
}

func (t *Task) UpdateDescription(description string) {
	t.Description = description
}

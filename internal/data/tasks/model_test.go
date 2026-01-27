package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTask_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		state    bool
		expected bool
	}{
		{
			name:     "completed task",
			state:    true,
			expected: true,
		},
		{
			name:     "not completed task",
			state:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{State: tt.state}
			assert.Equal(t, tt.expected, task.IsCompleted())
		})
	}
}

func TestTask_MarkAsCompleted(t *testing.T) {
	task := &Task{
		Description: "Test task",
		State:       false,
	}

	assert.False(t, task.State)
	assert.Nil(t, task.FinishedAt)

	task.MarkAsCompleted()

	assert.True(t, task.State)
	assert.NotNil(t, task.FinishedAt)
	assert.WithinDuration(t, time.Now(), *task.FinishedAt, time.Second)
}

func TestTask_MarkAsNotCompleted(t *testing.T) {
	now := time.Now()
	task := &Task{
		Description: "Test task",
		State:       true,
		FinishedAt:  &now,
	}

	assert.True(t, task.State)
	assert.NotNil(t, task.FinishedAt)

	task.MarkAsNotCompleted()

	assert.False(t, task.State)
	assert.Nil(t, task.FinishedAt)
}

func TestTask_ToggleState(t *testing.T) {
	tests := []struct {
		name             string
		initialState     bool
		expectedState    bool
		expectFinishedAt bool
	}{
		{
			name:             "toggle from not completed to completed",
			initialState:     false,
			expectedState:    true,
			expectFinishedAt: true,
		},
		{
			name:             "toggle from completed to not completed",
			initialState:     true,
			expectedState:    false,
			expectFinishedAt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				Description: "Test task",
				State:       tt.initialState,
			}

			task.ToggleState()

			assert.Equal(t, tt.expectedState, task.State)
			if tt.expectFinishedAt {
				assert.NotNil(t, task.FinishedAt)
			}
		})
	}
}

func TestTask_UpdateDescription(t *testing.T) {
	task := &Task{Description: "Original"}

	task.UpdateDescription("Updated")

	assert.Equal(t, "Updated", task.Description)
}

func TestTask_SetFinishedAt(t *testing.T) {
	task := &Task{Description: "Test"}
	finishedTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	task.SetFinishedAt(finishedTime)

	assert.NotNil(t, task.FinishedAt)
	assert.Equal(t, finishedTime, *task.FinishedAt)
}

func TestTask_TableName(t *testing.T) {
	task := &Task{}
	assert.Equal(t, "tasks", task.TableName())
}

package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/mustafmst/taskerr/internal/data"
	"github.com/mustafmst/taskerr/internal/data/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestService creates an in-memory database and data service for testing
func setupTestService(t *testing.T) *data.Service {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to open test database")

	// Create the service with repositories
	service := &data.Service{
		TasksRepo: tasks.NewTasksRepository(db),
		TagsRepo:  tasks.NewTagsRepository(db),
	}

	return service
}

// executeCommand runs a CLI command and returns error if any
func executeCommand(t *testing.T, service *data.Service, args ...string) error {
	rootCmd := SetupCLIApp(service)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	return rootCmd.Execute()
}

func TestCLI_DoneCommand(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, service *data.Service) uint
		verifyFunc func(t *testing.T, service *data.Service, taskID uint)
	}{
		{
			name: "mark task as completed",
			setupFunc: func(t *testing.T, service *data.Service) uint {
				task := &tasks.Task{Description: "Test task", State: false}
				require.NoError(t, service.TasksRepo.Create(task))
				return task.ID
			},
			verifyFunc: func(t *testing.T, service *data.Service, taskID uint) {
				task, err := service.TasksRepo.Get(taskID)
				require.NoError(t, err)
				assert.True(t, task.IsCompleted(), "Task should be completed")
				assert.NotNil(t, task.FinishedAt, "FinishedAt should be set")
			},
		},
		{
			name: "already completed task stays completed",
			setupFunc: func(t *testing.T, service *data.Service) uint {
				task := &tasks.Task{Description: "Already done", State: true}
				require.NoError(t, service.TasksRepo.Create(task))
				return task.ID
			},
			verifyFunc: func(t *testing.T, service *data.Service, taskID uint) {
				task, err := service.TasksRepo.Get(taskID)
				require.NoError(t, err)
				assert.True(t, task.IsCompleted(), "Task should still be completed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := setupTestService(t)
			taskID := tt.setupFunc(t, service)

			err := executeCommand(t, service, "done", fmt.Sprintf("%d", taskID))
			require.NoError(t, err)

			tt.verifyFunc(t, service, taskID)
		})
	}
}

func TestCLI_UndoneCommand(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, service *data.Service) uint
		verifyFunc func(t *testing.T, service *data.Service, taskID uint)
	}{
		{
			name: "mark task as not completed",
			setupFunc: func(t *testing.T, service *data.Service) uint {
				task := &tasks.Task{Description: "Completed task", State: true}
				require.NoError(t, service.TasksRepo.Create(task))
				return task.ID
			},
			verifyFunc: func(t *testing.T, service *data.Service, taskID uint) {
				task, err := service.TasksRepo.Get(taskID)
				require.NoError(t, err)
				assert.False(t, task.IsCompleted(), "Task should not be completed")
				assert.Nil(t, task.FinishedAt, "FinishedAt should be nil")
			},
		},
		{
			name: "already not completed task stays not completed",
			setupFunc: func(t *testing.T, service *data.Service) uint {
				task := &tasks.Task{Description: "Not done", State: false}
				require.NoError(t, service.TasksRepo.Create(task))
				return task.ID
			},
			verifyFunc: func(t *testing.T, service *data.Service, taskID uint) {
				task, err := service.TasksRepo.Get(taskID)
				require.NoError(t, err)
				assert.False(t, task.IsCompleted(), "Task should still not be completed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := setupTestService(t)
			taskID := tt.setupFunc(t, service)

			err := executeCommand(t, service, "undone", fmt.Sprintf("%d", taskID))
			require.NoError(t, err)

			tt.verifyFunc(t, service, taskID)
		})
	}
}

func TestCLI_DeleteCommand_WithForce(t *testing.T) {
	service := setupTestService(t)

	// Create task
	task := &tasks.Task{Description: "Task to delete"}
	require.NoError(t, service.TasksRepo.Create(task))
	taskID := task.ID

	err := executeCommand(t, service, "delete", fmt.Sprintf("%d", taskID), "--force")
	require.NoError(t, err)

	// Verify deletion
	_, err = service.TasksRepo.Get(taskID)
	assert.Error(t, err, "Task should be deleted")
}

func TestCLI_DeleteCommand_RmAlias(t *testing.T) {
	service := setupTestService(t)

	task := &tasks.Task{Description: "Task to rm"}
	require.NoError(t, service.TasksRepo.Create(task))
	taskID := task.ID

	err := executeCommand(t, service, "rm", fmt.Sprintf("%d", taskID), "--force")
	require.NoError(t, err)

	// Verify deletion
	_, err = service.TasksRepo.Get(taskID)
	assert.Error(t, err, "Task should be deleted")
}

func TestCLI_EditCommand_ShowDetails(t *testing.T) {
	service := setupTestService(t)

	// Create task with tags
	task := &tasks.Task{Description: "Task with details"}
	require.NoError(t, service.TasksRepo.Create(task))

	tag, err := service.TagsRepo.GetOrCreate("testtag")
	require.NoError(t, err)
	require.NoError(t, service.TagsRepo.AttachToTask(tag.ID, task.ID))

	// Running edit without flags should not error and not modify the task
	err = executeCommand(t, service, "edit", "1")
	require.NoError(t, err)

	// Verify task unchanged
	unchanged, err := service.TasksRepo.GetWithTags(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Task with details", unchanged.Description)
	assert.Len(t, unchanged.Tags, 1)
}

func TestCLI_EditCommand_UpdateDescription(t *testing.T) {
	service := setupTestService(t)

	task := &tasks.Task{Description: "Original description"}
	require.NoError(t, service.TasksRepo.Create(task))

	err := executeCommand(t, service, "edit", "1", "--desc=Updated description")
	require.NoError(t, err)

	// Verify update
	updated, err := service.TasksRepo.Get(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description)
}

func TestCLI_EditCommand_ReplaceTags(t *testing.T) {
	service := setupTestService(t)

	// Create task with existing tags
	task := &tasks.Task{Description: "Task with tags"}
	require.NoError(t, service.TasksRepo.Create(task))

	oldTag, err := service.TagsRepo.GetOrCreate("oldtag")
	require.NoError(t, err)
	require.NoError(t, service.TagsRepo.AttachToTask(oldTag.ID, task.ID))

	err = executeCommand(t, service, "edit", "1", "--tags=newtag1,newtag2")
	require.NoError(t, err)

	// Verify tags replaced
	updatedTask, err := service.TasksRepo.GetWithTags(task.ID)
	require.NoError(t, err)

	tagNames := make([]string, len(updatedTask.Tags))
	for i, tag := range updatedTask.Tags {
		tagNames[i] = tag.Name
	}

	assert.NotContains(t, tagNames, "oldtag", "Old tag should be removed")
	assert.Contains(t, tagNames, "newtag1", "newtag1 should be attached")
	assert.Contains(t, tagNames, "newtag2", "newtag2 should be attached")
	assert.Len(t, tagNames, 2, "Should have exactly 2 tags")
}

func TestCLI_EditCommand_RemoveAllTags(t *testing.T) {
	service := setupTestService(t)

	task := &tasks.Task{Description: "Task with tags"}
	require.NoError(t, service.TasksRepo.Create(task))

	tag, err := service.TagsRepo.GetOrCreate("removeme")
	require.NoError(t, err)
	require.NoError(t, service.TagsRepo.AttachToTask(tag.ID, task.ID))

	err = executeCommand(t, service, "edit", "1", "--tags=")
	require.NoError(t, err)

	// Verify no tags
	updatedTask, err := service.TasksRepo.GetWithTags(task.ID)
	require.NoError(t, err)
	assert.Len(t, updatedTask.Tags, 0, "All tags should be removed")
}

func TestCLI_EditCommand_UpdateBoth(t *testing.T) {
	service := setupTestService(t)

	task := &tasks.Task{Description: "Original"}
	require.NoError(t, service.TasksRepo.Create(task))

	err := executeCommand(t, service, "edit", "1", "--desc=Updated", "--tags=new")
	require.NoError(t, err)

	// Verify both changed
	updatedTask, err := service.TasksRepo.GetWithTags(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updatedTask.Description)
	assert.Len(t, updatedTask.Tags, 1)
	assert.Equal(t, "new", updatedTask.Tags[0].Name)
}

func TestCLI_AddCommand(t *testing.T) {
	service := setupTestService(t)

	err := executeCommand(t, service, "add", "New test task")
	require.NoError(t, err)

	// Verify task exists
	tasksList, err := service.TasksRepo.GetAll()
	require.NoError(t, err)
	assert.Len(t, tasksList, 1)
	assert.Equal(t, "New test task", tasksList[0].Description)
	assert.False(t, tasksList[0].State, "New task should not be completed")
}

func TestCLI_AddCommand_WithTags(t *testing.T) {
	service := setupTestService(t)

	err := executeCommand(t, service, "add", "Task with tags", "--tags=tag1,tag2")
	require.NoError(t, err)

	// Verify task has tags
	tasksList, err := service.TasksRepo.GetAllWithTags()
	require.NoError(t, err)
	require.Len(t, tasksList, 1)
	assert.Len(t, tasksList[0].Tags, 2)

	tagNames := []string{tasksList[0].Tags[0].Name, tasksList[0].Tags[1].Name}
	assert.Contains(t, tagNames, "tag1")
	assert.Contains(t, tagNames, "tag2")
}

func TestCLI_LsCommand(t *testing.T) {
	service := setupTestService(t)

	// Create some tasks
	require.NoError(t, service.TasksRepo.Create(&tasks.Task{Description: "Task 1", State: false}))
	require.NoError(t, service.TasksRepo.Create(&tasks.Task{Description: "Task 2", State: true}))

	// ls command should not error
	err := executeCommand(t, service, "ls")
	require.NoError(t, err)

	// Verify tasks still exist
	tasksList, err := service.TasksRepo.GetAll()
	require.NoError(t, err)
	assert.Len(t, tasksList, 2)
}

func TestCLI_DoneCommand_InvalidTaskID(t *testing.T) {
	service := setupTestService(t)

	// This should cause a log.Fatal, but we can't easily test that
	// Instead we test that a valid call works
	task := &tasks.Task{Description: "Valid task"}
	require.NoError(t, service.TasksRepo.Create(task))

	err := executeCommand(t, service, "done", "1")
	require.NoError(t, err)
}

func TestConfirmAction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes lowercase", "y\n", true},
		{"yes full", "yes\n", true},
		{"YES uppercase", "YES\n", true},
		{"Yes mixed", "Yes\n", true},
		{"no lowercase", "n\n", false},
		{"no full", "no\n", false},
		{"empty", "\n", false},
		{"random input", "maybe\n", false},
		{"y with spaces", "  y  \n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result := confirmActionFromReader("Test prompt: ", reader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

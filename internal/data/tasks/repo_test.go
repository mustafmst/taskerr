package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to open test database")

	// Run migrations
	err = db.AutoMigrate(&Task{}, &Tag{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

func TestTasksRepository_GetWithTags(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, db *gorm.DB) uint
		expectError bool
		expectTags  int
	}{
		{
			name: "get task with multiple tags",
			setupFunc: func(t *testing.T, db *gorm.DB) uint {
				task := &Task{Description: "Test task"}
				require.NoError(t, db.Create(task).Error)

				tag1 := &Tag{Name: "tag1", Color: "#ff0000"}
				tag2 := &Tag{Name: "tag2", Color: "#00ff00"}
				require.NoError(t, db.Create(tag1).Error)
				require.NoError(t, db.Create(tag2).Error)

				// Attach tags via join table
				require.NoError(t, db.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", task.ID, tag1.ID).Error)
				require.NoError(t, db.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", task.ID, tag2.ID).Error)

				return task.ID
			},
			expectError: false,
			expectTags:  2,
		},
		{
			name: "get task with no tags",
			setupFunc: func(t *testing.T, db *gorm.DB) uint {
				task := &Task{Description: "Task without tags"}
				require.NoError(t, db.Create(task).Error)
				return task.ID
			},
			expectError: false,
			expectTags:  0,
		},
		{
			name: "get non-existent task",
			setupFunc: func(t *testing.T, db *gorm.DB) uint {
				return 9999 // Non-existent ID
			},
			expectError: true,
			expectTags:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := &TasksRepository{db: db}

			taskID := tt.setupFunc(t, db)

			task, err := repo.GetWithTags(taskID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, taskID, task.ID)
				assert.Len(t, task.Tags, tt.expectTags)
			}
		})
	}
}

func TestTasksRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTasksRepository(db)

	task := &Task{
		Description: "New task",
		State:       false,
	}

	err := repo.Create(task)
	require.NoError(t, err)
	assert.NotZero(t, task.ID)
	assert.Equal(t, "New task", task.Description)
	assert.False(t, task.State)
}

func TestTasksRepository_Get(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, db *gorm.DB) uint
		expectError bool
	}{
		{
			name: "get existing task",
			setupFunc: func(t *testing.T, db *gorm.DB) uint {
				task := &Task{Description: "Existing task"}
				require.NoError(t, db.Create(task).Error)
				return task.ID
			},
			expectError: false,
		},
		{
			name: "get non-existent task",
			setupFunc: func(t *testing.T, db *gorm.DB) uint {
				return 9999
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := &TasksRepository{db: db}

			taskID := tt.setupFunc(t, db)
			task, err := repo.Get(taskID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, taskID, task.ID)
			}
		})
	}
}

func TestTasksRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTasksRepository(db)

	// Create initial task
	task := &Task{Description: "Original description", State: false}
	require.NoError(t, repo.Create(task))

	// Update task
	task.Description = "Updated description"
	task.MarkAsCompleted()

	err := repo.Update(task)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.Get(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description)
	assert.True(t, updated.State)
	assert.NotNil(t, updated.FinishedAt)
}

func TestTasksRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTasksRepository(db)

	// Create task
	task := &Task{Description: "Task to delete"}
	require.NoError(t, repo.Create(task))

	// Delete task
	err := repo.Delete(task.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(task.ID)
	assert.Error(t, err)
}

func TestTasksRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTasksRepository(db)

	// Create multiple tasks
	tasks := []*Task{
		{Description: "Task 1"},
		{Description: "Task 2"},
		{Description: "Task 3"},
	}

	for _, task := range tasks {
		require.NoError(t, repo.Create(task))
	}

	// Get all
	allTasks, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, allTasks, 3)
}

func TestTasksRepository_GetAllWithTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTasksRepository(db)

	// Create task with tags
	task := &Task{Description: "Task with tags"}
	require.NoError(t, repo.Create(task))

	tag := &Tag{Name: "testtag", Color: "#ffffff"}
	require.NoError(t, db.Create(tag).Error)
	require.NoError(t, db.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", task.ID, tag.ID).Error)

	// Get all with tags
	allTasks, err := repo.GetAllWithTags()
	require.NoError(t, err)
	require.Len(t, allTasks, 1)
	assert.Len(t, allTasks[0].Tags, 1)
	assert.Equal(t, "testtag", allTasks[0].Tags[0].Name)
}

func TestTasksRepository_GetDBState(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTasksRepository(db)

	// Empty database
	state, err := repo.GetDBState()
	require.NoError(t, err)
	assert.Equal(t, int64(0), state.Count)

	// Add tasks
	require.NoError(t, repo.Create(&Task{Description: "Task 1"}))
	require.NoError(t, repo.Create(&Task{Description: "Task 2"}))

	state, err = repo.GetDBState()
	require.NoError(t, err)
	assert.Equal(t, int64(2), state.Count)
	assert.False(t, state.LastUpdated.IsZero())
}

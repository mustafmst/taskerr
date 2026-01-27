package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTagsRepository(db)

	tag := &Tag{
		Name:        "TestTag",
		Color:       "#ff0000",
		Description: "Test description",
	}

	err := repo.Create(tag)
	require.NoError(t, err)
	assert.NotZero(t, tag.ID)
	assert.Equal(t, "testtag", tag.Name) // Should be lowercased
}

func TestTagsRepository_GetByName(t *testing.T) {
	tests := []struct {
		name        string
		searchName  string
		expectError bool
	}{
		{
			name:        "find existing tag (exact case)",
			searchName:  "findme",
			expectError: false,
		},
		{
			name:        "find existing tag (different case)",
			searchName:  "FINDME",
			expectError: false,
		},
		{
			name:        "find existing tag (mixed case)",
			searchName:  "FindMe",
			expectError: false,
		},
		{
			name:        "non-existent tag",
			searchName:  "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewTagsRepository(db)

			// Create a tag to find
			require.NoError(t, repo.Create(&Tag{Name: "findme", Color: "#000000"}))

			tag, err := repo.GetByName(tt.searchName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tag)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tag)
				assert.Equal(t, "findme", tag.Name)
			}
		})
	}
}

func TestTagsRepository_GetOrCreate(t *testing.T) {
	tests := []struct {
		name           string
		tagName        string
		setupExisting  bool
		expectNewColor bool
	}{
		{
			name:           "create new tag",
			tagName:        "newtag",
			setupExisting:  false,
			expectNewColor: true,
		},
		{
			name:           "get existing tag",
			tagName:        "existingtag",
			setupExisting:  true,
			expectNewColor: false,
		},
		{
			name:           "create with different case",
			tagName:        "UPPERCASE",
			setupExisting:  false,
			expectNewColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewTagsRepository(db)

			if tt.setupExisting {
				existing := &Tag{Name: tt.tagName, Color: "#existing"}
				require.NoError(t, repo.Create(existing))
			}

			tag, err := repo.GetOrCreate(tt.tagName)
			require.NoError(t, err)
			require.NotNil(t, tag)

			if tt.expectNewColor {
				// Should have a color from the palette
				assert.Contains(t, TagColorPalette, tag.Color)
			} else {
				assert.Equal(t, "#existing", tag.Color)
			}
		})
	}
}

func TestTagsRepository_AttachDetachToTask(t *testing.T) {
	db := setupTestDB(t)
	tagsRepo := NewTagsRepository(db)
	tasksRepo := NewTasksRepository(db)

	// Create task and tag
	task := &Task{Description: "Test task"}
	require.NoError(t, tasksRepo.Create(task))

	tag := &Tag{Name: "testtag", Color: "#ff0000"}
	require.NoError(t, tagsRepo.Create(tag))

	// Attach tag
	err := tagsRepo.AttachToTask(tag.ID, task.ID)
	require.NoError(t, err)

	// Verify attachment
	tags, err := tagsRepo.GetTaskTags(task.ID)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, tag.ID, tags[0].ID)

	// Detach tag
	err = tagsRepo.DetachFromTask(tag.ID, task.ID)
	require.NoError(t, err)

	// Verify detachment
	tags, err = tagsRepo.GetTaskTags(task.ID)
	require.NoError(t, err)
	assert.Len(t, tags, 0)
}

func TestTagsRepository_AttachToTask_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	tagsRepo := NewTagsRepository(db)
	tasksRepo := NewTasksRepository(db)

	task := &Task{Description: "Test task"}
	require.NoError(t, tasksRepo.Create(task))

	tag := &Tag{Name: "testtag", Color: "#ff0000"}
	require.NoError(t, tagsRepo.Create(tag))

	// Attach same tag twice (should not error due to INSERT OR IGNORE)
	err := tagsRepo.AttachToTask(tag.ID, task.ID)
	require.NoError(t, err)

	err = tagsRepo.AttachToTask(tag.ID, task.ID)
	require.NoError(t, err)

	// Should still only have one association
	tags, err := tagsRepo.GetTaskTags(task.ID)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}

func TestTagsRepository_CountTasksWithTag(t *testing.T) {
	db := setupTestDB(t)
	tagsRepo := NewTagsRepository(db)
	tasksRepo := NewTasksRepository(db)

	// Create tag
	tag := &Tag{Name: "popular", Color: "#ff0000"}
	require.NoError(t, tagsRepo.Create(tag))

	// Initially no tasks
	count, err := tagsRepo.CountTasksWithTag(tag.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create and attach tasks
	for i := 0; i < 3; i++ {
		task := &Task{Description: "Task"}
		require.NoError(t, tasksRepo.Create(task))
		require.NoError(t, tagsRepo.AttachToTask(tag.ID, task.ID))
	}

	count, err = tagsRepo.CountTasksWithTag(tag.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestTagsRepository_DeleteByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTagsRepository(db)

	// Create tag
	tag := &Tag{Name: "deleteme", Color: "#ff0000"}
	require.NoError(t, repo.Create(tag))

	// Verify exists
	_, err := repo.GetByName("deleteme")
	require.NoError(t, err)

	// Delete by name
	err = repo.DeleteByName("DeleteMe") // Different case
	require.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByName("deleteme")
	assert.Error(t, err)
}

func TestTagsRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTagsRepository(db)

	// Create multiple tags
	for _, name := range []string{"tag1", "tag2", "tag3"} {
		require.NoError(t, repo.Create(&Tag{Name: name, Color: "#000000"}))
	}

	tags, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, tags, 3)
}

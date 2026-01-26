package tasks

import (
	"strings"

	"gorm.io/gorm"
)

// TagsRepository handles database operations for tags
type TagsRepository struct {
	db         *gorm.DB
	colorIndex int
}

// NewTagsRepository creates a new instance of TagsRepository
func NewTagsRepository(db *gorm.DB) *TagsRepository {
	db.AutoMigrate(&Tag{})
	return &TagsRepository{db: db, colorIndex: 0}
}

// Create creates a new tag in the database
func (r *TagsRepository) Create(tag *Tag) error {
	tag.Name = strings.ToLower(strings.TrimSpace(tag.Name))
	return r.db.Create(tag).Error
}

// Get retrieves a tag by its ID
func (r *TagsRepository) Get(id uint) (*Tag, error) {
	var tag Tag
	if err := r.db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByName retrieves a tag by its name (case-insensitive)
func (r *TagsRepository) GetByName(name string) (*Tag, error) {
	var tag Tag
	name = strings.ToLower(strings.TrimSpace(name))
	if err := r.db.Where("name = ?", name).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetOrCreate retrieves a tag by name or creates it with a random color from palette
func (r *TagsRepository) GetOrCreate(name string) (*Tag, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	var tag Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err == gorm.ErrRecordNotFound {
		tag = Tag{
			Name:  name,
			Color: TagColorPalette[r.colorIndex%len(TagColorPalette)],
		}
		r.colorIndex++
		if err := r.db.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	} else if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetAll retrieves all tags from the database
func (r *TagsRepository) GetAll() ([]Tag, error) {
	var tags []Tag
	if err := r.db.Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// Update updates an existing tag in the database
func (r *TagsRepository) Update(tag *Tag) error {
	tag.Name = strings.ToLower(strings.TrimSpace(tag.Name))
	return r.db.Save(tag).Error
}

// Delete deletes a tag by its ID
func (r *TagsRepository) Delete(id uint) error {
	return r.db.Delete(&Tag{}, id).Error
}

// DeleteByName deletes a tag by its name
func (r *TagsRepository) DeleteByName(name string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	return r.db.Where("name = ?", name).Delete(&Tag{}).Error
}

// AttachToTask attaches a tag to a task
func (r *TagsRepository) AttachToTask(tagID uint, taskID uint) error {
	return r.db.Exec("INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID).Error
}

// DetachFromTask removes a tag from a task
func (r *TagsRepository) DetachFromTask(tagID uint, taskID uint) error {
	return r.db.Exec("DELETE FROM task_tags WHERE task_id = ? AND tag_id = ?", taskID, tagID).Error
}

// GetTaskTags retrieves all tags for a specific task
func (r *TagsRepository) GetTaskTags(taskID uint) ([]Tag, error) {
	var tags []Tag
	err := r.db.
		Joins("JOIN task_tags ON task_tags.tag_id = tags.id").
		Where("task_tags.task_id = ?", taskID).
		Find(&tags).Error
	return tags, err
}

// CountTasksWithTag returns the number of tasks that have this tag
func (r *TagsRepository) CountTasksWithTag(tagID uint) (int64, error) {
	var count int64
	err := r.db.Table("task_tags").Where("tag_id = ?", tagID).Count(&count).Error
	return count, err
}

package tasks

import (
	"time"

	"gorm.io/gorm"
)

type TasksRepository struct {
	db *gorm.DB
}

// DBState represents the current state of the database for change detection
type DBState struct {
	Count       int64
	LastUpdated time.Time
}

// NewTasksRepository creates a new instance of TasksRepository
func NewTasksRepository(db *gorm.DB) *TasksRepository {
	db.AutoMigrate(&Task{})
	return &TasksRepository{db: db}
}

// Create creates a new task in the database
func (r *TasksRepository) Create(task *Task) error {
	return r.db.Create(task).Error
}

// Get retrieves a task by its ID
func (r *TasksRepository) Get(id uint) (*Task, error) {
	var task Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetAll retrieves all tasks from the database
func (r *TasksRepository) GetAll() ([]Task, error) {
	var tasks []Task
	if err := r.db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetAllWithTags retrieves all tasks with their associated tags preloaded
func (r *TasksRepository) GetAllWithTags() ([]Task, error) {
	var tasks []Task
	if err := r.db.Preload("Tags").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// Update updates an existing task in the database
func (r *TasksRepository) Update(task *Task) error {
	return r.db.Save(task).Error
}

// Delete deletes a task by its ID
func (r *TasksRepository) Delete(id uint) error {
	return r.db.Delete(&Task{}, id).Error
}

// GetDBState returns the current database state for change detection
func (r *TasksRepository) GetDBState() (DBState, error) {
	var state DBState
	var count int64

	if err := r.db.Model(&Task{}).Count(&count).Error; err != nil {
		return state, err
	}

	// Get the latest updated task to determine last update time
	var latestTask Task
	err := r.db.Order("updated_at DESC").First(&latestTask).Error
	if err == nil {
		state.LastUpdated = latestTask.UpdatedAt
	}
	// If no tasks found, LastUpdated remains zero value

	state.Count = count
	return state, nil
}

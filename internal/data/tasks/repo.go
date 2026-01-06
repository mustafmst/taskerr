package tasks

import "gorm.io/gorm"

type TasksRepository struct {
	db *gorm.DB
}

// NewTasksRepository creates a new instance of TasksRepository
func NewTasksRepository(db *gorm.DB) *TasksRepository {
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

// Update updates an existing task in the database
func (r *TasksRepository) Update(task *Task) error {
	return r.db.Save(task).Error
}

// Delete deletes a task by its ID
func (r *TasksRepository) Delete(id uint) error {
	return r.db.Delete(&Task{}, id).Error
}

package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/mustafmst/taskerr/internal/config"
	"github.com/mustafmst/taskerr/internal/data/tasks"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Service struct {
	DB *gorm.DB
	// Add repositories here as needed, e.g., UserRepository, TaskRepository, etc.
	TasksRepo *tasks.TasksRepository
	TagsRepo  *tasks.TagsRepository
}

// DBState captures enough database state to detect external changes.
type DBState struct {
	TaskCount      int64
	TagCount       int64
	TaskTagCount   int64
	LastTaskUpdate time.Time
	LastTagUpdate  time.Time
}

// NewService initializes the data service using the configuration loaded from config package.
func NewService(cfg *config.Config) (*Service, error) {
	db, err := initDB(cfg)
	if err != nil {
		return nil, err
	}

	tasksRepo := tasks.NewTasksRepository(db)
	tagsRepo := tasks.NewTagsRepository(db)

	return &Service{DB: db, TasksRepo: tasksRepo, TagsRepo: tagsRepo}, nil
}

func (s *Service) Close() error {
	return nil // GORM doesn't require closing the DB connection explicitly
}

// GetDBState returns a compact snapshot used by the TUI polling refresh loop.
func (s *Service) GetDBState() (DBState, error) {
	var state DBState

	if err := s.DB.Model(&tasks.Task{}).Count(&state.TaskCount).Error; err != nil {
		return state, err
	}
	if err := s.DB.Model(&tasks.Tag{}).Count(&state.TagCount).Error; err != nil {
		return state, err
	}
	if err := s.DB.Table("task_tags").Count(&state.TaskTagCount).Error; err != nil {
		return state, err
	}

	var latestTask tasks.Task
	if err := s.DB.Order("updated_at DESC").First(&latestTask).Error; err == nil {
		state.LastTaskUpdate = latestTask.UpdatedAt
	}

	var latestTag tasks.Tag
	if err := s.DB.Order("updated_at DESC").First(&latestTag).Error; err == nil {
		state.LastTagUpdate = latestTag.UpdatedAt
	}

	return state, nil
}

// CreateTask creates a task and attaches both existing and new tags.
func (s *Service) CreateTask(description string, tagIDs []uint, newTagNames []string) (*tasks.Task, error) {
	task := &tasks.Task{
		Description: strings.TrimSpace(description),
		State:       false,
	}
	if err := s.TasksRepo.Create(task); err != nil {
		return nil, err
	}

	if err := s.syncTaskTags(task.ID, tagIDs, newTagNames); err != nil {
		return nil, err
	}

	return s.TasksRepo.Get(task.ID)
}

// UpdateTask updates a task description and fully replaces its tag set.
func (s *Service) UpdateTask(taskID uint, description string, tagIDs []uint, newTagNames []string) error {
	task, err := s.TasksRepo.Get(taskID)
	if err != nil {
		return err
	}

	task.Description = strings.TrimSpace(description)
	if err := s.TasksRepo.Update(task); err != nil {
		return err
	}

	return s.syncTaskTags(taskID, tagIDs, newTagNames)
}

// DeleteTask removes a task by ID.
func (s *Service) DeleteTask(taskID uint) error {
	return s.TasksRepo.Delete(taskID)
}

// DeleteTag removes a tag by ID.
func (s *Service) DeleteTag(tagID uint) error {
	return s.TagsRepo.Delete(tagID)
}

// AttachTagToTask attaches a tag, creating it if needed.
func (s *Service) AttachTagToTask(taskID uint, tagName string) (*tasks.Tag, error) {
	if _, err := s.TasksRepo.Get(taskID); err != nil {
		return nil, err
	}

	tag, err := s.TagsRepo.GetOrCreate(tagName)
	if err != nil {
		return nil, err
	}

	if err := s.TagsRepo.AttachToTask(tag.ID, taskID); err != nil {
		return nil, err
	}

	return tag, nil
}

// DetachTagFromTask detaches an existing tag from a task.
func (s *Service) DetachTagFromTask(taskID uint, tagName string) error {
	tag, err := s.TagsRepo.GetByName(tagName)
	if err != nil {
		return err
	}

	return s.TagsRepo.DetachFromTask(tag.ID, taskID)
}

func (s *Service) syncTaskTags(taskID uint, tagIDs []uint, newTagNames []string) error {
	currentTags, err := s.TagsRepo.GetTaskTags(taskID)
	if err != nil {
		return err
	}

	currentTagIDs := make(map[uint]bool, len(currentTags))
	for _, tag := range currentTags {
		currentTagIDs[tag.ID] = true
	}

	desiredTagIDs := make(map[uint]bool, len(tagIDs)+len(newTagNames))
	for _, tagID := range tagIDs {
		desiredTagIDs[tagID] = true
	}

	for _, tagName := range newTagNames {
		tag, err := s.TagsRepo.GetOrCreate(tagName)
		if err != nil {
			return err
		}
		desiredTagIDs[tag.ID] = true
	}

	for _, tag := range currentTags {
		if !desiredTagIDs[tag.ID] {
			if err := s.TagsRepo.DetachFromTask(tag.ID, taskID); err != nil {
				return err
			}
		}
	}

	for tagID := range desiredTagIDs {
		if !currentTagIDs[tagID] {
			if err := s.TagsRepo.AttachToTask(tagID, taskID); err != nil {
				return err
			}
		}
	}

	return nil
}

// initDB initializes a GORM DB object based on the provided configuration.
func initDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBProvider {
	case "sqlite":
		dialector = sqlite.Open(cfg.DBConnection)
	case "mysql":
		dialector = mysql.Open(cfg.DBConnection)
	case "postgres":
		dialector = postgres.Open(cfg.DBConnection)
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", cfg.DBProvider)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

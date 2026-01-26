package data

import (
	"fmt"

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

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

// GetStats returns computed statistics from the database
func (r *TasksRepository) GetStats() (*TaskStats, error) {
	stats := &TaskStats{}

	// 1. Basic counts (provider-agnostic via GORM)
	if err := r.getBasicCounts(stats); err != nil {
		return nil, err
	}

	// 2. Time-based completions (provider-specific SQL)
	if err := r.getTimeBasedCounts(stats); err != nil {
		return nil, err
	}

	// 3. Average completion time (provider-specific SQL)
	if err := r.getAvgCompletionTime(stats); err != nil {
		return nil, err
	}

	// 4. Daily activity (provider-specific SQL)
	if err := r.getDailyActivity(stats); err != nil {
		return nil, err
	}

	// 5. Tag statistics (provider-agnostic SQL)
	if err := r.getTagStats(stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// getBasicCounts retrieves total, completed, and incomplete task counts
func (r *TasksRepository) getBasicCounts(stats *TaskStats) error {
	if err := r.db.Model(&Task{}).Count(&stats.TotalTasks).Error; err != nil {
		return err
	}

	if err := r.db.Model(&Task{}).Where("state = ?", true).Count(&stats.CompletedTasks).Error; err != nil {
		return err
	}

	stats.IncompleteTasks = stats.TotalTasks - stats.CompletedTasks

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return nil
}

// getTimeBasedCounts retrieves today/week/month completion counts using provider-specific SQL
func (r *TasksRepository) getTimeBasedCounts(stats *TaskStats) error {
	dialect := r.db.Dialector.Name()

	var todaySQL, weekSQL, monthSQL string

	switch dialect {
	case "sqlite":
		todaySQL = `SELECT COUNT(*) FROM tasks WHERE state = 1 AND 
			(finished_at >= date('now', 'localtime', 'start of day') OR 
			(finished_at IS NULL AND updated_at >= date('now', 'localtime', 'start of day')))`
		weekSQL = `SELECT COUNT(*) FROM tasks WHERE state = 1 AND 
			(finished_at >= date('now', 'localtime', 'weekday 0', '-6 days') OR 
			(finished_at IS NULL AND updated_at >= date('now', 'localtime', 'weekday 0', '-6 days')))`
		monthSQL = `SELECT COUNT(*) FROM tasks WHERE state = 1 AND 
			(finished_at >= date('now', 'localtime', 'start of month') OR 
			(finished_at IS NULL AND updated_at >= date('now', 'localtime', 'start of month')))`
	case "mysql":
		todaySQL = `SELECT COUNT(*) FROM tasks WHERE state = 1 AND 
			(finished_at >= CURDATE() OR (finished_at IS NULL AND updated_at >= CURDATE()))`
		weekSQL = `SELECT COUNT(*) FROM tasks WHERE state = 1 AND 
			(finished_at >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY) OR 
			(finished_at IS NULL AND updated_at >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)))`
		monthSQL = `SELECT COUNT(*) FROM tasks WHERE state = 1 AND 
			(finished_at >= DATE_FORMAT(CURDATE(), '%Y-%m-01') OR 
			(finished_at IS NULL AND updated_at >= DATE_FORMAT(CURDATE(), '%Y-%m-01')))`
	case "postgres":
		todaySQL = `SELECT COUNT(*) FROM tasks WHERE state = true AND 
			(finished_at >= CURRENT_DATE OR (finished_at IS NULL AND updated_at >= CURRENT_DATE))`
		weekSQL = `SELECT COUNT(*) FROM tasks WHERE state = true AND 
			(finished_at >= date_trunc('week', CURRENT_DATE) OR 
			(finished_at IS NULL AND updated_at >= date_trunc('week', CURRENT_DATE)))`
		monthSQL = `SELECT COUNT(*) FROM tasks WHERE state = true AND 
			(finished_at >= date_trunc('month', CURRENT_DATE) OR 
			(finished_at IS NULL AND updated_at >= date_trunc('month', CURRENT_DATE)))`
	default:
		// Fallback to GORM for unknown dialects
		return nil
	}

	r.db.Raw(todaySQL).Scan(&stats.TodayCompleted)
	r.db.Raw(weekSQL).Scan(&stats.WeekCompleted)
	r.db.Raw(monthSQL).Scan(&stats.MonthCompleted)

	return nil
}

// getAvgCompletionTime calculates the average time to complete tasks
func (r *TasksRepository) getAvgCompletionTime(stats *TaskStats) error {
	dialect := r.db.Dialector.Name()

	var sql string
	switch dialect {
	case "sqlite":
		sql = `SELECT AVG((julianday(finished_at) - julianday(created_at)) * 86400) 
			FROM tasks WHERE state = 1 AND finished_at IS NOT NULL`
	case "mysql":
		sql = `SELECT AVG(TIMESTAMPDIFF(SECOND, created_at, finished_at)) 
			FROM tasks WHERE state = 1 AND finished_at IS NOT NULL`
	case "postgres":
		sql = `SELECT AVG(EXTRACT(EPOCH FROM (finished_at - created_at))) 
			FROM tasks WHERE state = true AND finished_at IS NOT NULL`
	default:
		return nil
	}

	var avgSeconds *float64
	r.db.Raw(sql).Scan(&avgSeconds)

	if avgSeconds != nil && *avgSeconds > 0 {
		stats.AvgCompletionTime = time.Duration(*avgSeconds) * time.Second
	}

	return nil
}

// getDailyActivity retrieves task completion counts for the last 7 days
func (r *TasksRepository) getDailyActivity(stats *TaskStats) error {
	dialect := r.db.Dialector.Name()
	now := time.Now()

	// Initialize day labels
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -6+i)
		stats.DayLabels[i] = day.Format("Mon")[:3]
	}

	var sql string
	switch dialect {
	case "sqlite":
		sql = `SELECT date(COALESCE(finished_at, updated_at)) as day, COUNT(*) as count
			FROM tasks
			WHERE state = 1 AND COALESCE(finished_at, updated_at) >= date('now', 'localtime', '-6 days')
			GROUP BY date(COALESCE(finished_at, updated_at))
			ORDER BY day`
	case "mysql":
		sql = `SELECT DATE(COALESCE(finished_at, updated_at)) as day, COUNT(*) as count
			FROM tasks
			WHERE state = 1 AND COALESCE(finished_at, updated_at) >= DATE_SUB(CURDATE(), INTERVAL 6 DAY)
			GROUP BY DATE(COALESCE(finished_at, updated_at))
			ORDER BY day`
	case "postgres":
		sql = `SELECT DATE(COALESCE(finished_at, updated_at)) as day, COUNT(*) as count
			FROM tasks
			WHERE state = true AND COALESCE(finished_at, updated_at) >= CURRENT_DATE - INTERVAL '6 days'
			GROUP BY DATE(COALESCE(finished_at, updated_at))
			ORDER BY day`
	default:
		return nil
	}

	type DayCount struct {
		Day   time.Time
		Count int64
	}
	var results []DayCount
	r.db.Raw(sql).Scan(&results)

	// Map results to array indices
	for _, result := range results {
		for i := 0; i < 7; i++ {
			targetDay := now.AddDate(0, 0, -6+i)
			if result.Day.Year() == targetDay.Year() &&
				result.Day.YearDay() == targetDay.YearDay() {
				stats.DailyActivity[i] = result.Count
				break
			}
		}
	}

	return nil
}

// getTagStats retrieves task counts per tag
func (r *TasksRepository) getTagStats(stats *TaskStats) error {
	sql := `SELECT t.name, COUNT(tt.task_id) as count
		FROM tags t
		LEFT JOIN task_tags tt ON t.id = tt.tag_id
		GROUP BY t.id, t.name
		ORDER BY count DESC`

	r.db.Raw(sql).Scan(&stats.TagStats)
	stats.TotalTags = len(stats.TagStats)

	return nil
}

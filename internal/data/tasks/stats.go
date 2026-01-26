package tasks

import "time"

// TaskStats holds computed statistics from the database
type TaskStats struct {
	// Basic counts
	TotalTasks      int64
	CompletedTasks  int64
	IncompleteTasks int64
	CompletionRate  float64

	// Time-based completions
	TodayCompleted int64
	WeekCompleted  int64
	MonthCompleted int64

	// Average completion time
	AvgCompletionTime time.Duration

	// Daily activity (last 7 days) - index 0 is 6 days ago, index 6 is today
	DailyActivity [7]int64
	DayLabels     [7]string

	// Tag statistics
	TotalTags int
	TagStats  []TagStat
}

// TagStat holds tag name and task count
type TagStat struct {
	Name  string
	Count int64
}

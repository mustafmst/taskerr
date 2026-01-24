package tasks

import "time"

// TagColorPalette provides colors for auto-created tags
var TagColorPalette = []string{
	"#e74c3c", // red
	"#3498db", // blue
	"#2ecc71", // green
	"#9b59b6", // purple
	"#f39c12", // orange
	"#1abc9c", // teal
	"#e91e63", // pink
	"#00bcd4", // cyan
}

// Tag represents a label that can be attached to tasks
type Tag struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Color       string    `gorm:"not null" json:"color"`
	Description string    `json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (t *Tag) TableName() string {
	return "tags"
}

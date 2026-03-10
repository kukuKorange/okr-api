package model

import "time"

type TodoItem struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	GoalID      *uint      `json:"goal_id" gorm:"index"`
	Content     string     `json:"content" gorm:"size:500;not null"`
	DueDate     *time.Time `json:"due_date" gorm:"type:date"`
	Priority    int        `json:"priority" gorm:"default:2"`
	IsCompleted bool       `json:"is_completed" gorm:"default:false"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Goal *Goal `json:"goal,omitempty" gorm:"foreignKey:GoalID"`
}

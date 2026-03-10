package model

import "time"

type Milestone struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	GoalID      uint       `json:"goal_id" gorm:"not null;index"`
	Title       string     `json:"title" gorm:"size:200;not null"`
	TargetValue float64    `json:"target_value"`
	TargetDate  time.Time  `json:"target_date" gorm:"type:date"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

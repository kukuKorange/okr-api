package model

import "time"

type GoalLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	GoalID       uint      `json:"goal_id" gorm:"not null;index"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	LogDate      time.Time `json:"log_date" gorm:"type:date;not null"`
	Value        float64   `json:"value" gorm:"not null"`
	Note         string    `json:"note" gorm:"size:1000"`
	PhotoURL     string    `json:"photo_url" gorm:"size:500"`
	PlanTomorrow string    `json:"plan_tomorrow" gorm:"size:1000"`
	CreatedAt    time.Time `json:"created_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

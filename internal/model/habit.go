package model

import "time"

type Habit struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	UserID             uint      `json:"user_id" gorm:"not null;index"`
	FamilyMemberTarget *uint     `json:"family_member_target" gorm:"index"`
	Name               string    `json:"name" gorm:"size:100;not null"`
	Icon               string    `json:"icon" gorm:"size:50;default:'📌'"`
	Color              string    `json:"color" gorm:"size:20;default:'#4F46E5'"`
	Frequency          string    `json:"frequency" gorm:"size:20;not null;default:'daily'"`
	WeeklyTarget       int       `json:"weekly_target" gorm:"default:1"`
	StartDate          time.Time `json:"start_date" gorm:"type:date"`
	IsArchived         bool      `json:"is_archived" gorm:"default:false"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	CheckIns []HabitCheckIn `json:"check_ins,omitempty" gorm:"foreignKey:HabitID"`
}

type HabitCheckIn struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	HabitID   uint      `json:"habit_id" gorm:"not null;index;uniqueIndex:idx_habit_date"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	CheckDate time.Time `json:"check_date" gorm:"type:date;not null;uniqueIndex:idx_habit_date"`
	Note      string    `json:"note" gorm:"size:500"`
	CreatedAt time.Time `json:"created_at"`

	Habit *Habit `json:"habit,omitempty" gorm:"foreignKey:HabitID"`
}

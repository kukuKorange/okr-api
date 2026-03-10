package model

import "time"

type Goal struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"user_id" gorm:"not null;index"`
	FamilyID     *uint      `json:"family_id" gorm:"index"`
	Title        string     `json:"title" gorm:"size:200;not null"`
	Category     string     `json:"category" gorm:"size:50;not null;default:'custom'"`
	Unit         string     `json:"unit" gorm:"size:20"`
	TargetValue  float64    `json:"target_value" gorm:"not null"`
	CurrentValue float64    `json:"current_value" gorm:"default:0"`
	StartValue   float64    `json:"start_value" gorm:"default:0"`
	Direction    string     `json:"direction" gorm:"size:20;not null;default:'increase'"`
	StartDate    time.Time  `json:"start_date" gorm:"type:date"`
	Deadline     time.Time  `json:"deadline" gorm:"type:date"`
	Status       string     `json:"status" gorm:"size:20;not null;default:'active'"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	User       *User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Logs       []GoalLog   `json:"logs,omitempty" gorm:"foreignKey:GoalID"`
	Milestones []Milestone `json:"milestones,omitempty" gorm:"foreignKey:GoalID"`
}

func (g *Goal) Progress() float64 {
	total := g.TargetValue - g.StartValue
	if total == 0 {
		return 100
	}
	current := g.CurrentValue - g.StartValue
	if g.Direction == "decrease" {
		current = g.StartValue - g.CurrentValue
		total = g.StartValue - g.TargetValue
	}
	if total == 0 {
		return 100
	}
	p := current / total * 100
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

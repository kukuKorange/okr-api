package model

import "time"

type ActivityLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	FamilyID   *uint     `json:"family_id" gorm:"index"`
	Action     string    `json:"action" gorm:"size:50;not null"`
	EntityType string    `json:"entity_type" gorm:"size:50"`
	EntityID   uint      `json:"entity_id"`
	Summary    string    `json:"summary" gorm:"size:500"`
	CreatedAt  time.Time `json:"created_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

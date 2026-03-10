package model

import "time"

type Family struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CreatorID  uint      `json:"creator_id" gorm:"not null;index"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	InviteCode string    `json:"invite_code" gorm:"size:10;uniqueIndex"`
	CreatedAt  time.Time `json:"created_at"`

	Creator *User          `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Members []FamilyMember `json:"members,omitempty" gorm:"foreignKey:FamilyID"`
}

type FamilyMember struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FamilyID    uint      `json:"family_id" gorm:"not null;uniqueIndex:idx_family_user"`
	UserID      uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_family_user"`
	Role        string    `json:"role" gorm:"size:20;not null;default:'member'"`
	DisplayName string    `json:"display_name" gorm:"size:50"`
	JoinedAt    time.Time `json:"joined_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

package model

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Nickname     string    `json:"nickname" gorm:"size:50;not null"`
	Email        string    `json:"email" gorm:"size:100;uniqueIndex"`
	Phone        string    `json:"phone" gorm:"size:20;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	AvatarURL    string    `json:"avatar_url" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

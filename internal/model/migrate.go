package model

import (
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&User{},
		&Family{},
		&FamilyMember{},
		&Goal{},
		&GoalLog{},
		&Milestone{},
		&TodoItem{},
		&Habit{},
		&HabitCheckIn{},
		&ActivityLog{},
		&Project{},
		&ProjectPhase{},
		&ProjectTask{},
	)
	if err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}
	log.Println("database migration completed")
}

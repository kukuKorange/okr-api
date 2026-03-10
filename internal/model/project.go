package model

import "time"

type Project struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	Title       string     `json:"title" gorm:"size:200;not null"`
	Description string     `json:"description" gorm:"size:2000"`
	Status      string     `json:"status" gorm:"size:20;not null;default:'planning'"`
	Priority    int        `json:"priority" gorm:"default:2"`
	Color       string     `json:"color" gorm:"size:20;default:'#4f46e5'"`
	Icon        string     `json:"icon" gorm:"size:50;default:'📋'"`
	StartDate   *time.Time `json:"start_date" gorm:"type:date"`
	Deadline    *time.Time `json:"deadline" gorm:"type:date"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Tasks  []ProjectTask  `json:"tasks,omitempty" gorm:"foreignKey:ProjectID"`
	Phases []ProjectPhase `json:"phases,omitempty" gorm:"foreignKey:ProjectID"`
}

// Progress returns completion percentage based on tasks
func (p *Project) Progress() float64 {
	if len(p.Tasks) == 0 {
		return 0
	}
	done := 0
	for _, t := range p.Tasks {
		if t.Status == "done" {
			done++
		}
	}
	return float64(done) / float64(len(p.Tasks)) * 100
}

func (p *Project) TaskStats() map[string]int {
	stats := map[string]int{"todo": 0, "in_progress": 0, "done": 0, "total": 0}
	for _, t := range p.Tasks {
		stats[t.Status]++
		stats["total"]++
	}
	return stats
}

type ProjectPhase struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"project_id" gorm:"not null;index"`
	Title     string    `json:"title" gorm:"size:100;not null"`
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`

	Tasks []ProjectTask `json:"tasks,omitempty" gorm:"foreignKey:PhaseID"`
}

type ProjectTask struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	ProjectID      uint       `json:"project_id" gorm:"not null;index"`
	PhaseID        *uint      `json:"phase_id" gorm:"index"`
	UserID         uint       `json:"user_id" gorm:"not null;index"`
	Title          string     `json:"title" gorm:"size:300;not null"`
	Description    string     `json:"description" gorm:"size:2000"`
	Status         string     `json:"status" gorm:"size:20;not null;default:'todo'"`
	Priority       int        `json:"priority" gorm:"default:2"`
	DueDate        *time.Time `json:"due_date" gorm:"type:date"`
	EstimatedHours float64    `json:"estimated_hours" gorm:"default:0"`
	ActualHours    float64    `json:"actual_hours" gorm:"default:0"`
	SortOrder      int        `json:"sort_order" gorm:"default:0"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

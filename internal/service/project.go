package service

import (
	"errors"
	"time"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"

	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// --- Project CRUD ---

func (s *ProjectService) Create(userID uint, p *model.Project) error {
	p.UserID = userID
	return s.db.Create(p).Error
}

func (s *ProjectService) List(userID uint, status string, page, size int) ([]map[string]interface{}, int64, error) {
	var projects []model.Project
	var total int64

	q := s.db.Scopes(middleware.TenantScope(userID)).Preload("Tasks")
	if status != "" {
		q = q.Where("status = ?", status)
	}

	q.Model(&model.Project{}).Count(&total)
	err := q.Order("CASE status WHEN 'in_progress' THEN 0 WHEN 'planning' THEN 1 WHEN 'completed' THEN 2 ELSE 3 END, updated_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(projects))
	for _, p := range projects {
		stats := p.TaskStats()
		result = append(result, map[string]interface{}{
			"id":          p.ID,
			"title":       p.Title,
			"description": p.Description,
			"status":      p.Status,
			"priority":    p.Priority,
			"color":       p.Color,
			"icon":        p.Icon,
			"start_date":  p.StartDate,
			"deadline":    p.Deadline,
			"progress":    p.Progress(),
			"task_stats":  stats,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		})
	}
	return result, total, nil
}

func (s *ProjectService) GetByID(userID, projectID uint) (*model.Project, error) {
	var p model.Project
	err := s.db.Preload("Tasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, created_at ASC")
	}).Preload("Phases", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&p, projectID).Error
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, errors.New("无权访问")
	}
	return &p, nil
}

func (s *ProjectService) Update(userID, projectID uint, updates map[string]interface{}) error {
	var p model.Project
	if err := s.db.First(&p, projectID).Error; err != nil {
		return err
	}
	if p.UserID != userID {
		return errors.New("无权修改")
	}
	return s.db.Model(&model.Project{}).Where("id = ?", projectID).Updates(updates).Error
}

func (s *ProjectService) Delete(userID, projectID uint) error {
	return s.Update(userID, projectID, map[string]interface{}{"status": "archived"})
}

// --- Phase CRUD ---

func (s *ProjectService) CreatePhase(userID, projectID uint, phase *model.ProjectPhase) error {
	if _, err := s.GetByID(userID, projectID); err != nil {
		return err
	}
	phase.ProjectID = projectID
	return s.db.Create(phase).Error
}

func (s *ProjectService) UpdatePhase(userID, phaseID uint, updates map[string]interface{}) error {
	var phase model.ProjectPhase
	if err := s.db.First(&phase, phaseID).Error; err != nil {
		return err
	}
	if _, err := s.GetByID(userID, phase.ProjectID); err != nil {
		return err
	}
	return s.db.Model(&model.ProjectPhase{}).Where("id = ?", phaseID).Updates(updates).Error
}

func (s *ProjectService) DeletePhase(userID, phaseID uint) error {
	var phase model.ProjectPhase
	if err := s.db.First(&phase, phaseID).Error; err != nil {
		return err
	}
	if _, err := s.GetByID(userID, phase.ProjectID); err != nil {
		return err
	}
	// Move tasks in this phase to no-phase
	s.db.Model(&model.ProjectTask{}).Where("phase_id = ?", phaseID).Update("phase_id", nil)
	return s.db.Delete(&phase).Error
}

// --- Task CRUD ---

func (s *ProjectService) CreateTask(userID, projectID uint, task *model.ProjectTask) error {
	if _, err := s.GetByID(userID, projectID); err != nil {
		return err
	}
	task.ProjectID = projectID
	task.UserID = userID
	return s.db.Create(task).Error
}

func (s *ProjectService) UpdateTask(userID, taskID uint, updates map[string]interface{}) error {
	var task model.ProjectTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return err
	}
	if _, err := s.GetByID(userID, task.ProjectID); err != nil {
		return err
	}

	// Handle status -> done transition
	if newStatus, ok := updates["status"]; ok && newStatus == "done" && task.Status != "done" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	if newStatus, ok := updates["status"]; ok && newStatus != "done" {
		updates["completed_at"] = nil
	}

	if err := s.db.Model(&model.ProjectTask{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return err
	}

	// Auto-check project completion
	s.checkProjectCompletion(task.ProjectID)
	return nil
}

func (s *ProjectService) DeleteTask(userID, taskID uint) error {
	var task model.ProjectTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return err
	}
	if _, err := s.GetByID(userID, task.ProjectID); err != nil {
		return err
	}
	return s.db.Delete(&task).Error
}

func (s *ProjectService) ReorderTasks(userID, projectID uint, taskIDs []uint) error {
	if _, err := s.GetByID(userID, projectID); err != nil {
		return err
	}
	for i, id := range taskIDs {
		s.db.Model(&model.ProjectTask{}).Where("id = ? AND project_id = ?", id, projectID).Update("sort_order", i)
	}
	return nil
}

// MoveTask changes a task's status and optionally its phase (kanban drag-drop)
func (s *ProjectService) MoveTask(userID, taskID uint, status string, phaseID *uint) error {
	updates := map[string]interface{}{"status": status}
	if phaseID != nil {
		updates["phase_id"] = phaseID
	}
	return s.UpdateTask(userID, taskID, updates)
}

func (s *ProjectService) checkProjectCompletion(projectID uint) {
	var total, done int64
	s.db.Model(&model.ProjectTask{}).Where("project_id = ?", projectID).Count(&total)
	s.db.Model(&model.ProjectTask{}).Where("project_id = ? AND status = ?", projectID, "done").Count(&done)
	if total > 0 && total == done {
		now := time.Now()
		s.db.Model(&model.Project{}).Where("id = ?", projectID).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": &now,
		})
	}
}

// GetOverdueTasks returns tasks past due date
func (s *ProjectService) GetOverdueTasks(userID uint) ([]map[string]interface{}, error) {
	var tasks []model.ProjectTask
	today := time.Now().Truncate(24 * time.Hour)
	err := s.db.Where("user_id = ? AND status != ? AND due_date < ?", userID, "done", today).
		Order("due_date ASC").Limit(20).Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, t := range tasks {
		var proj model.Project
		s.db.Select("id, title, icon").First(&proj, t.ProjectID)
		result = append(result, map[string]interface{}{
			"task":         t,
			"project_id":   proj.ID,
			"project_title": proj.Title,
			"project_icon":  proj.Icon,
		})
	}
	return result, nil
}

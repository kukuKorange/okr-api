package service

import (
	"errors"
	"time"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"

	"gorm.io/gorm"
)

type TodoService struct {
	db *gorm.DB
}

func NewTodoService(db *gorm.DB) *TodoService {
	return &TodoService{db: db}
}

func (s *TodoService) Create(userID uint, todo *model.TodoItem) error {
	todo.UserID = userID
	return s.db.Create(todo).Error
}

func (s *TodoService) List(userID uint, filter string) ([]model.TodoItem, error) {
	var todos []model.TodoItem
	q := s.db.Scopes(middleware.TenantScope(userID)).Preload("Goal")

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)

	switch filter {
	case "today":
		q = q.Where("(due_date >= ? AND due_date < ?) OR (due_date IS NULL AND is_completed = ?)", today, tomorrow, false)
	case "overdue":
		q = q.Where("due_date < ? AND is_completed = ?", today, false)
	case "completed":
		q = q.Where("is_completed = ?", true)
	default:
		q = q.Where("is_completed = ?", false)
	}

	err := q.Order("priority DESC, sort_order ASC, created_at DESC").Find(&todos).Error
	return todos, err
}

func (s *TodoService) Update(userID, todoID uint, updates map[string]interface{}) error {
	result := s.db.Model(&model.TodoItem{}).
		Where("id = ? AND user_id = ?", todoID, userID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return errors.New("待办不存在")
	}
	return result.Error
}

func (s *TodoService) Complete(userID, todoID uint) error {
	now := time.Now()
	return s.Update(userID, todoID, map[string]interface{}{
		"is_completed": true,
		"completed_at": &now,
	})
}

func (s *TodoService) Delete(userID, todoID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", todoID, userID).Delete(&model.TodoItem{})
	if result.RowsAffected == 0 {
		return errors.New("待办不存在")
	}
	return result.Error
}

func (s *TodoService) Reorder(userID uint, ids []uint) error {
	for i, id := range ids {
		s.db.Model(&model.TodoItem{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("sort_order", i)
	}
	return nil
}

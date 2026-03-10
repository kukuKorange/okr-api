package service

import (
	"errors"
	"math"
	"time"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"

	"gorm.io/gorm"
)

type GoalService struct {
	db *gorm.DB
}

func NewGoalService(db *gorm.DB) *GoalService {
	return &GoalService{db: db}
}

func (s *GoalService) Create(userID uint, goal *model.Goal) error {
	goal.UserID = userID
	if goal.StartDate.IsZero() {
		goal.StartDate = time.Now()
	}
	goal.CurrentValue = goal.StartValue
	return s.db.Create(goal).Error
}

func (s *GoalService) List(userID uint, status string, page, size int) ([]model.Goal, int64, error) {
	var goals []model.Goal
	var total int64

	q := s.db.Scopes(middleware.TenantScope(userID))

	// also include family shared goals
	var familyIDs []uint
	s.db.Model(&model.FamilyMember{}).Where("user_id = ?", userID).Pluck("family_id", &familyIDs)
	if len(familyIDs) > 0 {
		q = s.db.Where("user_id = ? OR family_id IN ?", userID, familyIDs)
	}

	if status != "" {
		q = q.Where("status = ?", status)
	}

	q.Model(&model.Goal{}).Count(&total)
	err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&goals).Error
	return goals, total, err
}

func (s *GoalService) GetByID(userID, goalID uint) (*model.Goal, error) {
	var goal model.Goal
	err := s.db.Preload("Milestones", func(db *gorm.DB) *gorm.DB {
		return db.Order("target_date ASC")
	}).First(&goal, goalID).Error
	if err != nil {
		return nil, err
	}
	if !s.canAccess(userID, &goal) {
		return nil, errors.New("无权访问")
	}
	return &goal, nil
}

func (s *GoalService) Update(userID, goalID uint, updates map[string]interface{}) error {
	goal, err := s.GetByID(userID, goalID)
	if err != nil {
		return err
	}
	if goal.UserID != userID {
		return errors.New("无权修改")
	}
	return s.db.Model(&model.Goal{}).Where("id = ?", goalID).Updates(updates).Error
}

func (s *GoalService) Archive(userID, goalID uint) error {
	return s.Update(userID, goalID, map[string]interface{}{"status": "archived"})
}

func (s *GoalService) AddLog(userID, goalID uint, log *model.GoalLog) error {
	goal, err := s.GetByID(userID, goalID)
	if err != nil {
		return err
	}

	log.GoalID = goalID
	log.UserID = userID
	if log.LogDate.IsZero() {
		log.LogDate = time.Now()
	}

	if err := s.db.Create(log).Error; err != nil {
		return err
	}

	// Update goal current_value
	s.db.Model(&model.Goal{}).Where("id = ?", goalID).Update("current_value", log.Value)

	// Check if goal completed
	completed := false
	if goal.Direction == "increase" && log.Value >= goal.TargetValue {
		completed = true
	} else if goal.Direction == "decrease" && log.Value <= goal.TargetValue {
		completed = true
	}
	if completed {
		now := time.Now()
		s.db.Model(&model.Goal{}).Where("id = ?", goalID).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": &now,
		})
	}

	// Check milestones
	s.checkMilestones(goal, log.Value)

	return nil
}

func (s *GoalService) checkMilestones(goal *model.Goal, currentValue float64) {
	var milestones []model.Milestone
	s.db.Where("goal_id = ? AND completed_at IS NULL", goal.ID).Find(&milestones)
	now := time.Now()
	for _, m := range milestones {
		reached := false
		if goal.Direction == "increase" && currentValue >= m.TargetValue {
			reached = true
		} else if goal.Direction == "decrease" && currentValue <= m.TargetValue {
			reached = true
		}
		if reached {
			s.db.Model(&model.Milestone{}).Where("id = ?", m.ID).Update("completed_at", &now)
		}
	}
}

func (s *GoalService) GetLogs(goalID uint, page, size int) ([]model.GoalLog, int64, error) {
	var logs []model.GoalLog
	var total int64
	q := s.db.Where("goal_id = ?", goalID)
	q.Model(&model.GoalLog{}).Count(&total)
	err := q.Preload("User").Order("log_date DESC").Offset((page - 1) * size).Limit(size).Find(&logs).Error
	return logs, total, err
}

func (s *GoalService) GetTrend(goalID uint) ([]model.GoalLog, error) {
	var logs []model.GoalLog
	err := s.db.Where("goal_id = ?", goalID).Order("log_date ASC").Find(&logs).Error
	return logs, err
}

func (s *GoalService) PredictCompletion(goalID uint) (map[string]interface{}, error) {
	var logs []model.GoalLog
	s.db.Where("goal_id = ?", goalID).Order("log_date DESC").Limit(14).Find(&logs)
	if len(logs) < 3 {
		return nil, errors.New("数据不足，至少需要3天记录")
	}

	var goal model.Goal
	if err := s.db.First(&goal, goalID).Error; err != nil {
		return nil, err
	}

	// Simple linear regression on recent data
	n := float64(len(logs))
	var sumX, sumY, sumXY, sumX2 float64
	baseTime := logs[len(logs)-1].LogDate
	for i, l := range logs {
		x := l.LogDate.Sub(baseTime).Hours() / 24
		y := l.Value
		_ = i
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return nil, errors.New("无法预测")
	}
	slope := (n*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / n

	if slope == 0 {
		return nil, errors.New("进度无变化，无法预测")
	}

	// Predict when target_value will be reached
	targetX := (goal.TargetValue - intercept) / slope
	lastX := logs[0].LogDate.Sub(baseTime).Hours() / 24
	daysRemaining := targetX - lastX

	if daysRemaining < 0 {
		daysRemaining = 0
	}

	predictedDate := time.Now().AddDate(0, 0, int(math.Ceil(daysRemaining)))

	return map[string]interface{}{
		"predicted_date":   predictedDate.Format("2006-01-02"),
		"days_remaining":   int(math.Ceil(daysRemaining)),
		"daily_rate":       slope,
		"confidence":       "low",
		"data_points_used": len(logs),
	}, nil
}

// CreateMilestone creates a new milestone for a goal
func (s *GoalService) CreateMilestone(userID, goalID uint, m *model.Milestone) error {
	if _, err := s.GetByID(userID, goalID); err != nil {
		return err
	}
	m.GoalID = goalID
	return s.db.Create(m).Error
}

func (s *GoalService) GetMilestones(goalID uint) ([]model.Milestone, error) {
	var milestones []model.Milestone
	err := s.db.Where("goal_id = ?", goalID).Order("target_date ASC").Find(&milestones).Error
	return milestones, err
}

func (s *GoalService) canAccess(userID uint, goal *model.Goal) bool {
	if goal.UserID == userID {
		return true
	}
	if goal.FamilyID != nil {
		var count int64
		s.db.Model(&model.FamilyMember{}).
			Where("family_id = ? AND user_id = ?", *goal.FamilyID, userID).
			Count(&count)
		return count > 0
	}
	return false
}

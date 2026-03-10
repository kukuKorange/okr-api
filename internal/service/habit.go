package service

import (
	"errors"
	"time"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"

	"gorm.io/gorm"
)

type HabitService struct {
	db *gorm.DB
}

func NewHabitService(db *gorm.DB) *HabitService {
	return &HabitService{db: db}
}

func (s *HabitService) Create(userID uint, habit *model.Habit) error {
	habit.UserID = userID
	if habit.StartDate.IsZero() {
		habit.StartDate = time.Now()
	}
	return s.db.Create(habit).Error
}

func (s *HabitService) List(userID uint) ([]model.Habit, error) {
	var habits []model.Habit
	err := s.db.Scopes(middleware.TenantScope(userID)).
		Where("is_archived = ?", false).
		Order("created_at ASC").
		Find(&habits).Error
	return habits, err
}

func (s *HabitService) Update(userID, habitID uint, updates map[string]interface{}) error {
	result := s.db.Model(&model.Habit{}).
		Where("id = ? AND user_id = ?", habitID, userID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return errors.New("习惯不存在")
	}
	return result.Error
}

func (s *HabitService) Archive(userID, habitID uint) error {
	return s.Update(userID, habitID, map[string]interface{}{"is_archived": true})
}

func (s *HabitService) CheckIn(userID, habitID uint, date time.Time, note string) (*model.HabitCheckIn, error) {
	var habit model.Habit
	if err := s.db.Where("id = ? AND user_id = ?", habitID, userID).First(&habit).Error; err != nil {
		return nil, errors.New("习惯不存在")
	}

	checkDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Check if already checked in today
	var existing model.HabitCheckIn
	err := s.db.Where("habit_id = ? AND check_date = ?", habitID, checkDate).First(&existing).Error
	if err == nil {
		return nil, errors.New("今天已经打过卡了")
	}

	checkIn := &model.HabitCheckIn{
		HabitID:   habitID,
		UserID:    userID,
		CheckDate: checkDate,
		Note:      note,
	}
	if err := s.db.Create(checkIn).Error; err != nil {
		return nil, err
	}
	return checkIn, nil
}

func (s *HabitService) UndoCheckIn(userID, habitID uint) error {
	today := time.Now().Truncate(24 * time.Hour)
	result := s.db.Where("habit_id = ? AND user_id = ? AND check_date = ?", habitID, userID, today).
		Delete(&model.HabitCheckIn{})
	if result.RowsAffected == 0 {
		return errors.New("今天未打卡")
	}
	return result.Error
}

func (s *HabitService) GetCalendar(habitID uint, year, month int) ([]map[string]interface{}, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	var checkIns []model.HabitCheckIn
	err := s.db.Where("habit_id = ? AND check_date >= ? AND check_date < ?", habitID, start, end).
		Order("check_date ASC").
		Find(&checkIns).Error
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, ci := range checkIns {
		result = append(result, map[string]interface{}{
			"date": ci.CheckDate.Format("2006-01-02"),
			"note": ci.Note,
		})
	}
	return result, nil
}

func (s *HabitService) GetTodayHabits(userID uint) ([]map[string]interface{}, error) {
	habits, err := s.List(userID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	var result []map[string]interface{}

	for _, h := range habits {
		checked := false
		var checkIn model.HabitCheckIn
		if err := s.db.Where("habit_id = ? AND check_date = ?", h.ID, today).First(&checkIn).Error; err == nil {
			checked = true
		}

		streak := s.calculateStreak(h.ID)

		result = append(result, map[string]interface{}{
			"habit":   h,
			"checked": checked,
			"streak":  streak,
		})
	}
	return result, nil
}

func (s *HabitService) calculateStreak(habitID uint) int {
	var checkIns []model.HabitCheckIn
	s.db.Where("habit_id = ?", habitID).Order("check_date DESC").Limit(365).Find(&checkIns)

	if len(checkIns) == 0 {
		return 0
	}

	streak := 0
	today := time.Now().Truncate(24 * time.Hour)
	expected := today

	// Allow streak to start from today or yesterday
	firstDate := checkIns[0].CheckDate.Truncate(24 * time.Hour)
	if firstDate.Equal(today) {
		expected = today
	} else if firstDate.Equal(today.AddDate(0, 0, -1)) {
		expected = today.AddDate(0, 0, -1)
	} else {
		return 0
	}

	for _, ci := range checkIns {
		d := ci.CheckDate.Truncate(24 * time.Hour)
		if d.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if d.Before(expected) {
			break
		}
	}
	return streak
}

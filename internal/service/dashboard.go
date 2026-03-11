package service

import (
	"time"

	"goaltrack/internal/model"

	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

func (s *DashboardService) GetDashboard(userID uint) (map[string]interface{}, error) {
	today := time.Now().Truncate(24 * time.Hour)

	// Active goals (top 3)
	var goals []model.Goal
	s.db.Where("user_id = ? AND status = ?", userID, "active").
		Order("deadline ASC").Limit(3).Find(&goals)

	goalCards := make([]map[string]interface{}, 0)
	for _, g := range goals {
		goalCards = append(goalCards, map[string]interface{}{
			"id":            g.ID,
			"title":         g.Title,
			"category":      g.Category,
			"progress":      g.Progress(),
			"current_value": g.CurrentValue,
			"target_value":  g.TargetValue,
			"unit":          g.Unit,
			"deadline":      g.Deadline.Format("2006-01-02"),
			"family_id":     g.FamilyID,
		})
	}

	// Today's habits
	var habits []model.Habit
	s.db.Where("user_id = ? AND is_archived = ?", userID, false).Find(&habits)

	totalHabits := len(habits)
	checkedCount := 0
	habitItems := make([]map[string]interface{}, 0)

	for _, h := range habits {
		var count int64
		s.db.Model(&model.HabitCheckIn{}).
			Where("habit_id = ? AND check_date = ?", h.ID, today).
			Count(&count)
		checked := count > 0
		if checked {
			checkedCount++
		}
		habitItems = append(habitItems, map[string]interface{}{
			"id":      h.ID,
			"name":    h.Name,
			"icon":    h.Icon,
			"color":   h.Color,
			"checked": checked,
		})
	}

	// Today's todos count
	var todoCount int64
	s.db.Model(&model.TodoItem{}).
		Where("user_id = ? AND is_completed = ? AND (due_date IS NULL OR due_date >= ?)", userID, false, today).
		Count(&todoCount)

	var overdueCount int64
	s.db.Model(&model.TodoItem{}).
		Where("user_id = ? AND is_completed = ? AND due_date < ?", userID, false, today).
		Count(&overdueCount)

	// All user projects for summary
	var allProjects []model.Project
	s.db.Preload("Tasks").Where("user_id = ?", userID).Find(&allProjects)

	statusCounts := map[string]int{
		"to_discuss":  0,
		"pending":     0,
		"planning":    0,
		"in_progress": 0,
		"completed":   0,
	}
	totalTasks := 0
	totalTasksDone := 0
	totalTasksInProgress := 0
	overdueProjects := 0
	overdueTasksCount := 0

	projectMap := make(map[uint]model.Project)
	for _, p := range allProjects {
		projectMap[p.ID] = p
		if _, ok := statusCounts[p.Status]; ok {
			statusCounts[p.Status]++
		}
		for _, t := range p.Tasks {
			totalTasks++
			if t.Status == "done" {
				totalTasksDone++
			} else if t.Status == "in_progress" {
				totalTasksInProgress++
			}
			if t.Status != "done" && t.DueDate != nil && t.DueDate.Before(today) {
				overdueTasksCount++
			}
		}
		if p.Deadline != nil && p.Status != "completed" && p.Deadline.Before(today) {
			overdueProjects++
		}
	}

	overallRate := 0.0
	if totalTasks > 0 {
		overallRate = float64(totalTasksDone) / float64(totalTasks) * 100
	}

	projectSummary := map[string]interface{}{
		"total_projects":      len(allProjects),
		"status_counts":       statusCounts,
		"total_tasks":         totalTasks,
		"tasks_done":          totalTasksDone,
		"tasks_in_progress":   totalTasksInProgress,
		"tasks_todo":          totalTasks - totalTasksDone - totalTasksInProgress,
		"overall_rate":        overallRate,
		"overdue_projects":    overdueProjects,
		"overdue_tasks_count": overdueTasksCount,
	}

	// Upcoming tasks: non-done tasks sorted by urgency (overdue first, then soonest deadline, then highest priority)
	var upcomingRaw []model.ProjectTask
	todayStr := today.Format("2006-01-02")
	s.db.Raw(`SELECT * FROM project_tasks WHERE user_id = ? AND status != 'done'
		ORDER BY priority DESC, CASE WHEN due_date < ? THEN 0 ELSE 1 END, due_date ASC NULLS LAST
		LIMIT 20`, userID, todayStr).Scan(&upcomingRaw)

	upcomingTasks := make([]map[string]interface{}, 0, len(upcomingRaw))
	for _, t := range upcomingRaw {
		item := map[string]interface{}{
			"id":         t.ID,
			"title":      t.Title,
			"status":     t.Status,
			"priority":   t.Priority,
			"due_date":   t.DueDate,
			"project_id": t.ProjectID,
		}
		if p, ok := projectMap[t.ProjectID]; ok {
			item["project_title"] = p.Title
			item["project_icon"] = p.Icon
			item["project_color"] = p.Color
		}
		upcomingTasks = append(upcomingTasks, item)
	}

	// Active projects (top 5) for detail cards
	projectCards := make([]map[string]interface{}, 0)
	for _, p := range allProjects {
		if p.Status == "completed" || p.Status == "archived" {
			continue
		}
		stats := p.TaskStats()
		projectCards = append(projectCards, map[string]interface{}{
			"id":         p.ID,
			"title":      p.Title,
			"icon":       p.Icon,
			"color":      p.Color,
			"status":     p.Status,
			"priority":   p.Priority,
			"progress":   p.Progress(),
			"task_stats": stats,
			"start_date": p.StartDate,
			"deadline":   p.Deadline,
		})
		if len(projectCards) >= 5 {
			break
		}
	}

	// Recent activities
	var activities []model.ActivityLog
	familyIDs := s.getUserFamilyIDs(userID)
	aq := s.db.Where("user_id = ?", userID)
	if len(familyIDs) > 0 {
		aq = s.db.Where("user_id = ? OR family_id IN ?", userID, familyIDs)
	}
	aq.Preload("User").Order("created_at DESC").Limit(5).Find(&activities)

	return map[string]interface{}{
		"goals":         goalCards,
		"habits":        habitItems,
		"habits_total":  totalHabits,
		"habits_done":   checkedCount,
		"todo_count":    todoCount,
		"overdue_count": overdueCount,
		"projects":         projectCards,
		"project_summary":  projectSummary,
		"upcoming_tasks":   upcomingTasks,
		"activities":    activities,
	}, nil
}

func (s *DashboardService) GetActivities(userID uint, page, size int) ([]model.ActivityLog, int64, error) {
	var activities []model.ActivityLog
	var total int64

	familyIDs := s.getUserFamilyIDs(userID)
	q := s.db.Where("user_id = ?", userID)
	if len(familyIDs) > 0 {
		q = s.db.Where("user_id = ? OR family_id IN ?", userID, familyIDs)
	}

	q.Model(&model.ActivityLog{}).Count(&total)
	err := q.Preload("User").Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&activities).Error
	return activities, total, err
}

func (s *DashboardService) LogActivity(userID uint, familyID *uint, action, entityType string, entityID uint, summary string) {
	log := &model.ActivityLog{
		UserID:     userID,
		FamilyID:   familyID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Summary:    summary,
	}
	s.db.Create(log)
}

func (s *DashboardService) getUserFamilyIDs(userID uint) []uint {
	var ids []uint
	s.db.Model(&model.FamilyMember{}).Where("user_id = ?", userID).Pluck("family_id", &ids)
	return ids
}

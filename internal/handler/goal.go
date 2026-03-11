package handler

import (
	"strconv"
	"time"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"
	"goaltrack/internal/service"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

func parseDateRequired(s string) time.Time {
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Now()
}

type GoalHandler struct {
	svc      *service.GoalService
	dashSvc  *service.DashboardService
}

func NewGoalHandler(svc *service.GoalService, dashSvc *service.DashboardService) *GoalHandler {
	return &GoalHandler{svc: svc, dashSvc: dashSvc}
}

func (h *GoalHandler) Create(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		Title       string  `json:"title" binding:"required"`
		Category    string  `json:"category"`
		Unit        string  `json:"unit"`
		TargetValue float64 `json:"target_value"`
		StartValue  float64 `json:"start_value"`
		Direction   string  `json:"direction"`
		StartDate   string  `json:"start_date"`
		Deadline    string  `json:"deadline"`
		FamilyID    *uint   `json:"family_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	goal := model.Goal{
		Title:       req.Title,
		Category:    req.Category,
		Unit:        req.Unit,
		TargetValue: req.TargetValue,
		StartValue:  req.StartValue,
		Direction:   req.Direction,
		StartDate:   parseDateRequired(req.StartDate),
		Deadline:    parseDateRequired(req.Deadline),
		FamilyID:    req.FamilyID,
	}
	if goal.Category == "" {
		goal.Category = "custom"
	}
	if goal.Direction == "" {
		goal.Direction = "increase"
	}
	if err := h.svc.Create(uid, &goal); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	h.dashSvc.LogActivity(uid, goal.FamilyID, "goal_created", "goal", goal.ID, "创建了目标: "+goal.Title)
	response.OK(c, goal)
}

func (h *GoalHandler) List(c *gin.Context) {
	uid := middleware.GetUserID(c)
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	goals, total, err := h.svc.List(uid, status, page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	goalList := make([]map[string]interface{}, 0)
	for _, g := range goals {
		goalList = append(goalList, map[string]interface{}{
			"id":            g.ID,
			"title":         g.Title,
			"category":      g.Category,
			"unit":          g.Unit,
			"target_value":  g.TargetValue,
			"current_value": g.CurrentValue,
			"start_value":   g.StartValue,
			"direction":     g.Direction,
			"progress":      g.Progress(),
			"deadline":      g.Deadline.Format("2006-01-02"),
			"status":        g.Status,
			"family_id":     g.FamilyID,
			"created_at":    g.CreatedAt,
		})
	}

	response.OK(c, response.PageData{List: goalList, Total: total, Page: page, Size: size})
}

func (h *GoalHandler) GetByID(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	goal, err := h.svc.GetByID(uid, uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	data := map[string]interface{}{
		"id":            goal.ID,
		"user_id":       goal.UserID,
		"family_id":     goal.FamilyID,
		"title":         goal.Title,
		"category":      goal.Category,
		"unit":          goal.Unit,
		"target_value":  goal.TargetValue,
		"current_value": goal.CurrentValue,
		"start_value":   goal.StartValue,
		"direction":     goal.Direction,
		"progress":      goal.Progress(),
		"start_date":    goal.StartDate.Format("2006-01-02"),
		"deadline":      goal.Deadline.Format("2006-01-02"),
		"status":        goal.Status,
		"completed_at":  goal.CompletedAt,
		"milestones":    goal.Milestones,
		"created_at":    goal.CreatedAt,
	}
	response.OK(c, data)
}

func (h *GoalHandler) Update(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	delete(updates, "id")
	delete(updates, "user_id")
	if err := h.svc.Update(uid, uint(id), updates); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *GoalHandler) Archive(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Archive(uid, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已归档")
}

func (h *GoalHandler) AddLog(c *gin.Context) {
	uid := middleware.GetUserID(c)
	goalID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		LogDate      string  `json:"log_date"`
		Value        float64 `json:"value"`
		Note         string  `json:"note"`
		PhotoURL     string  `json:"photo_url"`
		PlanTomorrow string  `json:"plan_tomorrow"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	logDate := parseDateRequired(req.LogDate)
	log := model.GoalLog{
		LogDate:      logDate,
		Value:        req.Value,
		Note:         req.Note,
		PhotoURL:     req.PhotoURL,
		PlanTomorrow: req.PlanTomorrow,
	}
	if err := h.svc.AddLog(uid, uint(goalID), &log); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, log)
}

func (h *GoalHandler) GetLogs(c *gin.Context) {
	goalID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	logs, total, err := h.svc.GetLogs(uint(goalID), page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, response.PageData{List: logs, Total: total, Page: page, Size: size})
}

func (h *GoalHandler) GetTrend(c *gin.Context) {
	goalID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	logs, err := h.svc.GetTrend(uint(goalID))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, logs)
}

func (h *GoalHandler) Predict(c *gin.Context) {
	goalID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := h.svc.PredictCompletion(uint(goalID))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *GoalHandler) CreateMilestone(c *gin.Context) {
	uid := middleware.GetUserID(c)
	goalID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Title       string  `json:"title" binding:"required"`
		TargetValue float64 `json:"target_value"`
		TargetDate  string  `json:"target_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	m := model.Milestone{
		Title:       req.Title,
		TargetValue: req.TargetValue,
		TargetDate:  parseDateRequired(req.TargetDate),
	}
	if err := h.svc.CreateMilestone(uid, uint(goalID), &m); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, m)
}

func (h *GoalHandler) GetMilestones(c *gin.Context) {
	goalID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	milestones, err := h.svc.GetMilestones(uint(goalID))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, milestones)
}

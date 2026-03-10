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

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

type ProjectHandler struct {
	svc     *service.ProjectService
	dashSvc *service.DashboardService
}

func NewProjectHandler(svc *service.ProjectService, dashSvc *service.DashboardService) *ProjectHandler {
	return &ProjectHandler{svc: svc, dashSvc: dashSvc}
}

// --- Project ---

func (h *ProjectHandler) Create(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    int    `json:"priority"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
		StartDate   string `json:"start_date"`
		Deadline    string `json:"deadline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p := model.Project{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		Color:       req.Color,
		Icon:        req.Icon,
		StartDate:   parseDate(req.StartDate),
		Deadline:    parseDate(req.Deadline),
	}
	if p.Status == "" {
		p.Status = "planning"
	}
	if p.Color == "" {
		p.Color = "#4f46e5"
	}
	if p.Icon == "" {
		p.Icon = "📋"
	}
	if err := h.svc.Create(uid, &p); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	h.dashSvc.LogActivity(uid, nil, "project_created", "project", p.ID, "创建了项目: "+p.Title)
	response.OK(c, p)
}

func (h *ProjectHandler) List(c *gin.Context) {
	uid := middleware.GetUserID(c)
	status := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := h.svc.List(uid, status, page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, response.PageData{List: list, Total: total, Page: page, Size: size})
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.svc.GetByID(uid, uint(id))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	data := map[string]interface{}{
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
		"task_stats":  p.TaskStats(),
		"tasks":       p.Tasks,
		"phases":      p.Phases,
		"completed_at": p.CompletedAt,
		"created_at":  p.CreatedAt,
	}
	response.OK(c, data)
}

func (h *ProjectHandler) Update(c *gin.Context) {
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

func (h *ProjectHandler) Archive(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uid, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已归档")
}

// --- Phase ---

func (h *ProjectHandler) CreatePhase(c *gin.Context) {
	uid := middleware.GetUserID(c)
	pid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var phase model.ProjectPhase
	if err := c.ShouldBindJSON(&phase); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CreatePhase(uid, uint(pid), &phase); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, phase)
}

func (h *ProjectHandler) UpdatePhase(c *gin.Context) {
	uid := middleware.GetUserID(c)
	phaseID, _ := strconv.ParseUint(c.Param("phaseId"), 10, 64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdatePhase(uid, uint(phaseID), updates); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *ProjectHandler) DeletePhase(c *gin.Context) {
	uid := middleware.GetUserID(c)
	phaseID, _ := strconv.ParseUint(c.Param("phaseId"), 10, 64)
	if err := h.svc.DeletePhase(uid, uint(phaseID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已删除")
}

// --- Task ---

func (h *ProjectHandler) CreateTask(c *gin.Context) {
	uid := middleware.GetUserID(c)
	pid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Title          string  `json:"title" binding:"required"`
		Description    string  `json:"description"`
		PhaseID        *uint   `json:"phase_id"`
		Priority       int     `json:"priority"`
		DueDate        string  `json:"due_date"`
		EstimatedHours float64 `json:"estimated_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	task := model.ProjectTask{
		Title:          req.Title,
		Description:    req.Description,
		PhaseID:        req.PhaseID,
		Priority:       req.Priority,
		DueDate:        parseDate(req.DueDate),
		EstimatedHours: req.EstimatedHours,
	}
	if task.Priority == 0 {
		task.Priority = 2
	}
	if err := h.svc.CreateTask(uid, uint(pid), &task); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, task)
}

func (h *ProjectHandler) UpdateTask(c *gin.Context) {
	uid := middleware.GetUserID(c)
	taskID, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	delete(updates, "id")
	delete(updates, "project_id")
	if err := h.svc.UpdateTask(uid, uint(taskID), updates); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *ProjectHandler) DeleteTask(c *gin.Context) {
	uid := middleware.GetUserID(c)
	taskID, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if err := h.svc.DeleteTask(uid, uint(taskID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已删除")
}

func (h *ProjectHandler) MoveTask(c *gin.Context) {
	uid := middleware.GetUserID(c)
	taskID, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var req struct {
		Status  string `json:"status" binding:"required"`
		PhaseID *uint  `json:"phase_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.MoveTask(uid, uint(taskID), req.Status, req.PhaseID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已移动")
}

func (h *ProjectHandler) ReorderTasks(c *gin.Context) {
	uid := middleware.GetUserID(c)
	pid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		TaskIDs []uint `json:"task_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ReorderTasks(uid, uint(pid), req.TaskIDs); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "排序成功")
}

func (h *ProjectHandler) GetOverdueTasks(c *gin.Context) {
	uid := middleware.GetUserID(c)
	tasks, err := h.svc.GetOverdueTasks(uid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, tasks)
}

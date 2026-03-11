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

type HabitHandler struct {
	svc     *service.HabitService
	dashSvc *service.DashboardService
}

func NewHabitHandler(svc *service.HabitService, dashSvc *service.DashboardService) *HabitHandler {
	return &HabitHandler{svc: svc, dashSvc: dashSvc}
}

func (h *HabitHandler) Create(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		Name               string `json:"name" binding:"required"`
		Icon               string `json:"icon"`
		Color              string `json:"color"`
		Frequency          string `json:"frequency"`
		WeeklyTarget       int    `json:"weekly_target"`
		StartDate          string `json:"start_date"`
		FamilyMemberTarget *uint  `json:"family_member_target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	startDate := time.Now()
	if req.StartDate != "" {
		for _, layout := range []string{"2006-01-02", time.RFC3339} {
			if t, err := time.Parse(layout, req.StartDate); err == nil {
				startDate = t
				break
			}
		}
	}
	habit := model.Habit{
		Name:               req.Name,
		Icon:               req.Icon,
		Color:              req.Color,
		Frequency:          req.Frequency,
		WeeklyTarget:       req.WeeklyTarget,
		StartDate:          startDate,
		FamilyMemberTarget: req.FamilyMemberTarget,
	}
	if habit.Icon == "" {
		habit.Icon = "📌"
	}
	if habit.Color == "" {
		habit.Color = "#4F46E5"
	}
	if habit.Frequency == "" {
		habit.Frequency = "daily"
	}
	if habit.WeeklyTarget == 0 {
		habit.WeeklyTarget = 1
	}
	if err := h.svc.Create(uid, &habit); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, habit)
}

func (h *HabitHandler) List(c *gin.Context) {
	uid := middleware.GetUserID(c)
	habits, err := h.svc.List(uid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, habits)
}

func (h *HabitHandler) Update(c *gin.Context) {
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

func (h *HabitHandler) Archive(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Archive(uid, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已归档")
}

func (h *HabitHandler) CheckIn(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Note string `json:"note"`
	}
	c.ShouldBindJSON(&req)

	checkIn, err := h.svc.CheckIn(uid, uint(id), time.Now(), req.Note)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	h.dashSvc.LogActivity(uid, nil, "habit_checked", "habit", uint(id), "完成了打卡")
	response.OK(c, checkIn)
}

func (h *HabitHandler) UndoCheckIn(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.UndoCheckIn(uid, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已取消打卡")
}

func (h *HabitHandler) GetCalendar(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))

	data, err := h.svc.GetCalendar(uint(id), year, month)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *HabitHandler) TodayHabits(c *gin.Context) {
	uid := middleware.GetUserID(c)
	data, err := h.svc.GetTodayHabits(uid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, data)
}

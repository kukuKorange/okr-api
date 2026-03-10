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
	var habit model.Habit
	if err := c.ShouldBindJSON(&habit); err != nil {
		response.BadRequest(c, err.Error())
		return
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

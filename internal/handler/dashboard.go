package handler

import (
	"strconv"

	"goaltrack/internal/middleware"
	"goaltrack/internal/service"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	uid := middleware.GetUserID(c)
	data, err := h.svc.GetDashboard(uid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, data)
}

func (h *DashboardHandler) GetActivities(c *gin.Context) {
	uid := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	activities, total, err := h.svc.GetActivities(uid, page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, response.PageData{List: activities, Total: total, Page: page, Size: size})
}

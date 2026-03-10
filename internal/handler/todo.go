package handler

import (
	"strconv"

	"goaltrack/internal/middleware"
	"goaltrack/internal/model"
	"goaltrack/internal/service"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	svc *service.TodoService
}

func NewTodoHandler(svc *service.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

func (h *TodoHandler) Create(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var todo model.TodoItem
	if err := c.ShouldBindJSON(&todo); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Create(uid, &todo); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, todo)
}

func (h *TodoHandler) List(c *gin.Context) {
	uid := middleware.GetUserID(c)
	filter := c.DefaultQuery("filter", "all")
	todos, err := h.svc.List(uid, filter)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, todos)
}

func (h *TodoHandler) Update(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	delete(updates, "id")
	delete(updates, "user_id")

	if isCompleted, ok := updates["is_completed"]; ok {
		if completed, ok := isCompleted.(bool); ok && completed {
			if err := h.svc.Complete(uid, uint(id)); err != nil {
				response.Fail(c, 400, err.Error())
				return
			}
			response.OKMsg(c, "已完成")
			return
		}
	}

	if err := h.svc.Update(uid, uint(id), updates); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *TodoHandler) Delete(c *gin.Context) {
	uid := middleware.GetUserID(c)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uid, uint(id)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已删除")
}

func (h *TodoHandler) Reorder(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Reorder(uid, req.IDs); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "排序成功")
}

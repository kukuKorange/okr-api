package handler

import (
	"strconv"

	"goaltrack/internal/middleware"
	"goaltrack/internal/service"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type FamilyHandler struct {
	svc *service.FamilyService
}

func NewFamilyHandler(svc *service.FamilyService) *FamilyHandler {
	return &FamilyHandler{svc: svc}
}

func (h *FamilyHandler) Create(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		Name string `json:"name" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	family, err := h.svc.Create(uid, req.Name)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, family)
}

func (h *FamilyHandler) Get(c *gin.Context) {
	uid := middleware.GetUserID(c)
	family, err := h.svc.GetByUserID(uid)
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.OK(c, family)
}

func (h *FamilyHandler) Join(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.JoinByCode(uid, req.Code); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "加入成功")
}

func (h *FamilyHandler) GetMembers(c *gin.Context) {
	uid := middleware.GetUserID(c)
	family, err := h.svc.GetByUserID(uid)
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	members, err := h.svc.GetMembers(family.ID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, members)
}

func (h *FamilyHandler) RemoveMember(c *gin.Context) {
	uid := middleware.GetUserID(c)
	memberID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.RemoveMember(uid, uint(memberID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OKMsg(c, "已移除")
}

func (h *FamilyHandler) RegenerateInviteCode(c *gin.Context) {
	uid := middleware.GetUserID(c)
	code, err := h.svc.RegenerateInviteCode(uid)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"invite_code": code})
}

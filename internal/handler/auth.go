package handler

import (
	"goaltrack/internal/middleware"
	"goaltrack/internal/service"
	jwtPkg "goaltrack/pkg/jwt"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.AuthService
	jwt *jwtPkg.Manager
}

func NewAuthHandler(svc *service.AuthService, jwt *jwtPkg.Manager) *AuthHandler {
	return &AuthHandler{svc: svc, jwt: jwt}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required,min=2,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.svc.Register(req.Nickname, req.Email, req.Phone, req.Password)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	accessToken, _ := h.jwt.GenerateAccessToken(user.ID)
	refreshToken, _ := h.jwt.GenerateRefreshToken(user.ID)

	response.OK(c, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Account  string `json:"account" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.svc.Login(req.Account, req.Password)
	if err != nil {
		response.Fail(c, 401, err.Error())
		return
	}

	accessToken, _ := h.jwt.GenerateAccessToken(user.ID)
	refreshToken, _ := h.jwt.GenerateRefreshToken(user.ID)

	response.OK(c, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	claims, err := h.jwt.ParseToken(req.RefreshToken)
	if err != nil || claims.Type != "refresh" {
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	accessToken, _ := h.jwt.GenerateAccessToken(claims.UserID)
	refreshToken, _ := h.jwt.GenerateRefreshToken(claims.UserID)

	response.OK(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	uid := middleware.GetUserID(c)
	user, err := h.svc.GetProfile(uid)
	if err != nil {
		response.ServerError(c, "用户不存在")
		return
	}
	response.OK(c, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	uid := middleware.GetUserID(c)
	var req struct {
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	if err := h.svc.UpdateProfile(uid, updates); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

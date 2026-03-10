package handler

import (
	"goaltrack/internal/service"
	"goaltrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	folder := c.DefaultPostForm("folder", "uploads")

	maxSize := int64(10 << 20) // 10MB default
	if folder == "avatars" {
		maxSize = 5 << 20 // 5MB for avatars
	}
	if file.Size > maxSize {
		response.BadRequest(c, "文件太大")
		return
	}

	url, err := h.svc.Upload(file, folder)
	if err != nil {
		response.ServerError(c, "上传失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"url": url})
}
